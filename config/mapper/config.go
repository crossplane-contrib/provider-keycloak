package mapper

import ujconfig "github.com/crossplane/upjet/pkg/config"

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
	})
	// Configure keycloak_generic_protocol_mapper for SamlProtocolMapper
	samlConfigurator := ujconfig.ResourceConfiguratorFn(func(r *ujconfig.Resource) {
		r.ShortGroup = Group
		r.Name = "keycloak_generic_protocol_mapper"
		r.Kind = "SamlProtocolMapper"
		r.References["client_scope_id"] = ujconfig.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/saml/v1alpha1.ClientScope",
		}
	})

	p.SetResourceConfigurator("keycloak_generic_protocol_mapper", ujconfig.ResourceConfiguratorChain{openIdConfigurator, samlConfigurator})
	p.AddResourceConfigurator("keycloak_generic_role_mapper", func(r *ujconfig.Resource) {
		r.ShortGroup = Group
	})

}
