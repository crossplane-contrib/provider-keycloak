/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type AuthenticationFlowBindingOverridesInitParameters struct {

	// Browser flow id, (flow needs to exist)
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/authenticationflow/v1alpha1.Flow
	BrowserID *string `json:"browserId,omitempty"`

	// Reference to a Flow in authenticationflow to populate browserId.
	// +kubebuilder:validation:Optional
	BrowserIDRef *v1.Reference `json:"browserIdRef,omitempty"`

	// Selector for a Flow in authenticationflow to populate browserId.
	// +kubebuilder:validation:Optional
	BrowserIDSelector *v1.Selector `json:"browserIdSelector,omitempty"`

	// Direct grant flow id (flow needs to exist)
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/authenticationflow/v1alpha1.Flow
	DirectGrantID *string `json:"directGrantId,omitempty"`

	// Reference to a Flow in authenticationflow to populate directGrantId.
	// +kubebuilder:validation:Optional
	DirectGrantIDRef *v1.Reference `json:"directGrantIdRef,omitempty"`

	// Selector for a Flow in authenticationflow to populate directGrantId.
	// +kubebuilder:validation:Optional
	DirectGrantIDSelector *v1.Selector `json:"directGrantIdSelector,omitempty"`
}

type AuthenticationFlowBindingOverridesObservation struct {

	// Browser flow id, (flow needs to exist)
	BrowserID *string `json:"browserId,omitempty"`

	// Direct grant flow id (flow needs to exist)
	DirectGrantID *string `json:"directGrantId,omitempty"`
}

type AuthenticationFlowBindingOverridesParameters struct {

	// Browser flow id, (flow needs to exist)
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/authenticationflow/v1alpha1.Flow
	// +kubebuilder:validation:Optional
	BrowserID *string `json:"browserId,omitempty"`

	// Reference to a Flow in authenticationflow to populate browserId.
	// +kubebuilder:validation:Optional
	BrowserIDRef *v1.Reference `json:"browserIdRef,omitempty"`

	// Selector for a Flow in authenticationflow to populate browserId.
	// +kubebuilder:validation:Optional
	BrowserIDSelector *v1.Selector `json:"browserIdSelector,omitempty"`

	// Direct grant flow id (flow needs to exist)
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/authenticationflow/v1alpha1.Flow
	// +kubebuilder:validation:Optional
	DirectGrantID *string `json:"directGrantId,omitempty"`

	// Reference to a Flow in authenticationflow to populate directGrantId.
	// +kubebuilder:validation:Optional
	DirectGrantIDRef *v1.Reference `json:"directGrantIdRef,omitempty"`

	// Selector for a Flow in authenticationflow to populate directGrantId.
	// +kubebuilder:validation:Optional
	DirectGrantIDSelector *v1.Selector `json:"directGrantIdSelector,omitempty"`
}

type AuthorizationInitParameters struct {

	// When true, resources can be managed remotely by the resource server. Defaults to false.
	AllowRemoteResourceManagement *bool `json:"allowRemoteResourceManagement,omitempty"`

	// Dictates how the policies associated with a given permission are evaluated and how a final decision is obtained. Could be one of AFFIRMATIVE, CONSENSUS, or UNANIMOUS. Applies to permissions.
	DecisionStrategy *string `json:"decisionStrategy,omitempty"`

	// When true, defaults set by Keycloak will be respected. Defaults to false.
	KeepDefaults *bool `json:"keepDefaults,omitempty"`

	// Dictates how policies are enforced when evaluating authorization requests. Can be one of ENFORCING, PERMISSIVE, or DISABLED.
	PolicyEnforcementMode *string `json:"policyEnforcementMode,omitempty"`
}

type AuthorizationObservation struct {

	// When true, resources can be managed remotely by the resource server. Defaults to false.
	AllowRemoteResourceManagement *bool `json:"allowRemoteResourceManagement,omitempty"`

	// Dictates how the policies associated with a given permission are evaluated and how a final decision is obtained. Could be one of AFFIRMATIVE, CONSENSUS, or UNANIMOUS. Applies to permissions.
	DecisionStrategy *string `json:"decisionStrategy,omitempty"`

	// When true, defaults set by Keycloak will be respected. Defaults to false.
	KeepDefaults *bool `json:"keepDefaults,omitempty"`

	// Dictates how policies are enforced when evaluating authorization requests. Can be one of ENFORCING, PERMISSIVE, or DISABLED.
	PolicyEnforcementMode *string `json:"policyEnforcementMode,omitempty"`
}

