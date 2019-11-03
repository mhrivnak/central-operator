package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	hrivnakv1alpha1 "github.com/mhrivnak/central-operator/pkg/apis/hrivnak/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/predicate"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_api")

// Add creates a new Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, watch *hrivnakv1alpha1.Watch) (controller.Controller, *Reconciler, error) {
	scheme := mgr.GetScheme()
	_, err := scheme.New(watch.Spec.GVK())
	if runtime.IsNotRegisteredError(err) {
		// Register the GVK with the schema
		scheme.AddKnownTypeWithName(watch.Spec.GVK(), &unstructured.Unstructured{})
		metav1.AddToGroupVersion(mgr.GetScheme(), schema.GroupVersion{
			Group:   watch.Spec.GVK().Group,
			Version: watch.Spec.GVK().Version,
		})
	} else if err != nil {
		log.Error(err, "")
		return nil, nil, err
	}

	reconciler := &Reconciler{
		client:           mgr.GetClient(),
		scheme:           mgr.GetScheme(),
		serviceName:      watch.Spec.ServiceName,
		groupVersionKind: watch.Spec.GVK(),
		finalizer:        watch.Spec.Finalizer,
	}

	// Add secondary watches to the scheme
	for _, ownedWatch := range watch.Spec.OwnedWatches {
		owned := &unstructured.Unstructured{}
		owned.SetGroupVersionKind(ownedWatch)
		reconciler.scheme.AddKnownTypes(ownedWatch.GroupVersion(), owned)
	}

	// Create a new controller
	opts := controller.Options{
		Reconciler:              reconciler,
		MaxConcurrentReconciles: watch.Spec.MaxConcurrentReconciles,
	}
	cname := fmt.Sprintf("%v-controller", strings.ToLower(watch.Spec.GVK().String()))
	c, err := controller.New(cname, mgr, opts)
	if err != nil {
		return nil, nil, err
	}

	// Watch the primary resource
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(watch.Spec.GVK())
	if err = c.Watch(&source.Kind{Type: u}, &handler.EnqueueRequestForObject{}, predicate.GenerationChangedPredicate{}); err != nil {
		return nil, nil, err
	}

	// Watch any secondary resources
	for _, ownedWatch := range watch.Spec.OwnedWatches {
		owned := &unstructured.Unstructured{}
		owned.SetGroupVersionKind(ownedWatch)

		if err = c.Watch(&source.Kind{Type: owned}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    u,
		}); err != nil {
			return nil, nil, err
		}
	}

	return c, reconciler, nil
}

// blank assignment to verify that Reconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &Reconciler{}

// Reconciler reconciles a generic resource by calling an API
type Reconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client           client.Client
	scheme           *runtime.Scheme
	serviceName      string
	groupVersionKind schema.GroupVersionKind
	stopped          bool
	mtx              sync.Mutex
	finalizer        string
}

// Stop ensures that the reconciler will not perform any additional actions.
func (r *Reconciler) Stop() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.stopped = true
}

// IsStopped returns true iff the reconciler is stopped
func (r *Reconciler) IsStopped() bool {
	return r.stopped
}

// Reconcile by calling an API
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	if r.stopped {
		reqLogger.Info("reconciler is stopped; skipping reconciliation")
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Reconciling with API")

	// Fetch the instance to make sure it exists. No sense calling the API if it doesn't.
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(r.groupVersionKind)
	err := r.client.Get(context.TODO(), request.NamespacedName, u)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if r.finalizer != "" {
		if u.GetDeletionTimestamp().IsZero() {
			// Add finalizer if it's missing
			if !containsString(u.GetFinalizers(), r.finalizer) {
				u.SetFinalizers(append(u.GetFinalizers(), r.finalizer))
				return reconcile.Result{}, r.client.Update(context.TODO(), u)
			}
		} else {
			if containsString(u.GetFinalizers(), r.finalizer) {
				// Call the API to let it do finalizer work
				result, err := r.callAPI(request)
				if err != nil {
					return result, err
				}
				// remove the finalizer only if the API responded without requeueing
				if result.Requeue == false && result.RequeueAfter == time.Duration(0) {
					u.SetFinalizers(removeString(u.GetFinalizers(), r.finalizer))
					return reconcile.Result{}, r.client.Update(context.TODO(), u)
				}
				return result, nil
			}
			// finalizer not present, so no-op
			return reconcile.Result{}, nil
		}
	}
	return r.callAPI(request)
}

func (r *Reconciler) callAPI(request reconcile.Request) (reconcile.Result, error) {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(request)
	url := fmt.Sprintf("http://%s", r.serviceName)
	response, err := http.Post(url, "application/json; charset=utf-8", b)
	if err != nil {
		return reconcile.Result{}, err
	}
	result := reconcile.Result{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func containsString(x []string, y string) bool {
	for _, value := range x {
		if value == y {
			return true
		}
	}
	return false
}

func removeString(x []string, y string) []string {
	ret := []string{}
	for _, value := range x {
		if value != y {
			ret = append(ret, value)
		}
	}
	return ret
}
