package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WatchSpec defines the desired state of Watch. The provided Group, Version,
// and Kind will be the primary resource type for a new controller. The
// controller will watch that GVK with the operator-sdk's
// GenerationChangedPredicate, so changes only to the primary resource's Status
// will not cause reconciliation to happen.
// +k8s:openapi-gen=true
type WatchSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// ServiceName is the Name property of a Service that will be sent an HTTP
	// request to port 80 to implement reconciliation.
	ServiceName string `json:"serviceName"`

	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`

	// Finalizer, if set, will cause a finalizer with the given name to be
	// automatically managed on each primary resource. This operator will ensure
	// the finalizer is present until the resource is deleted. Upon deletion, it
	// will call the API as many times as it takes until the API responds with
	// Requeue == false and RequeueAfter set to zero. After that, the finalizer
	// will be automatically removed.
	// WARNING: do not modify this value
	Finalizer string `json:"finalizer,omitempty"`

	// OwnedWatches is a list of GVKs for resource types that will have an owner
	// reference to the primary resource and that should be watched by this
	// operator with the "EnqueueRequestForOwner" handler.
	OwnedWatches []schema.GroupVersionKind `json:"ownedWatches,omitempty"`

	// MaxConcurrentReconciles is the maximum number of concurrent Reconciles
	// that should run. Defaults to 1.
	MaxConcurrentReconciles int `json:"maxConcurrentReconciles,omitempty"`
}

// GVK returns the GVK for the primary resource
func (w *WatchSpec) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   w.Group,
		Version: w.Version,
		Kind:    w.Kind,
	}
}

// WatchState represents the state of a Watch as a single word.
type WatchState string

const (
	// WatchActive indicates that the Watch has a controller that is active
	WatchActive WatchState = "WatchActive"

	// WatchStopped indicates that the Watch has a controller that was stopped.
	// This state happens when a Watch is deleted and then a new one is created
	// with the same GVK. The operator will need to be re-started to clear the
	// old controller.
	WatchStopped WatchState = "WatchStopped"
)

// WatchStatus defines the observed state of Watch
// +k8s:openapi-gen=true
type WatchStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// State represents the state of a Watch as a single word.
	State WatchState `json:"state,omitempty"`

	// Reason is a human-readable explanation for the State.
	Reason string `json:"reason,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Watch is the Schema for the watches API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=watches,scope=Namespaced
type Watch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WatchSpec   `json:"spec,omitempty"`
	Status WatchStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WatchList contains a list of Watch
type WatchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Watch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Watch{}, &WatchList{})
}
