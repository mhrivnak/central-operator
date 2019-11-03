// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/mhrivnak/central-operator/pkg/apis/hrivnak/v1alpha1.Watch":       schema_pkg_apis_hrivnak_v1alpha1_Watch(ref),
		"github.com/mhrivnak/central-operator/pkg/apis/hrivnak/v1alpha1.WatchSpec":   schema_pkg_apis_hrivnak_v1alpha1_WatchSpec(ref),
		"github.com/mhrivnak/central-operator/pkg/apis/hrivnak/v1alpha1.WatchStatus": schema_pkg_apis_hrivnak_v1alpha1_WatchStatus(ref),
	}
}

func schema_pkg_apis_hrivnak_v1alpha1_Watch(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Watch is the Schema for the watches API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/mhrivnak/central-operator/pkg/apis/hrivnak/v1alpha1.WatchSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/mhrivnak/central-operator/pkg/apis/hrivnak/v1alpha1.WatchStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/mhrivnak/central-operator/pkg/apis/hrivnak/v1alpha1.WatchSpec", "github.com/mhrivnak/central-operator/pkg/apis/hrivnak/v1alpha1.WatchStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_hrivnak_v1alpha1_WatchSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "WatchSpec defines the desired state of Watch. The provided Group, Version, and Kind will be the primary resource type for a new controller. The controller will watch that GVK with the operator-sdk's GenerationChangedPredicate, so changes only to the primary resource's Status will not cause reconciliation to happen.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"serviceName": {
						SchemaProps: spec.SchemaProps{
							Description: "ServiceName is the Name property of a Service that will be sent an HTTP request to port 80 to implement reconciliation.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"group": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"version": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"kind": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"finalizer": {
						SchemaProps: spec.SchemaProps{
							Description: "Finalizer, if set, will cause a finalizer with the given name to be automatically managed on each primary resource. This operator will ensure the finalizer is present until the resource is deleted. Upon deletion, it will call the API as many times as it takes until the API responds with Requeue == false and RequeueAfter set to zero. After that, the finalizer will be automatically removed. WARNING: do not modify this value",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"ownedWatches": {
						SchemaProps: spec.SchemaProps{
							Description: "OwnedWatches is a list of GVKs for resource types that will have an owner reference to the primary resource and that should be watched by this operator with the \"EnqueueRequestForOwner\" handler.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/apimachinery/pkg/runtime/schema.GroupVersionKind"),
									},
								},
							},
						},
					},
					"maxConcurrentReconciles": {
						SchemaProps: spec.SchemaProps{
							Description: "MaxConcurrentReconciles is the maximum number of concurrent Reconciles that should run. Defaults to 1.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
				},
				Required: []string{"serviceName", "group", "version", "kind", "maxConcurrentReconciles"},
			},
		},
		Dependencies: []string{
			"k8s.io/apimachinery/pkg/runtime/schema.GroupVersionKind"},
	}
}

func schema_pkg_apis_hrivnak_v1alpha1_WatchStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "WatchStatus defines the observed state of Watch",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"state": {
						SchemaProps: spec.SchemaProps{
							Description: "State represents the state of a Watch as a single word.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"reason": {
						SchemaProps: spec.SchemaProps{
							Description: "Reason is a human-readable explanation for the State.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
			},
		},
	}
}
