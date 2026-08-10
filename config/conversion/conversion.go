// Package conversion contains helpers to manage breaking CRD schema changes by
// bumping the API version of a resource and registering the conversion
// functions that the provider's conversion webhook uses to translate objects
// between the old (spoke) and the new (hub) API versions.
package conversion

import (
	"fmt"

	"github.com/crossplane/upjet/v2/pkg/config"
	ujconversion "github.com/crossplane/upjet/v2/pkg/config/conversion"
)

// BumpVersionForIntToStringChange freezes the current API version of the given
// resource as a previous (spoke) version, promotes newVersion to the current
// (hub and storage) version and registers bidirectional conversion functions
// for the given fields, whose type changed from a number to a string.
//
// fieldPaths must be given relative to spec.forProvider, e.g.
// "clientSecretWoVersion". The conversions are registered for
// spec.forProvider, spec.initProvider and status.atProvider.
//
// The number values are converted with the IntToString/StringToInt modes rather
// than FloatToString/StringToFloat: the CRD schema type of the old version is
// `number`, but the values are integers and strconv.FormatFloat would render
// them in scientific notation (202605192 -> "2.02605192e+08").
func BumpVersionForIntToStringChange(r *config.Resource, newVersion string, fieldPaths ...string) {
	previousVersion := r.Version
	r.Version = newVersion
	r.PreviousVersions = append(r.PreviousVersions, previousVersion)
	r.SetCRDStorageVersion(newVersion)
	r.SetCRDHubVersion(newVersion)

	// The default identity conversion copies every field as-is, which fails for
	// fields whose type changed. Replace it with one that skips those fields;
	// they are handled by the type conversions registered below.
	identity := ujconversion.NewIdentityConversionExpandPaths(ujconversion.AllVersions, ujconversion.AllVersions, ujconversion.DefaultPathPrefixes(), fieldPaths...)
	if len(r.Conversions) > 0 {
		r.Conversions[0] = identity
	} else {
		r.Conversions = []ujconversion.Conversion{identity}
	}

	for _, f := range fieldPaths {
		for _, prefix := range ujconversion.DefaultPathPrefixes() {
			p := fmt.Sprintf("%s.%s", prefix, f)
			r.Conversions = append(r.Conversions,
				ujconversion.NewFieldTypeConversion(previousVersion, newVersion, p, ujconversion.IntToString),
				ujconversion.NewFieldTypeConversion(newVersion, previousVersion, p, ujconversion.StringToInt),
			)
		}
	}
}
