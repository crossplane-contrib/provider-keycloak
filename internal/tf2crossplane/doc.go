/*
Copyright 2024 The Crossplane Authors.

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

// Package tf2crossplane converts Terraform HCL that uses the Keycloak
// Terraform provider into equivalent Crossplane Managed Resource manifests
// for provider-keycloak.
//
// The provider is generated with Upjet on top of the Keycloak Terraform
// provider, which means every CRD's spec.forProvider maps 1:1 onto the
// Terraform resource arguments. The converter leverages that by loading the
// exact same provider configuration the code generator uses
// (config.GetProvider) as the single source of truth for:
//
//   - the Terraform resource name -> CRD GroupVersionKind mapping,
//   - the snake_case -> camelCase field name transformation, and
//   - the cross-resource reference wiring (e.g. realm_id -> realmIdRef).
package tf2crossplane
