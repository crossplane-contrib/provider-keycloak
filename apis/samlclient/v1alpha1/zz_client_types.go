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

type AuthenticationFlowBindingOverridesInitParameters struct {

	// Browser flow id, (flow needs to exist)
	BrowserID *string `json:"browserId,omitempty" tf:"browser_id,omitempty"`

	// Direct grant flow id (flow needs to exist)
	DirectGrantID *string `json:"directGrantId,omitempty" tf:"direct_grant_id,omitempty"`
}

type AuthenticationFlowBindingOverridesObservation struct {

	// Browser flow id, (flow needs to exist)
	BrowserID *string `json:"browserId,omitempty" tf:"browser_id,omitempty"`

	// Direct grant flow id (flow needs to exist)
	DirectGrantID *string `json:"directGrantId,omitempty" tf:"direct_grant_id,omitempty"`
}

type AuthenticationFlowBindingOverridesParameters struct {

	// Browser flow id, (flow needs to exist)
	// +kubebuilder:validation:Optional
	BrowserID *string `json:"browserId,omitempty" tf:"browser_id,omitempty"`

	// Direct grant flow id (flow needs to exist)
	// +kubebuilder:validation:Optional
	DirectGrantID *string `json:"directGrantId,omitempty" tf:"direct_grant_id,omitempty"`
}