type AuthorizationParameters struct {

	// When true, resources can be managed remotely by the resource server. Defaults to false.
	// +kubebuilder:validation:Optional
	AllowRemoteResourceManagement *bool `json:"allowRemoteResourceManagement,omitempty"`

	// Dictates how the policies associated with a given permission are evaluated and how a final decision is obtained. Could be one of AFFIRMATIVE, CONSENSUS, or UNANIMOUS. Applies to permissions.
	// +kubebuilder:validation:Optional
	DecisionStrategy *string `json:"decisionStrategy,omitempty"`

	// When true, defaults set by Keycloak will be respected. Defaults to false.
	// +kubebuilder:validation:Optional
	KeepDefaults *bool `json:"keepDefaults,omitempty"`

	// Dictates how policies are enforced when evaluating authorization requests. Can be one of ENFORCING, PERMISSIVE, or DISABLED.
	// +kubebuilder:validation:Optional
	PolicyEnforcementMode *string `json:"policyEnforcementMode"`
}

type OpenidClientAttributes struct {
	PkceCodeChallengeMethod               *string   `json:"pkce.code.challenge.method"`
	ExcludeSessionStateFromAuthResponse   *bool     `json:"exclude.session.state.from.auth.response"`
	AccessTokenLifespan                   *string   `json:"access.token.lifespan"`
	LoginTheme                            *string   `json:"login_theme"`
	ClientOfflineSessionIdleTimeout       *string   `json:"client.offline.session.idle.timeout,omitempty"`
	DisplayOnConsentScreen                *bool     `json:"display.on.consent.screen"`
	ConsentScreenText                     *string   `json:"consent.screen.text"`
	ClientOfflineSessionMaxLifespan       *string   `json:"client.offline.session.max.lifespan,omitempty"`
	ClientSessionIdleTimeout              *string   `json:"client.session.idle.timeout,omitempty"`
	ClientSessionMaxLifespan              *string   `json:"client.session.max.lifespan,omitempty"`
	UseRefreshTokens                      *bool     `json:"use.refresh.tokens"`
	UseRefreshTokensClientCredentials     *bool     `json:"client_credentials.use_refresh_token"`
	BackchannelLogoutUrl                  *string   `json:"backchannel.logout.url"`
	FrontchannelLogoutUrl                 *string   `json:"frontchannel.logout.url"`
	BackchannelLogoutRevokeOfflineTokens  *bool     `json:"backchannel.logout.revoke.offline.tokens"`
	BackchannelLogoutSessionRequired      *bool     `json:"backchannel.logout.session.required"`
	Oauth2DeviceAuthorizationGrantEnabled *bool     `json:"oauth2.device.authorization.grant.enabled"`
	Oauth2DeviceCodeLifespan              *string   `json:"oauth2.device.code.lifespan,omitempty"`
	Oauth2DevicePollingInterval           *string   `json:"oauth2.device.polling.interval,omitempty"`
	PostLogoutRedirectUris                []*string `json:"post.logout.redirect.uris,omitempty"`
}

