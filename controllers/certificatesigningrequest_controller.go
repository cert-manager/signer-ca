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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	capi "k8s.io/api/certificates/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capihelper "github.com/cert-manager/signer-ca/internal/api"
	"github.com/cert-manager/signer-ca/internal/kubernetes/signer"
)

// CertificateSigningRequestSigningReconciler reconciles a CertificateSigningRequest object
type CertificateSigningRequestSigningReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	SignerName    string
	Signer        *signer.Signer
	EventRecorder record.EventRecorder
}

// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/status,verbs=patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *CertificateSigningRequestSigningReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("certificatesigningrequest", req.NamespacedName)
	var csr capi.CertificateSigningRequest
	if err := r.Client.Get(ctx, req.NamespacedName, &csr); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("error %q getting CSR", err)
	}
	switch {
	case !csr.DeletionTimestamp.IsZero():
		log.V(1).Info("CSR has been deleted. Ignoring.")
	case csr.Spec.SignerName == nil:
		log.V(1).Info("CSR does not have a signer name. Ignoring.")
	case *csr.Spec.SignerName != r.SignerName:
		log.V(1).Info("CSR signer name does not match. Ignoring.", "signer-name", csr.Spec.SignerName)
	case csr.Status.Certificate != nil:
		log.V(1).Info("CSR has already been signed. Ignoring.")
	case !capihelper.IsCertificateRequestApproved(&csr):
		log.V(1).Info("CSR is not approved, Ignoring.")
	default:
		log.V(1).Info("Signing")
		x509cr, err := capihelper.ParseCSR(csr.Spec.Request)
		if err != nil {
			log.Error(err, "unable to parse csr")
			r.EventRecorder.Event(&csr, v1.EventTypeWarning, "SigningFailed", "Unable to parse the CSR request")
			return ctrl.Result{}, nil
		}
		cert, err := r.Signer.Sign(x509cr, csr.Spec.Usages)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error auto signing csr: %v", err)
		}
		patch := client.MergeFrom(csr.DeepCopy())
		csr.Status.Certificate = cert
		if err := r.Client.Status().Patch(ctx, &csr, patch); err != nil {
			return ctrl.Result{}, fmt.Errorf("error patching CSR: %v", err)
		}
		r.EventRecorder.Event(&csr, v1.EventTypeNormal, "Signed", "The CSR has been signed")
	}
	return ctrl.Result{}, nil
}

func (r *CertificateSigningRequestSigningReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capi.CertificateSigningRequest{}).
		Complete(r)
}
