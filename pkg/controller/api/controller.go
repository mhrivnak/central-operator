package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
func Add(mgr manager.Manager, watch *hrivnakv1alpha1.Watch) (*APIReconciler, error) {
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
		return nil, err
	}

	reconciler := &APIReconciler{
		client:           mgr.GetClient(),
		scheme:           mgr.GetScheme(),
		serviceName:      watch.Spec.ServiceName,
		groupVersionKind: watch.Spec.GVK(),
	}

	// Create a new controller
	c, err := controller.New(fmt.Sprintf("%v-controller", strings.ToLower(watch.Spec.GVK().String())), mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return nil, err
	}

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(watch.Spec.GVK())
	if err = c.Watch(&source.Kind{Type: u}, &handler.EnqueueRequestForObject{}, predicate.GenerationChangedPredicate{}); err != nil {
		return nil, err
	}

	return reconciler, nil
}

// blank assignment to verify that APIReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &APIReconciler{}

// APIReconciler reconciles a generic resource by calling an API
type APIReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client           client.Client
	scheme           *runtime.Scheme
	serviceName      string
	groupVersionKind schema.GroupVersionKind
}

// Reconcile by calling an API
func (r *APIReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
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

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(request)
	url := fmt.Sprintf("http://%s", r.serviceName)
	response, err := http.Post(url, "application/json; charset=utf-8", b)
	if err != nil {
		reqLogger.Error(err, "Failed to access API")
		return reconcile.Result{}, err
	}
	result := reconcile.Result{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		reqLogger.Error(err, "Could not deserialize response")
		return result, err
	}

	return result, nil
}