// OpenIDConfig contains the fields specific to OpenID clients.
type OpenIDConfig struct {
	Attributes                             *OpenidClientAttributes   `json:"attributes,omitempty"`
	AccessTokenLifespan                    *string                   `json:"accessTokenLifespan,omitempty"`
	AccessType                             *string                   `json:"accessType,omitempty"`
	AdminURL                               *string                   `json:"adminUrl,omitempty"`
	Authorization                          []AuthorizationParameters `json:"authorization,omitempty"`
	BackchannelLogoutRevokeOfflineSessions *bool                     `json:"backchannelLogoutRevokeOfflineSessions,omitempty"`
	BackchannelLogoutSessionRequired       *bool                     `json:"backchannelLogoutSessionRequired,omitempty"`
	BackchannelLogoutURL                   *string                   `json:"backchannelLogoutUrl,omitempty"`
	ClientAuthenticatorType                *string                   `json:"clientAuthenticatorType,omitempty"`
	ClientOfflineSessionIdleTimeout        *string                   `json:"clientOfflineSessionIdleTimeout,omitempty"`
	ClientOfflineSessionMaxLifespan        *string                   `json:"clientOfflineSessionMaxLifespan,omitempty"`
	ClientSecretSecretRef                  *v1.SecretKeySelector     `json:"clientSecretSecretRef,omitempty"`
	ClientSessionIdleTimeout               *string                   `json:"clientSessionIdleTimeout,omitempty"`
	ClientSessionMaxLifespan               *string                   `json:"clientSessionMaxLifespan,omitempty"`
	ConsentRequired                        *bool                     `json:"consentRequired,omitempty"`
	ConsentScreenText                      *string                   `json:"consentScreenText,omitempty"`
	DirectAccessGrantsEnabled              *bool                     `json:"directAccessGrantsEnabled,omitempty"`
	DisplayOnConsentScreen                 *bool                     `json:"displayOnConsentScreen,omitempty"`
	ExcludeSessionStateFromAuthResponse    *bool                     `json:"excludeSessionStateFromAuthResponse,omitempty"`
	FrontchannelLogoutEnabled              *bool                     `json:"frontchannelLogoutEnabled,omitempty"`
	FrontchannelLogoutURL                  *string                   `json:"frontchannelLogoutUrl,omitempty"`
	ImplicitFlowEnabled                    *bool                     `json:"implicitFlowEnabled,omitempty"`
	Import                                 *bool                     `json:"import,omitempty"`
	Oauth2DeviceAuthorizationGrantEnabled  *bool                     `json:"oauth2DeviceAuthorizationGrantEnabled,omitempty"`
	Oauth2DeviceCodeLifespan               *string                   `json:"oauth2DeviceCodeLifespan,omitempty"`
	Oauth2DevicePollingInterval            *string                   `json:"oauth2DevicePollingInterval,omitempty"`
	PkceCodeChallengeMethod                *string                   `json:"pkceCodeChallengeMethod,omitempty"`
	RootURL                                *string                   `json:"rootUrl,omitempty"`
	ServiceAccountsEnabled                 *bool                     `json:"serviceAccountsEnabled,omitempty"`
	StandardFlowEnabled                    *bool                     `json:"standardFlowEnabled,omitempty"`
	UseRefreshTokens                       *bool                     `json:"useRefreshTokens,omitempty"`
	UseRefreshTokensClientCredentials      *bool                     `json:"useRefreshTokensClientCredentials,omitempty"`
	ValidPostLogoutRedirectUris            []*string                 `json:"validPostLogoutRedirectUris,omitempty"`
	ValidRedirectUris                      []*string                 `json:"validRedirectUris,omitempty"`
	WebOrigins                             []*string                 `json:"webOrigins,omitempty"`
}

type SamlClientAttributes struct {
	IncludeAuthnStatement           *bool  `json:"saml.authnstatement"`
	SignDocuments                   *bool  `json:"saml.server.signature"`
	SignAssertions                  *bool  `json:"saml.assertion.signature"`
	EncryptAssertions               *bool  `json:"saml.encrypt"`
	ClientSignatureRequired         *bool  `json:"saml.client.signature"`
	ForcePostBinding                *bool  `json:"saml.force.post.binding"`
	ForceNameIdFormat               *bool  `json:"saml_force_name_id_format"`
	SignatureAlgorithm              string `json:"saml.signature.algorithm"`
	SignatureKeyName                string `json:"saml.server.signature.keyinfo.xmlSigKeyInfoKeyNameTransformer"`
	CanonicalizationMethod          string `json:"saml_signature_canonicalization_method"`
	NameIdFormat                    string `json:"saml_name_id_format"`
	SigningCertificate              string `json:"saml.signing.certificate,omitempty"`
	SigningPrivateKey               string `json:"saml.signing.private.key"`
	EncryptionCertificate           string `json:"saml.encryption.certificate"`
	IDPInitiatedSSOURLName          string `json:"saml_idp_initiated_sso_url_name"`
	IDPInitiatedSSORelayState       string `json:"saml_idp_initiated_sso_relay_state"`
	AssertionConsumerPostURL        string `json:"saml_assertion_consumer_url_post"`
	AssertionConsumerRedirectURL    string `json:"saml_assertion_consumer_url_redirect"`
	LogoutServicePostBindingURL     string `json:"saml_single_logout_service_url_post"`
	LogoutServiceRedirectBindingURL string `json:"saml_single_logout_service_url_redirect"`
	LoginTheme                      string `json:"login_theme"`
}

