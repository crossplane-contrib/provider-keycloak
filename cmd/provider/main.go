/*
Copyright 2021 Upbound Inc.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kingpin/v2"
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/gate"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/customresourcesgate"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	tjcontroller "github.com/crossplane/upjet/v2/pkg/controller"
	authv1 "k8s.io/api/authorization/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	apisCluster "github.com/crossplane-contrib/provider-keycloak/apis/cluster"
	apisNamespaced "github.com/crossplane-contrib/provider-keycloak/apis/namespaced"
	resolverapis "github.com/crossplane-contrib/provider-keycloak/internal/apis"
	"github.com/crossplane-contrib/provider-keycloak/config"
	// resolverapis "github.com/crossplane-contrib/provider-keycloak/internal/apis"
	"github.com/crossplane-contrib/provider-keycloak/internal/clients"
	controllerCluster "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster"
	controllerNamespaced "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced"
	"github.com/crossplane-contrib/provider-keycloak/internal/features"
	// "github.com/crossplane-contrib/provider-keycloak/internal/version"
)

const (
	webhookTLSCertDirEnvVar = "WEBHOOK_TLS_CERT_DIR"
	tlsServerCertDirEnvVar  = "TLS_SERVER_CERTS_DIR"
	certsDirEnvVar          = "CERTS_DIR"
	tlsServerCertDir        = "/tls/server"
)

func deprecationAction(flagName string) kingpin.Action {
	return func(c *kingpin.ParseContext) error {
		_, err := fmt.Fprintf(os.Stderr, "warning: Command-line flag %q is deprecated and no longer used. It will be removed in a future release. Please remove it from all of your configurations (ControllerConfigs, etc.).\n", flagName)
		kingpin.FatalIfError(err, "Failed to print the deprecation notice.")
		return nil
	}
}

func main() {
	var (
		app                     = kingpin.New(filepath.Base(os.Args[0]), "Terraform based Crossplane provider for Vault").DefaultEnvars()
		debug                   = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		syncPeriod              = app.Flag("sync", "Controller manager sync period such as 300ms, 1.5h, or 2h45m").Short('s').Default("1h").Duration()
		pollInterval            = app.Flag("poll", "Poll interval controls how often an individual resource should be checked for drift.").Default("10m").Duration()
		pollStateMetricInterval = app.Flag("poll-state-metric", "State metric recording interval").Default("5s").Duration()
		leaderElection          = app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").OverrideDefaultFromEnvar("LEADER_ELECTION").Bool()
		maxReconcileRate        = app.Flag("max-reconcile-rate", "The global maximum rate per second at which resources may checked for drift from the desired state.").Default("10").Int()
		webhookPort             = app.Flag("webhook-port", "The port the webhook listens on").Default("9443").Envar("WEBHOOK_PORT").Int()
		metricsBindAddress      = app.Flag("metrics-bind-address", "The address the metrics server listens on").Default(":8080").Envar("METRICS_BIND_ADDRESS").String()

		enableManagementPolicies = app.Flag("enable-management-policies", "Enable support for Management Policies.").Default("true").Envar("ENABLE_MANAGEMENT_POLICIES").Bool()

		certsDirSet = false
		// we record whether the command-line option "--certs-dir" was supplied
		// in the registered PreAction for the flag.
		certsDir = app.Flag("certs-dir", "The directory that contains the server key and certificate.").Default(tlsServerCertDir).Envar(certsDirEnvVar).PreAction(func(_ *kingpin.ParseContext) error {
			certsDirSet = true
			return nil
		}).String()

		// deprecated command-line arguments with the ESS support removal
		_ = app.Flag("namespace", "Namespace used to set as default scope in default secret store config.").Default("upbound-system").Envar("POD_NAMESPACE").Hidden().Action(deprecationAction("namespace")).String()
		_ = app.Flag("enable-external-secret-stores", "Enable support for ExternalSecretStores.").Default("false").Envar("ENABLE_EXTERNAL_SECRET_STORES").Hidden().Action(deprecationAction("enable-external-secret-stores")).Bool()
		_ = app.Flag("ess-tls-cert-dir", "Path of ESS TLS certificates.").Envar("ESS_TLS_CERTS_DIR").Hidden().Action(deprecationAction("ess-tls-cert-dir")).String()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-vault"))
	if *debug {
		// The controller-runtime runs with a no-op logger by default. It is
		// *very* verbose even at info level, so we only provide it a real
		// logger when we're running in debug mode.
		ctrl.SetLogger(zl)
	}

	log.Debug("Starting", "sync-period", syncPeriod.String(), "poll-interval", pollInterval.String(), "max-reconcile-rate", *maxReconcileRate)

	// currently, we configure the jitter to be the 5% of the poll interval
	pollJitter := time.Duration(float64(*pollInterval) * 0.05)
	log.Debug("Starting", "sync-interval", syncPeriod.String(),
		"poll-interval", pollInterval.String(), "poll-jitter", pollJitter, "max-reconcile-rate", *maxReconcileRate)

	cfg, err := ctrl.GetConfig()
	kingpin.FatalIfError(err, "Cannot get API server rest config")

	// Get the TLS certs directory from the environment variables set by
	// Crossplane if they're available.
	// In older XP versions we used WEBHOOK_TLS_CERT_DIR, in newer versions
	// we use TLS_SERVER_CERTS_DIR. If an explicit certs dir is not supplied
	// via the command-line options, then these environment variables are used
	// instead.
	if !certsDirSet {
		// backwards-compatibility concerns
		xpCertsDir := os.Getenv(certsDirEnvVar)
		if xpCertsDir == "" {
			xpCertsDir = os.Getenv(tlsServerCertDirEnvVar)
		}
		if xpCertsDir == "" {
			xpCertsDir = os.Getenv(webhookTLSCertDirEnvVar)
		}
		// we probably don't need this condition but just to be on the
		// safe side, if we are missing any kingpin machinery details...
		if xpCertsDir != "" {
			*certsDir = xpCertsDir
		}
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		LeaderElection:   *leaderElection,
		LeaderElectionID: "crossplane-leader-election-provider-keycloak",
		Cache: cache.Options{
			SyncPeriod: syncPeriod,
		},
		Metrics: metricsserver.Options{
			BindAddress: *metricsBindAddress,
		},
		WebhookServer: webhook.NewServer(
			webhook.Options{
				CertDir: *certsDir,
				Port:    *webhookPort,
			}),
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
		LeaseDuration:              func() *time.Duration { d := 60 * time.Second; return &d }(),
		RenewDeadline:              func() *time.Duration { d := 50 * time.Second; return &d }(),
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")
	kingpin.FatalIfError(apisCluster.AddToScheme(mgr.GetScheme()), "Cannot add Keycloak APIs to scheme")
	kingpin.FatalIfError(resolverapis.BuildScheme(apisCluster.AddToSchemes), "Cannot register cluster scoped APIs with the API resolver's runtime scheme")

	kingpin.FatalIfError(apisNamespaced.AddToScheme(mgr.GetScheme()), "Cannot add Keycloak APIs to scheme")
	kingpin.FatalIfError(resolverapis.BuildScheme(apisNamespaced.AddToSchemes), "Cannot register namespace scoped APIs with the API resolver's runtime scheme")
	kingpin.FatalIfError(apiextensionsv1.AddToScheme(mgr.GetScheme()), "Cannot add api-extensions APIs to scheme")

	metricRecorder := managed.NewMRMetricRecorder()
	stateMetrics := statemetrics.NewMRStateMetrics()

	metrics.Registry.MustRegister(metricRecorder)
	metrics.Registry.MustRegister(stateMetrics)

	providerCluster, err := config.GetProvider(false)
	kingpin.FatalIfError(err, "Cannot initialize the cluster provider configuration")
	providerNamespaced, err := config.GetProviderNamespaced(false)
	kingpin.FatalIfError(err, "Cannot initialize the namespaced provider configuration")
	featureFlags := &feature.Flags{}
	optsCluster := tjcontroller.Options{
		Options: xpcontroller.Options{
			Logger:                  log,
			GlobalRateLimiter:       ratelimiter.NewGlobal(*maxReconcileRate),
			PollInterval:            *pollInterval,
			MaxConcurrentReconciles: *maxReconcileRate,
			Features:                featureFlags,
			MetricOptions: &xpcontroller.MetricOptions{
				PollStateMetricInterval: *pollStateMetricInterval,
				MRMetrics:               metricRecorder,
				MRStateMetrics:          stateMetrics,
			},
		},
		Provider:              providerCluster,
		SetupFn:               clients.TerraformSetupBuilder(),
		PollJitter:            pollJitter,
		OperationTrackerStore: tjcontroller.NewOperationStore(log),
	}
	optsNamespaced := tjcontroller.Options{
		Options: xpcontroller.Options{
			Logger:                  log,
			GlobalRateLimiter:       ratelimiter.NewGlobal(*maxReconcileRate),
			PollInterval:            *pollInterval,
			MaxConcurrentReconciles: *maxReconcileRate,
			Features:                featureFlags,
			MetricOptions: &xpcontroller.MetricOptions{
				PollStateMetricInterval: *pollStateMetricInterval,
				MRMetrics:               metricRecorder,
				MRStateMetrics:          stateMetrics,
			},
		},
		Provider:              providerNamespaced,
		SetupFn:               clients.TerraformSetupBuilder(),
		PollJitter:            pollJitter,
		OperationTrackerStore: tjcontroller.NewOperationStore(log),
	}

	if *enableManagementPolicies {
		optsCluster.Features.Enable(features.EnableBetaManagementPolicies)
		optsNamespaced.Features.Enable(features.EnableBetaManagementPolicies)
		log.Info("Beta feature enabled", "flag", features.EnableBetaManagementPolicies)
	}

	canSafeStart, err := canWatchCRD(mgr)
	kingpin.FatalIfError(err, "SafeStart precheck failed")
	if canSafeStart {
		crdGate := new(gate.Gate[schema.GroupVersionKind])
		optsCluster.Gate = crdGate
		optsNamespaced.Gate = crdGate
		kingpin.FatalIfError(customresourcesgate.Setup(mgr, optsNamespaced.Options), "Cannot setup CRD gate")
		kingpin.FatalIfError(controllerCluster.SetupGated(mgr, optsCluster), "Cannot setup Vault controllers")
		kingpin.FatalIfError(controllerNamespaced.SetupGated(mgr, optsNamespaced), "Cannot setup Vault controllers")
	} else {
		log.Info("Provider has missing RBAC permissions for watching CRDs, controller SafeStart capability will be disabled")
		kingpin.FatalIfError(controllerCluster.Setup(mgr, optsCluster), "Cannot setup Vault controllers")
		kingpin.FatalIfError(controllerNamespaced.Setup(mgr, optsNamespaced), "Cannot setup Vault controllers")
	}

	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}

func canWatchCRD(mgr manager.Manager) (bool, error) {
	ctx := context.Background()
	if err := authv1.AddToScheme(mgr.GetScheme()); err != nil {
		return false, err
	}
	verbs := []string{"get", "list", "watch"}
	for _, verb := range verbs {
		sar := &authv1.SelfSubjectAccessReview{
			Spec: authv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Group:    "apiextensions.k8s.io",
					Resource: "customresourcedefinitions",
					Verb:     verb,
				},
			},
		}
		if err := mgr.GetClient().Create(ctx, sar); err != nil {
			return false, errors.Wrapf(err, "unable to perform RBAC check for verb %s on CustomResourceDefinitions", verbs)
		}
		if !sar.Status.Allowed {
			return false, nil
		}
	}
	return true, nil
}