type ClientInitParameters struct {

	// Always list this client in the Account UI, even if the user does not have an active session.
	AlwaysDisplayInConsole *bool `json:"alwaysDisplayInConsole,omitempty" tf:"always_display_in_console,omitempty"`

	// SAML POST Binding URL for the client's assertion consumer service (login responses).
	AssertionConsumerPostURL *string `json:"assertionConsumerPostUrl,omitempty" tf:"assertion_consumer_post_url,omitempty"`

	// SAML Redirect Binding URL for the client's assertion consumer service (login responses).
	AssertionConsumerRedirectURL *string `json:"assertionConsumerRedirectUrl,omitempty" tf:"assertion_consumer_redirect_url,omitempty"`

	// Override realm authentication flow bindings
	AuthenticationFlowBindingOverrides []AuthenticationFlowBindingOverridesInitParameters `json:"authenticationFlowBindingOverrides,omitempty" tf:"authentication_flow_binding_overrides,omitempty"`

	// When specified, this URL will be used whenever Keycloak needs to link to this client.
	BaseURL *string `json:"baseUrl,omitempty" tf:"base_url,omitempty"`

	// The Canonicalization Method for XML signatures. Should be one of "EXCLUSIVE", "EXCLUSIVE_WITH_COMMENTS", "INCLUSIVE", or "INCLUSIVE_WITH_COMMENTS". Defaults to "EXCLUSIVE".
	CanonicalizationMethod *string `json:"canonicalizationMethod,omitempty" tf:"canonicalization_method,omitempty"`

	// The unique ID of this client, referenced in the URI during authentication and in issued tokens.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/openidclient/v1alpha1.Client
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-keycloak/config/common.UUIDExtractor()
	ClientID *string `json:"clientId,omitempty" tf:"client_id,omitempty"`

	// Reference to a Client in openidclient to populate clientId.
	// +kubebuilder:validation:Optional
	ClientIDRef *v1.Reference `json:"clientIdRef,omitempty" tf:"-"`

	// Selector for a Client in openidclient to populate clientId.
	// +kubebuilder:validation:Optional
	ClientIDSelector *v1.Selector `json:"clientIdSelector,omitempty" tf:"-"`

	// When true, Keycloak will expect that documents originating from a client will be signed using the certificate and/or key configured via signing_certificate and signing_private_key. Defaults to true.
	ClientSignatureRequired *bool `json:"clientSignatureRequired,omitempty" tf:"client_signature_required,omitempty"`

	// When true, users have to consent to client access. Defaults to false.
	ConsentRequired *bool `json:"consentRequired,omitempty" tf:"consent_required,omitempty"`

	// The description of this client in the GUI.
	Description *string `json:"description,omitempty" tf:"description,omitempty"`

	// When false, this client will not be able to initiate a login or obtain access tokens. Defaults to true.
	Enabled *bool `json:"enabled,omitempty" tf:"enabled,omitempty"`

	// When true, the SAML assertions will be encrypted by Keycloak using the client's public key. Defaults to false.
	EncryptAssertions *bool `json:"encryptAssertions,omitempty" tf:"encrypt_assertions,omitempty"`

	// If assertions for the client are encrypted, this certificate will be used for encryption.
	EncryptionCertificate *string `json:"encryptionCertificate,omitempty" tf:"encryption_certificate,omitempty"`

	// A map of key/value pairs to add extra configuration attributes to this client. Use this attribute at your own risk, as s may conflict with top-level configuration attributes in future provider updates.
	// +mapType=granular
	ExtraConfig map[string]*string `json:"extraConfig,omitempty" tf:"extra_config,omitempty"`

	// Ignore requested NameID subject format and use the one defined in name_id_format instead. Defaults to false.
	ForceNameIDFormat *bool `json:"forceNameIdFormat,omitempty" tf:"force_name_id_format,omitempty"`

	// When true, Keycloak will always respond to an authentication request via the SAML POST Binding. Defaults to true.
	ForcePostBinding *bool `json:"forcePostBinding,omitempty" tf:"force_post_binding,omitempty"`

	// When true, this client will require a browser redirect in order to perform a logout. Defaults to true.
	FrontChannelLogout *bool `json:"frontChannelLogout,omitempty" tf:"front_channel_logout,omitempty"`

	// - Allow to include all roles mappings in the access token
	FullScopeAllowed *bool `json:"fullScopeAllowed,omitempty" tf:"full_scope_allowed,omitempty"`

	// Relay state you want to send with SAML request when you want to do IDP Initiated SSO.
	IdpInitiatedSsoRelayState *string `json:"idpInitiatedSsoRelayState,omitempty" tf:"idp_initiated_sso_relay_state,omitempty"`

	// URL fragment name to reference client when you want to do IDP Initiated SSO.
	IdpInitiatedSsoURLName *string `json:"idpInitiatedSsoUrlName,omitempty" tf:"idp_initiated_sso_url_name,omitempty"`

	// When true, an AuthnStatement will be included in the SAML response. Defaults to true.
	IncludeAuthnStatement *bool `json:"includeAuthnStatement,omitempty" tf:"include_authn_statement,omitempty"`

	// The login theme of this client.
	LoginTheme *string `json:"loginTheme,omitempty" tf:"login_theme,omitempty"`

	// SAML POST Binding URL for the client's single logout service.
	LogoutServicePostBindingURL *string `json:"logoutServicePostBindingUrl,omitempty" tf:"logout_service_post_binding_url,omitempty"`

	// SAML Redirect Binding URL for the client's single logout service.
	LogoutServiceRedirectBindingURL *string `json:"logoutServiceRedirectBindingUrl,omitempty" tf:"logout_service_redirect_binding_url,omitempty"`

	// When specified, this URL will be used for all SAML requests.
	MasterSAMLProcessingURL *string `json:"masterSamlProcessingUrl,omitempty" tf:"master_saml_processing_url,omitempty"`

	// The display name of this client in the GUI.
	Name *string `json:"name,omitempty" tf:"name,omitempty"`

	// Sets the Name ID format for the subject.
	NameIDFormat *string `json:"nameIdFormat,omitempty" tf:"name_id_format,omitempty"`

	// The realm this client is attached to.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/realm/v1alpha1.Realm
	RealmID *string `json:"realmId,omitempty" tf:"realm_id,omitempty"`

	// Reference to a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDRef *v1.Reference `json:"realmIdRef,omitempty" tf:"-"`

	// Selector for a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDSelector *v1.Selector `json:"realmIdSelector,omitempty" tf:"-"`

	// When specified, this value is prepended to all relative URLs.
	RootURL *string `json:"rootUrl,omitempty" tf:"root_url,omitempty"`

	// When true, the SAML assertions will be signed by Keycloak using the realm's private key, and embedded within the SAML XML Auth response. Defaults to false.
	SignAssertions *bool `json:"signAssertions,omitempty" tf:"sign_assertions,omitempty"`

	// When true, the SAML document will be signed by Keycloak using the realm's private key. Defaults to true.
	SignDocuments *bool `json:"signDocuments,omitempty" tf:"sign_documents,omitempty"`

	// The signature algorithm used to sign documents. Should be one of "RSA_SHA1", "RSA_SHA256", "RSA_SHA256_MGF1, "RSA_SHA512", "RSA_SHA512_MGF1" or "DSA_SHA1".
	SignatureAlgorithm *string `json:"signatureAlgorithm,omitempty" tf:"signature_algorithm,omitempty"`

	// The value of the KeyName element within the signed SAML document. Should be one of "NONE", "KEY_ID", or "CERT_SUBJECT". Defaults to "KEY_ID".
	SignatureKeyName *string `json:"signatureKeyName,omitempty" tf:"signature_key_name,omitempty"`

	// If documents or assertions from the client are signed, this certificate will be used to verify the signature.
	SigningCertificate *string `json:"signingCertificate,omitempty" tf:"signing_certificate,omitempty"`

	// If documents or assertions from the client are signed, this private key will be used to verify the signature.
	SigningPrivateKey *string `json:"signingPrivateKey,omitempty" tf:"signing_private_key,omitempty"`

	// When specified, Keycloak will use this list to validate given Assertion Consumer URLs specified in the authentication request.
	// +listType=set
	ValidRedirectUris []*string `json:"validRedirectUris,omitempty" tf:"valid_redirect_uris,omitempty"`
}

