/*
Copyright 2020 The cert-manager authors.

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

package main

import (
	"flag"
	"os"
	"time"

	capi "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports

	"github.com/cert-manager/signer-ca/controllers"
	"github.com/cert-manager/signer-ca/internal/kubernetes/signer"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = capi.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var leaderElectionID string
	var signerName string
	var caCertPath string
	var caKeyPath string
	var certificateDuration time.Duration
	var debugLogging bool

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&leaderElectionID, "leader-election-id", "signer-ca-leader-election",
		"The name of the configmap used to coordinate leader election between controller-managers.")
	flag.StringVar(&signerName, "signer-name", "example.com/foo", "Only sign CSR with this .spec.signerName.")
	flag.StringVar(&caCertPath, "ca-cert-path", "/etc/pki/example.com/foo/ca.pem", "Sign CSR with this certificate file.")
	flag.StringVar(&caKeyPath, "ca-key-path", "/etc/pki/example.com/foo/ca-key.pem", "Sign CSR with this private key file.")
	flag.DurationVar(&certificateDuration, "certificate-duration", time.Hour, "The duration of the signed certificates.")
	flag.BoolVar(&debugLogging, "debug-logging", true, "Enable debug logging.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(debugLogging)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   leaderElectionID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	signer, err := signer.NewSigner(caCertPath, caKeyPath, certificateDuration)
	if err != nil {
		setupLog.Error(err, "unable to start signer")
		os.Exit(1)
	}
	if err := (&controllers.CertificateSigningRequestSigningReconciler{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName("CSRSigningReconciler"),
		Scheme:        mgr.GetScheme(),
		SignerName:    signerName,
		Signer:        signer,
		EventRecorder: mgr.GetEventRecorderFor("CSRSigningReconciler"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Controller", "controller", "CSRSigningReconciler")
		os.Exit(1)

	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
