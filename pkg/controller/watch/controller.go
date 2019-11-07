package watch

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	hrivnakv1alpha1 "github.com/mhrivnak/central-operator/pkg/apis/hrivnak/v1alpha1"
	apicontroller "github.com/mhrivnak/central-operator/pkg/controller/api"
	"github.com/operator-framework/operator-sdk/pkg/ansible/proxy/controllermap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_watch")

// FinalizerName is the string used in the Watch finalizer
const FinalizerName string = "finalizer.central-operator.hrivnak.org"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Watch Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, cMap *controllermap.ControllerMap) error {
	return add(mgr, newReconciler(mgr, cMap))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, cMap *controllermap.ControllerMap) reconcile.Reconciler {
	return &ReconcileWatch{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		mgr:    mgr,
		cMap:   cMap,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("watch-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Watch
	err = c.Watch(&source.Kind{Type: &hrivnakv1alpha1.Watch{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileWatch implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWatch{}

// ReconcileWatch reconciles a Watch object
type ReconcileWatch struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client      client.Client
	scheme      *runtime.Scheme
	mgr         manager.Manager
	reconcilers map[string]*apicontroller.Reconciler
	cMap        *controllermap.ControllerMap
	mtx         sync.Mutex
}

// Reconcile reads that state of the cluster for a Watch object and makes changes based on the state read
// and what is in the Watch.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWatch) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Watch")

	// Fetch the Watch instance
	instance := &hrivnakv1alpha1.Watch{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add finalizer if it's missing
		if !containsString(instance.ObjectMeta.Finalizers, FinalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, FinalizerName)
			return reconcile.Result{}, r.client.Update(context.TODO(), instance)
		}
	} else {
		// Cleanup then remove finalizer
		if containsString(instance.ObjectMeta.Finalizers, FinalizerName) {
			r.ensureAPIControllerStopped(instance)

			// remove finalizer
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, FinalizerName)
			return reconcile.Result{}, r.client.Update(context.TODO(), instance)
		}
		// finalizer not present, so no-op
		return reconcile.Result{}, nil
	}

	if r.cMap == nil {
		return reconcile.Result{}, fmt.Errorf("missing controller map")
	}

	gvk := instance.Spec.GVK()
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.reconcilers == nil {
		r.reconcilers = make(map[string]*apicontroller.Reconciler)
	}
	reconciler, ok := r.reconcilers[gvk.String()]
	if !ok {
		controller, reconciler, err := apicontroller.Add(r.mgr, instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.reconcilers[gvk.String()] = reconciler
		r.cMap.Store(gvk, &controllermap.Contents{
			Controller: controller,
		})
	}

	// update the status if necessary
	newStatus := hrivnakv1alpha1.WatchStatus{}
	if reconciler.IsStopped() {
		newStatus.State = hrivnakv1alpha1.WatchStopped
		newStatus.Reason = "The operator must be re-started to clear a previously-existing controller for the same GVK"
	} else {
		newStatus.State = hrivnakv1alpha1.WatchActive
		newStatus.Reason = ""
	}
	if !reflect.DeepEqual(newStatus, instance.Status) {
		newInstance := instance.DeepCopy()
		newInstance.Status = newStatus
		return reconcile.Result{}, r.client.Status().Update(context.TODO(), newInstance)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileWatch) ensureAPIControllerStopped(instance *hrivnakv1alpha1.Watch) {
	gvk := instance.Spec.GVK()
	reconciler, ok := r.reconcilers[gvk.String()]
	if ok {
		reconciler.Stop()
	}
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
