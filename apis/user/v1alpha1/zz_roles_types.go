/*
Copyright 2022 Upbound Inc.
*/

// Code generated by upjet. DO NOT EDIT.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type RolesInitParameters struct {

	// Indicates if the list of roles is exhaustive. In this case, roles that are manually added to the user will be removed. Defaults to true.
	Exhaustive *bool `json:"exhaustive,omitempty" tf:"exhaustive,omitempty"`

	// The realm this user exists in.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/realm/v1alpha1.Realm
	RealmID *string `json:"realmId,omitempty" tf:"realm_id,omitempty"`

	// Reference to a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDRef *v1.Reference `json:"realmIdRef,omitempty" tf:"-"`

	// Selector for a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDSelector *v1.Selector `json:"realmIdSelector,omitempty" tf:"-"`

	// A list of role IDs to map to the user
	// +listType=set
	RoleIds []*string `json:"roleIds,omitempty" tf:"role_ids,omitempty"`

	// The ID of the user this resource should manage roles for.
	// +crossplane:generate:reference:type=User
	UserID *string `json:"userId,omitempty" tf:"user_id,omitempty"`

	// Reference to a User to populate userId.
	// +kubebuilder:validation:Optional
	UserIDRef *v1.Reference `json:"userIdRef,omitempty" tf:"-"`

	// Selector for a User to populate userId.
	// +kubebuilder:validation:Optional
	UserIDSelector *v1.Selector `json:"userIdSelector,omitempty" tf:"-"`
}

type RolesObservation struct {

	// Indicates if the list of roles is exhaustive. In this case, roles that are manually added to the user will be removed. Defaults to true.
	Exhaustive *bool `json:"exhaustive,omitempty" tf:"exhaustive,omitempty"`

	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	// The realm this user exists in.
	RealmID *string `json:"realmId,omitempty" tf:"realm_id,omitempty"`

	// A list of role IDs to map to the user
	// +listType=set
	RoleIds []*string `json:"roleIds,omitempty" tf:"role_ids,omitempty"`

	// The ID of the user this resource should manage roles for.
	UserID *string `json:"userId,omitempty" tf:"user_id,omitempty"`
}

type RolesParameters struct {

	// Indicates if the list of roles is exhaustive. In this case, roles that are manually added to the user will be removed. Defaults to true.
	// +kubebuilder:validation:Optional
	Exhaustive *bool `json:"exhaustive,omitempty" tf:"exhaustive,omitempty"`

	// The realm this user exists in.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/realm/v1alpha1.Realm
	// +kubebuilder:validation:Optional
	RealmID *string `json:"realmId,omitempty" tf:"realm_id,omitempty"`

	// Reference to a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDRef *v1.Reference `json:"realmIdRef,omitempty" tf:"-"`

	// Selector for a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDSelector *v1.Selector `json:"realmIdSelector,omitempty" tf:"-"`

	// A list of role IDs to map to the user
	// +kubebuilder:validation:Optional
	// +listType=set
	RoleIds []*string `json:"roleIds,omitempty" tf:"role_ids,omitempty"`

	// The ID of the user this resource should manage roles for.
	// +crossplane:generate:reference:type=User
	// +kubebuilder:validation:Optional
	UserID *string `json:"userId,omitempty" tf:"user_id,omitempty"`

	// Reference to a User to populate userId.
	// +kubebuilder:validation:Optional
	UserIDRef *v1.Reference `json:"userIdRef,omitempty" tf:"-"`

	// Selector for a User to populate userId.
	// +kubebuilder:validation:Optional
	UserIDSelector *v1.Selector `json:"userIdSelector,omitempty" tf:"-"`
}

// RolesSpec defines the desired state of Roles
type RolesSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     RolesParameters `json:"forProvider"`
	// THIS IS A BETA FIELD. It will be honored
	// unless the Management Policies feature flag is disabled.
	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// The same fields are also added to the terraform ignore_changes hook, to
	// avoid updating them after creation. This is useful for fields that are
	// required on creation, but we do not desire to update them after creation,
	// for example because of an external controller is managing them, like an
	// autoscaler.
	InitProvider RolesInitParameters `json:"initProvider,omitempty"`
}

// RolesStatus defines the observed state of Roles.
type RolesStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        RolesObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// Roles is the Schema for the Roless API.
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,keycloak}
type Roles struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.roleIds) || (has(self.initProvider) && has(self.initProvider.roleIds))",message="spec.forProvider.roleIds is a required parameter"
	Spec   RolesSpec   `json:"spec"`
	Status RolesStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RolesList contains a list of Roless
type RolesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Roles `json:"items"`
}

// Repository type metadata.
var (
	Roles_Kind             = "Roles"
	Roles_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: Roles_Kind}.String()
	Roles_KindAPIVersion   = Roles_Kind + "." + CRDGroupVersion.String()
	Roles_GroupVersionKind = CRDGroupVersion.WithKind(Roles_Kind)
)

func init() {
	SchemeBuilder.Register(&Roles{}, &RolesList{})
}