type ClientObservation struct {

	// Always list this client in the Account UI, even if the user does not have an active session.
	AlwaysDisplayInConsole *bool `json:"alwaysDisplayInConsole,omitempty" tf:"always_display_in_console,omitempty"`

	// SAML POST Binding URL for the client's assertion consumer service (login responses).
	AssertionConsumerPostURL *string `json:"assertionConsumerPostUrl,omitempty" tf:"assertion_consumer_post_url,omitempty"`

	// SAML Redirect Binding URL for the client's assertion consumer service (login responses).
	AssertionConsumerRedirectURL *string `json:"assertionConsumerRedirectUrl,omitempty" tf:"assertion_consumer_redirect_url,omitempty"`

	// Override realm authentication flow bindings
	AuthenticationFlowBindingOverrides []AuthenticationFlowBindingOverridesObservation `json:"authenticationFlowBindingOverrides,omitempty" tf:"authentication_flow_binding_overrides,omitempty"`

	// When specified, this URL will be used whenever Keycloak needs to link to this client.
	BaseURL *string `json:"baseUrl,omitempty" tf:"base_url,omitempty"`

	// The Canonicalization Method for XML signatures. Should be one of "EXCLUSIVE", "EXCLUSIVE_WITH_COMMENTS", "INCLUSIVE", or "INCLUSIVE_WITH_COMMENTS". Defaults to "EXCLUSIVE".
	CanonicalizationMethod *string `json:"canonicalizationMethod,omitempty" tf:"canonicalization_method,omitempty"`

	// The unique ID of this client, referenced in the URI during authentication and in issued tokens.
	ClientID *string `json:"clientId,omitempty" tf:"client_id,omitempty"`

	// When true, Keycloak will expect that documents originating from a client will be signed using the certificate and/or key configured via signing_certificate and signing_private_key. Defaults to true.
	ClientSignatureRequired *bool `json:"clientSignatureRequired,omitempty" tf:"client_signature_required,omitempty"`

	// When true, users have to consent to client access. Defaults to false.
	ConsentRequired *bool `json:"consentRequired,omitempty" tf:"consent_required,omitempty"`

	// The description of this client in the GUI.
	Description *string `json:"description,omitempty" tf:"description,omitempty"`

	// When false, this client will not be able to initiate a login or obtain access tokens. Defaults to true.
	Enabled *bool `json:"enabled,omitempty" tf:"enabled,omitempty"`

	// When true, the SAML assertions will be encrypted by Keycloak using the client's public key. Defaults to false.
	EncryptAssertions *bool `json:"encryptAssertions,omitempty" tf:"encrypt_assertions,omitempty"`

	// If assertions for the client are encrypted, this certificate will be used for encryption.
	EncryptionCertificate *string `json:"encryptionCertificate,omitempty" tf:"encryption_certificate,omitempty"`

	// (Computed) The sha1sum fingerprint of the encryption certificate. If the encryption certificate is not in correct base64 format, this will be left empty.
	EncryptionCertificateSha1 *string `json:"encryptionCertificateSha1,omitempty" tf:"encryption_certificate_sha1,omitempty"`

	// A map of key/value pairs to add extra configuration attributes to this client. Use this attribute at your own risk, as s may conflict with top-level configuration attributes in future provider updates.
	// +mapType=granular
	ExtraConfig map[string]*string `json:"extraConfig,omitempty" tf:"extra_config,omitempty"`

	// Ignore requested NameID subject format and use the one defined in name_id_format instead. Defaults to false.
	ForceNameIDFormat *bool `json:"forceNameIdFormat,omitempty" tf:"force_name_id_format,omitempty"`

	// When true, Keycloak will always respond to an authentication request via the SAML POST Binding. Defaults to true.
	ForcePostBinding *bool `json:"forcePostBinding,omitempty" tf:"force_post_binding,omitempty"`

	// When true, this client will require a browser redirect in order to perform a logout. Defaults to true.
	FrontChannelLogout *bool `json:"frontChannelLogout,omitempty" tf:"front_channel_logout,omitempty"`

	// - Allow to include all roles mappings in the access token
	FullScopeAllowed *bool `json:"fullScopeAllowed,omitempty" tf:"full_scope_allowed,omitempty"`

	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	// Relay state you want to send with SAML request when you want to do IDP Initiated SSO.
	IdpInitiatedSsoRelayState *string `json:"idpInitiatedSsoRelayState,omitempty" tf:"idp_initiated_sso_relay_state,omitempty"`

	// URL fragment name to reference client when you want to do IDP Initiated SSO.
	IdpInitiatedSsoURLName *string `json:"idpInitiatedSsoUrlName,omitempty" tf:"idp_initiated_sso_url_name,omitempty"`

	// When true, an AuthnStatement will be included in the SAML response. Defaults to true.
	IncludeAuthnStatement *bool `json:"includeAuthnStatement,omitempty" tf:"include_authn_statement,omitempty"`

	// The login theme of this client.
	LoginTheme *string `json:"loginTheme,omitempty" tf:"login_theme,omitempty"`

	// SAML POST Binding URL for the client's single logout service.
	LogoutServicePostBindingURL *string `json:"logoutServicePostBindingUrl,omitempty" tf:"logout_service_post_binding_url,omitempty"`

	// SAML Redirect Binding URL for the client's single logout service.
	LogoutServiceRedirectBindingURL *string `json:"logoutServiceRedirectBindingUrl,omitempty" tf:"logout_service_redirect_binding_url,omitempty"`

	// When specified, this URL will be used for all SAML requests.
	MasterSAMLProcessingURL *string `json:"masterSamlProcessingUrl,omitempty" tf:"master_saml_processing_url,omitempty"`

	// The display name of this client in the GUI.
	Name *string `json:"name,omitempty" tf:"name,omitempty"`

	// Sets the Name ID format for the subject.
	NameIDFormat *string `json:"nameIdFormat,omitempty" tf:"name_id_format,omitempty"`

	// The realm this client is attached to.
	RealmID *string `json:"realmId,omitempty" tf:"realm_id,omitempty"`

	// When specified, this value is prepended to all relative URLs.
	RootURL *string `json:"rootUrl,omitempty" tf:"root_url,omitempty"`

	// When true, the SAML assertions will be signed by Keycloak using the realm's private key, and embedded within the SAML XML Auth response. Defaults to false.
	SignAssertions *bool `json:"signAssertions,omitempty" tf:"sign_assertions,omitempty"`

	// When true, the SAML document will be signed by Keycloak using the realm's private key. Defaults to true.
	SignDocuments *bool `json:"signDocuments,omitempty" tf:"sign_documents,omitempty"`

	// The signature algorithm used to sign documents. Should be one of "RSA_SHA1", "RSA_SHA256", "RSA_SHA256_MGF1, "RSA_SHA512", "RSA_SHA512_MGF1" or "DSA_SHA1".
	SignatureAlgorithm *string `json:"signatureAlgorithm,omitempty" tf:"signature_algorithm,omitempty"`

	// The value of the KeyName element within the signed SAML document. Should be one of "NONE", "KEY_ID", or "CERT_SUBJECT". Defaults to "KEY_ID".
	SignatureKeyName *string `json:"signatureKeyName,omitempty" tf:"signature_key_name,omitempty"`

	// If documents or assertions from the client are signed, this certificate will be used to verify the signature.
	SigningCertificate *string `json:"signingCertificate,omitempty" tf:"signing_certificate,omitempty"`

	// (Computed) The sha1sum fingerprint of the signing certificate. If the signing certificate is not in correct base64 format, this will be left empty.
	SigningCertificateSha1 *string `json:"signingCertificateSha1,omitempty" tf:"signing_certificate_sha1,omitempty"`

	// If documents or assertions from the client are signed, this private key will be used to verify the signature.
	SigningPrivateKey *string `json:"signingPrivateKey,omitempty" tf:"signing_private_key,omitempty"`

	// (Computed) The sha1sum fingerprint of the signing private key. If the signing private key is not in correct base64 format, this will be left empty.
	SigningPrivateKeySha1 *string `json:"signingPrivateKeySha1,omitempty" tf:"signing_private_key_sha1,omitempty"`

	// When specified, Keycloak will use this list to validate given Assertion Consumer URLs specified in the authentication request.
	// +listType=set
	ValidRedirectUris []*string `json:"validRedirectUris,omitempty" tf:"valid_redirect_uris,omitempty"`
}

