/*
Copyright 2022 Upbound Inc.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// A ProviderConfigSpec defines the desired state of a ProviderConfig.
// A ProviderConfigSpec defines the desired state of a ProviderConfig.
type ProviderConfigSpec struct {
	// Credentials required to authenticate to this provider.
	Credentials ProviderCredentials `json:"credentials"`

	// ConnectionSecretKeys customizes the connection secret keys published
	// for managed resources that use this ProviderConfig, e.g. renaming
	// "clientID"/"clientSecret" to whatever keys a consumer expects (such as
	// "client-id"/"client-secret" for Envoy Gateway's OIDC SecurityPolicy).
	// This is applied out-of-band from the managed resource reconciliation
	// itself, by a separate controller: by default it adds the configured
	// keys to the resource's own connection secret (the one named by
	// spec.writeConnectionSecretToRef), and it can optionally publish a
	// separate, transformed copy instead. See
	// internal/controller/cluster/connectionsecrettransform for details.
	// +optional
	ConnectionSecretKeys *ConnectionSecretKeys `json:"connectionSecretKeys,omitempty"`
}

// ProviderCredentials required to authenticate.
type ProviderCredentials struct {
	// Source of the provider credentials.
	// +kubebuilder:validation:Enum=None;Secret;Environment;Filesystem
	Source xpv1.CredentialsSource `json:"source"`

	xpv1.CommonCredentialSelectors `json:",inline"`
}

// ConnectionSecretTransformMode selects where the configured keys are
// published.
// +kubebuilder:validation:Enum=InPlace;SeparateSecret
type ConnectionSecretTransformMode string

const (
	// TransformModeInPlace adds the configured keys to the managed
	// resource's own connection secret, i.e. the one named by its
	// spec.writeConnectionSecretToRef. No second secret is created. Since
	// the provider's own connection secret publisher re-adds every key it
	// owns on the next reconcile, renaming here is additive: the new key is
	// added next to the original one rather than replacing it. This is the
	// default.
	TransformModeInPlace ConnectionSecretTransformMode = "InPlace"

	// TransformModeSeparateSecret publishes a second, transformed secret
	// next to the connection secret and leaves the connection secret
	// untouched. Unlike InPlace this can drop the original key names, at
	// the cost of an additional secret that consumers have to point at.
	TransformModeSeparateSecret ConnectionSecretTransformMode = "SeparateSecret"
)

// ConnectionSecretKeys configures how connection secret keys are renamed for
// managed resources using this ProviderConfig.
type ConnectionSecretKeys struct {
	// Mode selects whether the renamed/added keys are written into the
	// managed resource's own connection secret ("InPlace", the default) or
	// into a separate, transformed secret published next to it
	// ("SeparateSecret").
	// +optional
	Mode ConnectionSecretTransformMode `json:"mode,omitempty"`

	// Rename maps an existing connection secret key (as published by the
	// provider, e.g. "clientID") to the key name that should be used
	// instead (e.g. "client-id"). Keys that are not listed here are left
	// alone. In the default "InPlace" mode the original key stays as well,
	// i.e. the new name is an additional alias for the same value; in
	// "SeparateSecret" mode only the new name is published.
	// +optional
	Rename map[string]string `json:"rename,omitempty"`

	// Add maps a new connection secret key to a value looked up elsewhere,
	// so that it is added to the connection secret alongside the (renamed)
	// values published by the provider. Each value is a source
	// expression of the form "status:<dot.path>" (a field under the owning
	// managed resource's status, e.g. "status:atProvider.clientId") or
	// "providerConfig:<dot.path>" (a field of this ProviderConfig, e.g.
	// "providerConfig:metadata.name"). Only scalar (string, number,
	// boolean) fields are supported; the source path is never resolved
	// against the credentials Secret itself, so secret material referenced
	// only by name/namespace/key cannot leak this way.
	// +optional
	Add map[string]string `json:"add,omitempty"`
}

// A ProviderConfigStatus reflects the observed state of a ProviderConfig.
type ProviderConfigStatus struct {
	xpv1.ProviderConfigStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A ProviderConfig configures a keycloak provider.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentials.secretRef.name",priority=1
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,keycloak}
type ProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderConfigSpec   `json:"spec"`
	Status ProviderConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderConfigList contains a list of ProviderConfig.
type ProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfig `json:"items"`
}

// +kubebuilder:object:root=true

// A ProviderConfigUsage indicates that a resource is using a ProviderConfig.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="CONFIG-NAME",type="string",JSONPath=".providerConfigRef.name"
// +kubebuilder:printcolumn:name="RESOURCE-KIND",type="string",JSONPath=".resourceRef.kind"
// +kubebuilder:printcolumn:name="RESOURCE-NAME",type="string",JSONPath=".resourceRef.name"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,keycloak}
type ProviderConfigUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	xpv1.ProviderConfigUsage `json:",inline"`
}

// +kubebuilder:object:root=true

// ProviderConfigUsageList contains a list of ProviderConfigUsage
type ProviderConfigUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfigUsage `json:"items"`
}

// LocalSecretReference identifies a Secret in the same namespace as the
// referencing object.
type LocalSecretReference struct {
	// Name of the Secret.
	Name string `json:"name"`
}

// ConnectionSecretTransformSpec defines the desired renaming/field-addition
// applied to a connection secret.
type ConnectionSecretTransformSpec struct {
	// SourceSecretRef identifies the connection secret to transform. It must
	// be a Crossplane connection secret (type
	// "connection.crossplane.io/v1alpha1") owned by a managed resource, in
	// the same namespace as this ConnectionSecretTransform.
	SourceSecretRef LocalSecretReference `json:"sourceSecretRef"`

	// TransformedSecretName overrides the name of the republished secret in
	// "SeparateSecret" mode. Defaults to the source secret's name plus the
	// "-transformed" suffix. Takes precedence over the source resource's
	// "keycloak.crossplane.io/connection-secret-transform-name" annotation,
	// if both are set. It has no effect in the default "InPlace" mode,
	// where the keys are written into the source secret itself.
	// +optional
	TransformedSecretName string `json:"transformedSecretName,omitempty"`

	// Mode selects whether the renamed/added keys are written into the
	// source connection secret itself ("InPlace", the default) or into a
	// separate, transformed secret published next to it
	// ("SeparateSecret"). Takes precedence over the source resource's
	// "keycloak.crossplane.io/connection-secret-transform-mode" annotation
	// and its ProviderConfig's spec.connectionSecretKeys.mode.
	// +optional
	Mode ConnectionSecretTransformMode `json:"mode,omitempty"`

	// Rename maps an existing connection secret key (as published by the
	// provider, e.g. "clientID") to the key name that should be used
	// instead (e.g. "client-id"). Keys that are not listed here are left
	// alone. In the default "InPlace" mode the original key stays as well,
	// i.e. the new name is an additional alias for the same value; in
	// "SeparateSecret" mode only the new name is published. Merged on
	// top of (and taking precedence over) any renaming configured via the
	// source resource's ProviderConfig or its own rename annotation.
	// +optional
	Rename map[string]string `json:"rename,omitempty"`

	// Add maps a new connection secret key to a value looked up elsewhere,
	// so that it is added to the connection secret (or, in
	// "SeparateSecret" mode, to the transformed copy) alongside the
	// (renamed) values published by the provider. Each value is a source
	// expression of the form "status:<dot.path>" (a field under the
	// managed resource owning the source secret, e.g.
	// "status:atProvider.clientId") or "providerConfig:<dot.path>" (a
	// field of the ProviderConfig that resource references, e.g.
	// "providerConfig:metadata.name"; only available when that resource is
	// cluster-scoped). Only scalar (string, number, boolean) fields are
	// supported; the source path is never resolved against the
	// credentials Secret itself, so secret material referenced only by
	// name/namespace/key cannot leak this way. Merged on top of (and
	// taking precedence over) any fields added via the source resource's
	// ProviderConfig or its own add-fields annotation.
	// +optional
	Add map[string]string `json:"add,omitempty"`
}

// ConnectionSecretTransformStatus reflects the observed state of a
// ConnectionSecretTransform.
type ConnectionSecretTransformStatus struct {
	xpv1.ConditionedStatus `json:",inline"`

	// TransformedSecretName is the name of the secret the configured keys
	// were last successfully published to: the source connection secret
	// itself in "InPlace" mode, or the separate transformed secret in
	// "SeparateSecret" mode.
	// +optional
	TransformedSecretName string `json:"transformedSecretName,omitempty"`
}

// +kubebuilder:object:root=true

// A ConnectionSecretTransform republishes a managed resource's connection
// secret with keys renamed and/or extended, without requiring the managed
// resource itself to be edited. It is an alternative to configuring the
// same behaviour via the ProviderConfig's spec.connectionSecretKeys or the
// source resource's own annotations - useful when the resource is owned by
// a Composition or otherwise not convenient to annotate directly. All three
// configuration sources may be combined; see
// ConnectionSecretTransformSpec.Rename and .Add for the precedence rules.
// The reconciling logic lives in
// internal/controller/cluster/connectionsecrettransform, alongside the
// annotation/ProviderConfig-driven behaviour.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SOURCE-SECRET",type="string",JSONPath=".spec.sourceSecretRef.name"
// +kubebuilder:printcolumn:name="TRANSFORMED-SECRET",type="string",JSONPath=".status.transformedSecretName"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories={crossplane,provider,keycloak}
type ConnectionSecretTransform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionSecretTransformSpec   `json:"spec"`
	Status ConnectionSecretTransformStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConnectionSecretTransformList contains a list of ConnectionSecretTransform.
type ConnectionSecretTransformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConnectionSecretTransform `json:"items"`
}
