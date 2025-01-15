package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/red-hat-storage/odf-operator/controllers"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ClusterServiceVersionDefaulter struct {
	client.Client

	Decoder           admission.Decoder
	OperatorNamespace string

	ctx context.Context
	log logr.Logger
}

// +kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions,verbs=create;update;patch

func (r *ClusterServiceVersionDefaulter) Handle(ctx context.Context, req admission.Request) admission.Response {

	r.ctx = ctx
	r.log = log.FromContext(ctx)
	r.log.Info("request received for csv review")

	instance := &opv1a1.ClusterServiceVersion{}
	err := r.Decoder.Decode(req, instance)
	if err != nil {
		r.log.Error(err, "failed decoding admission review as csv")
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed decoding admission review as csv: %v", err))
	}

	r.mutateCsv(instance)

	marshaledInstance, err := json.Marshal(instance)
	if err != nil {
		r.log.Error(err, "failed marshaling csv")
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed marshaling csv: %v", err))
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledInstance)
}

func (r *ClusterServiceVersionDefaulter) mutateCsv(instance *opv1a1.ClusterServiceVersion) {

	ok := r.isOdfManagedCSV(instance)
	if !ok {
		r.log.Info("ignoring requested csv as it is not relevant")
		return
	}

	r.log.Info("mutating requested csv")
	for i := range instance.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		instance.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec.Replicas = ptr.To(int32(0))
	}
}

func (r *ClusterServiceVersionDefaulter) isOdfManagedCSV(instance *opv1a1.ClusterServiceVersion) bool {

	if instance.Namespace != r.OperatorNamespace {
		return false
	}

	for _, pkgName := range controllers.PkgNames {
		if strings.HasPrefix(instance.ObjectMeta.Name, pkgName) {
			return true
		}
	}

	return false
}

func (r *ClusterServiceVersionDefaulter) SetupWebhookWithManager(mgr ctrl.Manager) error {

	mgr.GetWebhookServer().Register(controllers.WebhookPath, &webhook.Admission{Handler: r})

	return nil
}