type ClientParameters struct {

	// Always list this client in the Account UI, even if the user does not have an active session.
	// +kubebuilder:validation:Optional
	AlwaysDisplayInConsole *bool `json:"alwaysDisplayInConsole,omitempty" tf:"always_display_in_console,omitempty"`

	// SAML POST Binding URL for the client's assertion consumer service (login responses).
	// +kubebuilder:validation:Optional
	AssertionConsumerPostURL *string `json:"assertionConsumerPostUrl,omitempty" tf:"assertion_consumer_post_url,omitempty"`

	// SAML Redirect Binding URL for the client's assertion consumer service (login responses).
	// +kubebuilder:validation:Optional
	AssertionConsumerRedirectURL *string `json:"assertionConsumerRedirectUrl,omitempty" tf:"assertion_consumer_redirect_url,omitempty"`

	// Override realm authentication flow bindings
	// +kubebuilder:validation:Optional
	AuthenticationFlowBindingOverrides []AuthenticationFlowBindingOverridesParameters `json:"authenticationFlowBindingOverrides,omitempty" tf:"authentication_flow_binding_overrides,omitempty"`

	// When specified, this URL will be used whenever Keycloak needs to link to this client.
	// +kubebuilder:validation:Optional
	BaseURL *string `json:"baseUrl,omitempty" tf:"base_url,omitempty"`

	// The Canonicalization Method for XML signatures. Should be one of "EXCLUSIVE", "EXCLUSIVE_WITH_COMMENTS", "INCLUSIVE", or "INCLUSIVE_WITH_COMMENTS". Defaults to "EXCLUSIVE".
	// +kubebuilder:validation:Optional
	CanonicalizationMethod *string `json:"canonicalizationMethod,omitempty" tf:"canonicalization_method,omitempty"`

	// The unique ID of this client, referenced in the URI during authentication and in issued tokens.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/openidclient/v1alpha1.Client
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-keycloak/config/common.UUIDExtractor()
	// +kubebuilder:validation:Optional
	ClientID *string `json:"clientId,omitempty" tf:"client_id,omitempty"`

	// Reference to a Client in openidclient to populate clientId.
	// +kubebuilder:validation:Optional
	ClientIDRef *v1.Reference `json:"clientIdRef,omitempty" tf:"-"`

	// Selector for a Client in openidclient to populate clientId.
	// +kubebuilder:validation:Optional
	ClientIDSelector *v1.Selector `json:"clientIdSelector,omitempty" tf:"-"`

	// When true, Keycloak will expect that documents originating from a client will be signed using the certificate and/or key configured via signing_certificate and signing_private_key. Defaults to true.
	// +kubebuilder:validation:Optional
	ClientSignatureRequired *bool `json:"clientSignatureRequired,omitempty" tf:"client_signature_required,omitempty"`

	// When true, users have to consent to client access. Defaults to false.
	// +kubebuilder:validation:Optional
	ConsentRequired *bool `json:"consentRequired,omitempty" tf:"consent_required,omitempty"`

	// The description of this client in the GUI.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty" tf:"description,omitempty"`

	// When false, this client will not be able to initiate a login or obtain access tokens. Defaults to true.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty" tf:"enabled,omitempty"`

	// When true, the SAML assertions will be encrypted by Keycloak using the client's public key. Defaults to false.
	// +kubebuilder:validation:Optional
	EncryptAssertions *bool `json:"encryptAssertions,omitempty" tf:"encrypt_assertions,omitempty"`

	// If assertions for the client are encrypted, this certificate will be used for encryption.
	// +kubebuilder:validation:Optional
	EncryptionCertificate *string `json:"encryptionCertificate,omitempty" tf:"encryption_certificate,omitempty"`

	// A map of key/value pairs to add extra configuration attributes to this client. Use this attribute at your own risk, as s may conflict with top-level configuration attributes in future provider updates.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	ExtraConfig map[string]*string `json:"extraConfig,omitempty" tf:"extra_config,omitempty"`

	// Ignore requested NameID subject format and use the one defined in name_id_format instead. Defaults to false.
	// +kubebuilder:validation:Optional
	ForceNameIDFormat *bool `json:"forceNameIdFormat,omitempty" tf:"force_name_id_format,omitempty"`

	// When true, Keycloak will always respond to an authentication request via the SAML POST Binding. Defaults to true.
	// +kubebuilder:validation:Optional
	ForcePostBinding *bool `json:"forcePostBinding,omitempty" tf:"force_post_binding,omitempty"`

	// When true, this client will require a browser redirect in order to perform a logout. Defaults to true.
	// +kubebuilder:validation:Optional
	FrontChannelLogout *bool `json:"frontChannelLogout,omitempty" tf:"front_channel_logout,omitempty"`

	// - Allow to include all roles mappings in the access token
	// +kubebuilder:validation:Optional
	FullScopeAllowed *bool `json:"fullScopeAllowed,omitempty" tf:"full_scope_allowed,omitempty"`

	// Relay state you want to send with SAML request when you want to do IDP Initiated SSO.
	// +kubebuilder:validation:Optional
	IdpInitiatedSsoRelayState *string `json:"idpInitiatedSsoRelayState,omitempty" tf:"idp_initiated_sso_relay_state,omitempty"`

	// URL fragment name to reference client when you want to do IDP Initiated SSO.
	// +kubebuilder:validation:Optional
	IdpInitiatedSsoURLName *string `json:"idpInitiatedSsoUrlName,omitempty" tf:"idp_initiated_sso_url_name,omitempty"`

	// When true, an AuthnStatement will be included in the SAML response. Defaults to true.
	// +kubebuilder:validation:Optional
	IncludeAuthnStatement *bool `json:"includeAuthnStatement,omitempty" tf:"include_authn_statement,omitempty"`

	// The login theme of this client.
	// +kubebuilder:validation:Optional
	LoginTheme *string `json:"loginTheme,omitempty" tf:"login_theme,omitempty"`

	// SAML POST Binding URL for the client's single logout service.
	// +kubebuilder:validation:Optional
	LogoutServicePostBindingURL *string `json:"logoutServicePostBindingUrl,omitempty" tf:"logout_service_post_binding_url,omitempty"`

	// SAML Redirect Binding URL for the client's single logout service.
	// +kubebuilder:validation:Optional
	LogoutServiceRedirectBindingURL *string `json:"logoutServiceRedirectBindingUrl,omitempty" tf:"logout_service_redirect_binding_url,omitempty"`

	// When specified, this URL will be used for all SAML requests.
	// +kubebuilder:validation:Optional
	MasterSAMLProcessingURL *string `json:"masterSamlProcessingUrl,omitempty" tf:"master_saml_processing_url,omitempty"`

	// The display name of this client in the GUI.
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty" tf:"name,omitempty"`

	// Sets the Name ID format for the subject.
	// +kubebuilder:validation:Optional
	NameIDFormat *string `json:"nameIdFormat,omitempty" tf:"name_id_format,omitempty"`

	// The realm this client is attached to.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-keycloak/apis/realm/v1alpha1.Realm
	// +kubebuilder:validation:Optional
	RealmID *string `json:"realmId,omitempty" tf:"realm_id,omitempty"`

	// Reference to a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDRef *v1.Reference `json:"realmIdRef,omitempty" tf:"-"`

	// Selector for a Realm in realm to populate realmId.
	// +kubebuilder:validation:Optional
	RealmIDSelector *v1.Selector `json:"realmIdSelector,omitempty" tf:"-"`

	// When specified, this value is prepended to all relative URLs.
	// +kubebuilder:validation:Optional
	RootURL *string `json:"rootUrl,omitempty" tf:"root_url,omitempty"`

	// When true, the SAML assertions will be signed by Keycloak using the realm's private key, and embedded within the SAML XML Auth response. Defaults to false.
	// +kubebuilder:validation:Optional
	SignAssertions *bool `json:"signAssertions,omitempty" tf:"sign_assertions,omitempty"`

	// When true, the SAML document will be signed by Keycloak using the realm's private key. Defaults to true.
	// +kubebuilder:validation:Optional
	SignDocuments *bool `json:"signDocuments,omitempty" tf:"sign_documents,omitempty"`

	// The signature algorithm used to sign documents. Should be one of "RSA_SHA1", "RSA_SHA256", "RSA_SHA256_MGF1, "RSA_SHA512", "RSA_SHA512_MGF1" or "DSA_SHA1".
	// +kubebuilder:validation:Optional
	SignatureAlgorithm *string `json:"signatureAlgorithm,omitempty" tf:"signature_algorithm,omitempty"`

	// The value of the KeyName element within the signed SAML document. Should be one of "NONE", "KEY_ID", or "CERT_SUBJECT". Defaults to "KEY_ID".
	// +kubebuilder:validation:Optional
	SignatureKeyName *string `json:"signatureKeyName,omitempty" tf:"signature_key_name,omitempty"`

	// If documents or assertions from the client are signed, this certificate will be used to verify the signature.
	// +kubebuilder:validation:Optional
	SigningCertificate *string `json:"signingCertificate,omitempty" tf:"signing_certificate,omitempty"`

	// If documents or assertions from the client are signed, this private key will be used to verify the signature.
	// +kubebuilder:validation:Optional
	SigningPrivateKey *string `json:"signingPrivateKey,omitempty" tf:"signing_private_key,omitempty"`

	// When specified, Keycloak will use this list to validate given Assertion Consumer URLs specified in the authentication request.
	// +kubebuilder:validation:Optional
	// +listType=set
	ValidRedirectUris []*string `json:"validRedirectUris,omitempty" tf:"valid_redirect_uris,omitempty"`
}

// ClientSpec defines the desired state of Client
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

// ClientStatus defines the observed state of Client.
type ClientStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        ClientObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// Client is the Schema for the Clients API.
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,keycloak}
type Client struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClientSpec   `json:"spec"`
	Status            ClientStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClientList contains a list of Clients
type ClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Client `json:"items"`
}

// Repository type metadata.
var (
	Client_Kind             = "Client"
	Client_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: Client_Kind}.String()
	Client_KindAPIVersion   = Client_Kind + "." + CRDGroupVersion.String()
	Client_GroupVersionKind = CRDGroupVersion.WithKind(Client_Kind)
)

func init() {
	SchemeBuilder.Register(&Client{}, &ClientList{})
}
