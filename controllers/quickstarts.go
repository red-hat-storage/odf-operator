package controllers

import (
	"context"

	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	consolev1 "github.com/openshift/api/console/v1"
)

func (r *StorageSystemReconciler) ensureQuickStarts(logger logr.Logger) error {
	qscrd := extv1.CustomResourceDefinition{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: "consolequickstarts.console.openshift.io", Namespace: ""}, &qscrd)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(2).Info("No custom resource definition found for consolequickstart. Skipping quickstart initialization")
			return nil
		}
		return err
	}
	if len(AllQuickStarts) == 0 {
		logger.Info("No quickstarts found")
		return nil
	}
	for _, qs := range AllQuickStarts {
		cqs := consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, &cqs)
		if err != nil {
			logger.Error(err, "Failed to unmarshal ConsoleQuickStart", "ConsoleQuickStartString", string(qs))
			continue
		}
		found := consolev1.ConsoleQuickStart{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: cqs.Name, Namespace: cqs.Namespace}, &found)
		if err != nil {
			if errors.IsNotFound(err) {
				err = r.Client.Create(context.TODO(), &cqs)
				if err != nil {
					logger.Error(err, "Failed to create quickstart", "Name", cqs.Name, "Namespace", cqs.Namespace)
					return nil
				}
				logger.Info("Creating quickstarts", "Name", cqs.Name, "Namespace", cqs.Namespace)
				continue
			}
			logger.Error(err, "Error has occurred when fetching quickstarts")
			return nil
		}
		found.Spec = cqs.Spec
		err = r.Client.Update(context.TODO(), &found)
		if err != nil {
			logger.Error(err, "Failed to update quickstart", "Name", cqs.Name, "Namespace", cqs.Namespace)
			return nil
		}
		logger.Info("Updating quickstarts", "Name", cqs.Name, "Namespace", cqs.Namespace)
	}
	return nil
}
