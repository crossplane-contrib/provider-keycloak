package mapper

import (
	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane/upjet/pkg/config"
	ujconfig "github.com/crossplane/upjet/pkg/config"
)

// Group is the short group name for the resources in this package
var Group = "mapper"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *ujconfig.Provider) {

	//p.AddResourceConfigurator("keycloak_generic_protocol_mapper", func(r *ujconfig.Resource) {
	//	r.ShortGroup = Group
	//	r.Name = "keycloak_generic_protocol_mapper"
	//	r.Kind = "OpenIdProtocolMapper"
	//	r.References["client_scope_id"] = ujconfig.Reference{
	//		Type: "github.com/crossplane-contrib/provider-keycloak/apis/openid/v1alpha1.ClientScope",
	//	}
	//})
	//p.AddResourceConfigurator("keycloak_generic_protocol_mapper", func(r *ujconfig.Resource) {
	//	r.ShortGroup = Group
	//	r.Name = "keycloak_generic_protocol_mapper"
	//	r.Kind = "SamlProtocolMapper"
	//	r.References["client_scope_id"] = ujconfig.Reference{
	//		Type: "github.com/crossplane-contrib/provider-keycloak/apis/saml/v1alpha1.ClientScope",
	//	}
	//})

	// Configure keycloak_generic_protocol_mapper for OpenIdProtocolMapper
	openIdConfigurator := ujconfig.ResourceConfiguratorFn(func(r *ujconfig.Resource) {
		r.ShortGroup = Group
		r.Name = "keycloak_generic_protocol_mapper"
		r.Kind = "OpenIdProtocolMapper"
		r.References["client_scope_id"] = ujconfig.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/openid/v1alpha1.ClientScope",
		}
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
	})
	// Configure keycloak_generic_protocol_mapper for SamlProtocolMapper
	samlConfigurator := ujconfig.ResourceConfiguratorFn(func(r *ujconfig.Resource) {
		r.ShortGroup = Group
		r.Name = "keycloak_generic_protocol_mapper"
		r.Kind = "SamlProtocolMapper"
		r.References["client_scope_id"] = ujconfig.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/saml/v1alpha1.ClientScope",
		}
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.SamlClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

	p.SetResourceConfigurator("keycloak_generic_protocol_mapper", ujconfig.ResourceConfiguratorChain{openIdConfigurator, samlConfigurator})

	// TODO: Add for SAML
	p.AddResourceConfigurator("keycloak_generic_role_mapper", func(r *ujconfig.Resource) {
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_saml_script_protocol_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.SamlClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_saml_user_attribute_protocol_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.SamlClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_saml_user_property_protocol_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.SamlClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

}
