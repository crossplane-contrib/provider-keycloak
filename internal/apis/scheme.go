// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package apis

import (
	xpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var s = runtime.NewScheme()

// GetManagedResource is Function to eliminate cross references using a transformer scheme
func GetManagedResource(group, version, kind, listKind string) (xpresource.Managed, xpresource.ManagedList, error) {

	// Define a function to get the managed resource based on group and version
	getResource := func(group, version, kind, listKind string) (xpresource.Managed, xpresource.ManagedList, error) {
		gv := schema.GroupVersion{
			Group:   group,
			Version: version,
		}
		kingGVK := gv.WithKind(kind)
		m, err := s.New(kingGVK)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get a new API object of GVK %q from the runtime scheme", kingGVK)
		}

		listGVK := gv.WithKind(listKind)
		l, err := s.New(listGVK)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get a new API object list of GVK %q from the runtime scheme", listGVK)
		}
		return m.(xpresource.Managed), l.(xpresource.ManagedList), nil
	}

	// Check for the special case input
	if group == "openidclient.keycloak.crossplane.io" && version == "v1alpha1" && kind == "Client" && listKind == "ClientList" {
		// Try the special case input first
		m, l, err := getResource(group, version, kind, listKind)

		if err == nil && m.GetName() != "" {
			return m, l, nil
		}
		// Fallback to the alternative input if the first attempt fails
		return getResource("samlclient.keycloak.crossplane.io", version, kind, listKind)
	}

	// For all other cases, use the provided input directly
	return getResource(group, version, kind, listKind)
}

// BuildScheme builds the runtime scheme for the Crossplane resources
func BuildScheme(sb runtime.SchemeBuilder) error {
	return errors.Wrap(sb.AddToScheme(s), "failed to register the GVKs with the runtime scheme")
}
