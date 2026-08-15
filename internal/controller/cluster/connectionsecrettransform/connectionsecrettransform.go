/*
Copyright 2022 Upbound Inc.
*/

// Package connectionsecrettransform implements a hand-written (i.e. not
// upjet-generated) controller that republishes the connection secret of a
// keycloak_openid_client (Client) managed resource with renamed keys, as
// configured on the ProviderConfig the resource uses.
//
// Background: connection secret keys are published today via
// config.Resource.Sensitive.AdditionalConnectionDetailsFn (see
// config/openidclient/config.go), which upjet invokes with only the
// Terraform resource's own state - it has no access to the resource's
// resource.Managed object or its ProviderConfig (see
// github.com/crossplane/upjet/v2 v2.3.0, pkg/resource/sensitive.go:
// GetConnectionDetails). That makes it impossible to rename keys, or add
// ProviderConfig-derived values, from inside that hook on a per-ProviderConfig
// basis. This controller sidesteps that limitation entirely by operating
// downstream, on already-materialized Kubernetes objects (the Secret written
// by the Client's own controller, the owning Client, and its ProviderConfig),
// all of which a normal controller-runtime client can read directly.
package connectionsecrettransform

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/crossplane/upjet/v2/pkg/controller"

	clientv1alpha2 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha2"
	clusterv1beta1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/v1beta1"
)

const (
	// transformedSecretSuffix is appended to the name of the source
	// connection secret to build the name of the republished, transformed
	// one. The two live side by side: the original is left untouched so
	// that upjet's tfstate-rebuild-from-connection-secret machinery keeps
	// working, and consumers that need renamed keys can point at the
	// "-transformed" secret instead.
	transformedSecretSuffix = "-transformed"

	// transformedSecretLabel marks secrets written by this controller, so
	// that it can ignore its own output when reconciling Secret events
	// instead of trying to transform an already-transformed secret.
	transformedSecretLabel = "keycloak.crossplane.io/connection-secret-transform"

	clientKind = "Client"
)

// Reconciler republishes a Client's connection secret with keys renamed
// according to the spec.connectionSecretKeys.rename map configured on the
// Client's ProviderConfig.
type Reconciler struct {
	client client.Client
}

// Setup adds a controller that watches connection Secrets owned by
// openidclient.Client managed resources and republishes them with renamed
// keys, as configured on the resource's ProviderConfig.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := "connectionsecrettransform/" + clientv1alpha2.CRDGroup

	r := &Reconciler{client: mgr.GetClient()}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&corev1.Secret{}).
		Complete(r)
}

// SetupGated adds this controller directly, bypassing the CRD-establishment
// gate other controllers use. Unlike upjet-generated controllers, this one
// does not own a CRD of its own: it only needs the Secret, Client and
// ProviderConfig kinds it reads to already be registered with the client's
// scheme, which is guaranteed at manager start-up regardless of CRD
// establishment order.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	return Setup(mgr, o)
}

// Reconcile implements reconcile.Reconciler. It is a no-op for any Secret
// that is not the connection secret of a Client managed resource whose
// ProviderConfig configures connectionSecretKeys.rename.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	secret := &corev1.Secret{}
	if err := r.client.Get(ctx, req.NamespacedName, secret); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Never try to transform our own output.
	if secret.Labels[transformedSecretLabel] == "true" {
		return reconcile.Result{}, nil
	}

	owner := findClientOwner(secret.OwnerReferences)
	if owner == nil {
		return reconcile.Result{}, nil
	}

	mr := &clientv1alpha2.Client{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: owner.Name}, mr); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if mr.Spec.ProviderConfigReference == nil {
		return reconcile.Result{}, nil
	}

	pc := &clusterv1beta1.ProviderConfig{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: mr.Spec.ProviderConfigReference.Name}, pc); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if pc.Spec.ConnectionSecretKeys == nil || len(pc.Spec.ConnectionSecretKeys.Rename) == 0 {
		return reconcile.Result{}, nil
	}

	data := make(map[string][]byte, len(secret.Data))
	for k, v := range secret.Data {
		newKey := k
		if renamed, ok := pc.Spec.ConnectionSecretKeys.Rename[k]; ok && renamed != "" {
			newKey = renamed
		}
		data[newKey] = v
	}

	out := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name + transformedSecretSuffix,
			Namespace: secret.Namespace,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, out, func() error {
		if out.Labels == nil {
			out.Labels = map[string]string{}
		}
		out.Labels[transformedSecretLabel] = "true"
		out.Type = secret.Type
		out.Data = data
		// Tie the transformed secret's lifecycle to the source secret so
		// it is garbage-collected together with it (and, transitively,
		// with the owning Client).
		return controllerutil.SetOwnerReference(secret, out, r.client.Scheme())
	})
	return reconcile.Result{}, err
}

func findClientOwner(refs []metav1.OwnerReference) *metav1.OwnerReference {
	for i := range refs {
		if refs[i].Kind != clientKind {
			continue
		}
		if !strings.HasPrefix(refs[i].APIVersion, clientv1alpha2.CRDGroup+"/") {
			continue
		}
		return &refs[i]
	}
	return nil
}
