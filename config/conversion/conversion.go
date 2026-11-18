// Package conversion contains helpers to manage breaking CRD schema changes by
// bumping the API version of a resource and registering the conversion
// functions that the provider's conversion webhook uses to translate objects
// between the old (spoke) and the new (hub) API versions.
package conversion

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	ujconversion "github.com/crossplane/upjet/v2/pkg/config/conversion"
)

const alphaVersionPrefix = "v1alpha"

// FieldTypeChange describes a field whose type changed between API versions.
type FieldTypeChange struct {
	SourceVersion string
	FieldPath     string
}

// BumpVersionsForSingletonListEmbedding freezes the current API version of
// every resource with singleton-list fields before those fields are embedded as
// objects in the next API version.
func BumpVersionsForSingletonListEmbedding(p *config.Provider, typeChanges map[string][]FieldTypeChange) error {
	for name, r := range p.Resources {
		listPaths := r.CRDListConversionPaths()
		if len(listPaths) == 0 {
			continue
		}
		r.TerraformConversions = append(r.TerraformConversions, config.NewTFSingletonConversion())
		nextVersion, err := nextAlphaVersion(r.Version)
		if err != nil {
			return fmt.Errorf("cannot bump API version for resource %q: %w", name, err)
		}
		previousVersion := r.Version
		bumpVersion(r, nextVersion)
		excludedPaths := append([]string(nil), listPaths...)
		for _, change := range typeChanges[name] {
			excludedPaths = append(excludedPaths, change.FieldPath)
		}
		identity := ujconversion.NewIdentityConversionExpandPaths(
			ujconversion.AllVersions,
			ujconversion.AllVersions,
			ujconversion.DefaultPathPrefixes(),
			excludedPaths...,
		)
		if len(r.Conversions) == 0 {
			r.Conversions = []ujconversion.Conversion{identity}
		} else {
			r.Conversions[0] = identity
		}

		for _, oldVersion := range r.PreviousVersions {
			r.Conversions = append(r.Conversions,
				ujconversion.NewSingletonListConversion(
					oldVersion,
					nextVersion,
					ujconversion.DefaultPathPrefixes(),
					listPaths,
					ujconversion.ToEmbeddedObject,
				),
				ujconversion.NewSingletonListConversion(
					nextVersion,
					oldVersion,
					ujconversion.DefaultPathPrefixes(),
					listPaths,
					ujconversion.ToSingletonList,
				),
			)
		}

		for _, change := range typeChanges[name] {
			for _, prefix := range ujconversion.DefaultPathPrefixes() {
				path := fmt.Sprintf("%s.%s", prefix, change.FieldPath)
				r.Conversions = append(r.Conversions,
					ujconversion.NewFieldTypeConversion(change.SourceVersion, nextVersion, path, ujconversion.IntToString),
					ujconversion.NewFieldTypeConversion(nextVersion, change.SourceVersion, path, ujconversion.StringToInt),
					ujconversion.NewFieldRenameConversion(previousVersion, path, nextVersion, path),
					ujconversion.NewFieldRenameConversion(nextVersion, path, previousVersion, path),
				)
			}
		}
	}
	return nil
}

func nextAlphaVersion(current string) (string, error) {
	if !strings.HasPrefix(current, alphaVersionPrefix) {
		return "", fmt.Errorf("unsupported API version %q", current)
	}
	n, err := strconv.Atoi(strings.TrimPrefix(current, alphaVersionPrefix))
	if err != nil {
		return "", fmt.Errorf("unsupported API version %q", current)
	}
	return fmt.Sprintf("%s%d", alphaVersionPrefix, n+1), nil
}

func bumpVersion(r *config.Resource, nextVersion string) {
	r.PreviousVersions = append(r.PreviousVersions, r.Version)
	r.Version = nextVersion
	r.SetCRDStorageVersion(nextVersion)
	r.SetCRDHubVersion(nextVersion)
}

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
	bumpVersion(r, newVersion)

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