// SamlConfig contains the fields specific to SAML clients.
type SamlConfig struct {
	Attributes                      *SamlClientAttributes `json:"attributes,omitempty"`
	AssertionConsumerPostURL        *string               `json:"assertionConsumerPostUrl,omitempty"`
	AssertionConsumerRedirectURL    *string               `json:"assertionConsumerRedirectUrl,omitempty"`
	CanonicalizationMethod          *string               `json:"canonicalizationMethod,omitempty"`
	ClientSignatureRequired         *bool                 `json:"clientSignatureRequired,omitempty"`
	EncryptAssertions               *bool                 `json:"encryptAssertions,omitempty"`
	EncryptionCertificate           *string               `json:"encryptionCertificate,omitempty"`
	ForceNameIDFormat               *bool                 `json:"forceNameIdFormat,omitempty"`
	ForcePostBinding                *bool                 `json:"forcePostBinding,omitempty"`
	FrontChannelLogout              *bool                 `json:"frontChannelLogout,omitempty"`
	IdpInitiatedSsoRelayState       *string               `json:"idpInitiatedSsoRelayState,omitempty"`
	IdpInitiatedSsoURLName          *string               `json:"idpInitiatedSsoUrlName,omitempty"`
	IncludeAuthnStatement           *bool                 `json:"includeAuthnStatement,omitempty"`
	LogoutServicePostBindingURL     *string               `json:"logoutServicePostBindingUrl,omitempty"`
	LogoutServiceRedirectBindingURL *string               `json:"logoutServiceRedirectBindingUrl,omitempty"`
	MasterSAMLProcessingURL         *string               `json:"masterSamlProcessingUrl,omitempty"`
	NameIDFormat                    *string               `json:"nameIdFormat,omitempty"`
	RootURL                         *string               `json:"rootUrl,omitempty"`
	SignAssertions                  *bool                 `json:"signAssertions,omitempty"`
	SignDocuments                   *bool                 `json:"signDocuments,omitempty"`
	SignatureAlgorithm              *string               `json:"signatureAlgorithm,omitempty"`
	SignatureKeyName                *string               `json:"signatureKeyName,omitempty"`
	SigningCertificate              *string               `json:"signingCertificate,omitempty"`
	SigningPrivateKey               *string               `json:"signingPrivateKey,omitempty"`
	ValidRedirectUris               []*string             `json:"validRedirectUris,omitempty"`
}

// ClientInitParameters are the configurable fields of a Client.
type ClientInitParameters struct {
	ClientID string `json:"clientId"`

	// The realm this client is attached to.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/realm/v1alpha1.Realm
	RealmID *string `json:"realmId,omitempty" tf:"realm_id,omitempty"`

	// Reference to a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDRef *v1.Reference `json:"realmIdRef,omitempty" tf:"-"`

	// Selector for a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDSelector *v1.Selector `json:"realmIdSelector,omitempty" tf:"-"`

	Name                               *string                                        `json:"name,omitempty"`
	Description                        *string                                        `json:"description,omitempty"`
	BaseURL                            *string                                        `json:"baseUrl,omitempty"`
	Enabled                            *bool                                          `json:"enabled,omitempty"`
	LoginTheme                         *string                                        `json:"loginTheme,omitempty"`
	ExtraConfig                        map[string]*string                             `json:"extraConfig,omitempty"`
	FullScopeAllowed                   *bool                                          `json:"fullScopeAllowed,omitempty"`
	AuthenticationFlowBindingOverrides []AuthenticationFlowBindingOverridesParameters `json:"authenticationFlowBindingOverrides,omitempty"`
	OpenIdConfig                       OpenIDConfig                                   `json:"openIdConfig,omitempty"`
	SamlConfig                         SamlConfig                                     `json:"samlConfig,omitempty"`
}

// ClientObservation are the configurable fields of a Client.
type ClientObservation struct {
	ClientID string `json:"clientId"`

	// The realm this client is attached to.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/realm/v1alpha1.Realm
	RealmID *string `json:"realmId,omitempty" tf:"realm_id,omitempty"`

	// Reference to a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDRef *v1.Reference `json:"realmIdRef,omitempty" tf:"-"`

	// Selector for a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDSelector *v1.Selector `json:"realmIdSelector,omitempty" tf:"-"`

	Name                               *string                                        `json:"name,omitempty"`
	Description                        *string                                        `json:"description,omitempty"`
	BaseURL                            *string                                        `json:"baseUrl,omitempty"`
	Enabled                            *bool                                          `json:"enabled,omitempty"`
	LoginTheme                         *string                                        `json:"loginTheme,omitempty"`
	ExtraConfig                        map[string]*string                             `json:"extraConfig,omitempty"`
	FullScopeAllowed                   *bool                                          `json:"fullScopeAllowed,omitempty"`
	AuthenticationFlowBindingOverrides []AuthenticationFlowBindingOverridesParameters `json:"authenticationFlowBindingOverrides,omitempty"`
	OpenIdConfig                       OpenIDConfig                                   `json:"openIdConfig,omitempty"`
	SamlConfig                         SamlConfig                                     `json:"samlConfig,omitempty"`
}

// ClientParameters are the configurable fields of a Client.
type ClientParameters struct {
	ClientID string `json:"clientId"`

	// The realm this client is attached to.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/realm/v1alpha1.Realm
	RealmID *string `json:"realmId,omitempty" tf:"realm_id,omitempty"`

	// Reference to a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDRef *v1.Reference `json:"realmIdRef,omitempty" tf:"-"`

	// Selector for a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDSelector *v1.Selector `json:"realmIdSelector,omitempty" tf:"-"`

	Name                               *string                                        `json:"name,omitempty"`
	Description                        *string                                        `json:"description,omitempty"`
	BaseURL                            *string                                        `json:"baseUrl,omitempty"`
	Enabled                            *bool                                          `json:"enabled,omitempty"`
	LoginTheme                         *string                                        `json:"loginTheme,omitempty"`
	ExtraConfig                        map[string]*string                             `json:"extraConfig,omitempty"`
	FullScopeAllowed                   *bool                                          `json:"fullScopeAllowed,omitempty"`
	AuthenticationFlowBindingOverrides []AuthenticationFlowBindingOverridesParameters `json:"authenticationFlowBindingOverrides,omitempty"`
	OpenIdConfig                       OpenIDConfig                                   `json:"openIdConfig,omitempty"`
	SamlConfig                         SamlConfig                                     `json:"samlConfig,omitempty"`
}

// A ClientSpec defines the desired state of a Client.
type ClientSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     ClientParameters `json:"forProvider"`
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
	InitProvider ClientInitParameters `json:"initProvider,omitempty"`
}

// A ClientStatus represents the observed state of a Client.
type ClientStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        ClientObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Client is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,keycloak}
type Client struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClientSpec   `json:"spec"`
	Status ClientStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClientList contains a list of Client
type ClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Client `json:"items"`
}

// Client type metadata.
var (
	ClientKind             = reflect.TypeOf(Client{}).Name()
	ClientGroupKind        = schema.GroupKind{Group: Group, Kind: ClientKind}.String()
	ClientKindAPIVersion   = ClientKind + "." + SchemeGroupVersion.String()
	ClientGroupVersionKind = SchemeGroupVersion.WithKind(ClientKind)
)

func init() {
	SchemeBuilder.Register(&Client{}, &ClientList{})
}
