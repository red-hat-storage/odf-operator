/*
Copyright 2021 Red Hat OpenShift Data Foundation.

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
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	consolev1 "github.com/openshift/api/console/v1"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
)

func (r *StorageSystemReconciler) ensureQuickStarts(logger logr.Logger) error {
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
		err = r.Client.Get(r.ctx, types.NamespacedName{Name: cqs.Name, Namespace: cqs.Namespace}, &found)
		if err != nil {
			if errors.IsNotFound(err) {
				err = r.Client.Create(r.ctx, &cqs)
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
		err = r.Client.Update(r.ctx, &found)
		if err != nil {
			logger.Error(err, "Failed to update quickstart", "Name", cqs.Name, "Namespace", cqs.Namespace)
			return nil
		}
		logger.Info("Updating quickstarts", "Name", cqs.Name, "Namespace", cqs.Namespace)
	}
	return nil
}

func (r *StorageSystemReconciler) deleteQuickStarts(logger logr.Logger, instance *odfv1alpha1.StorageSystem) {
	if len(AllQuickStarts) == 0 {
		logger.Info("No quickstarts found.")
	}

	allSSDeleted, err := r.areAllStorageSystemsMarkedForDeletion(instance.Namespace)
	if err != nil {
		// Log the error but not fail the operator
		logger.Error(err, "Failed to List", "Kind", "StorageSystem")
		return
	}

	if !allSSDeleted {
		return
	}

	for _, qs := range AllQuickStarts {
		cqs := consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, &cqs)
		if err != nil {
			logger.Error(err, "Failed to unmarshal ConsoleQuickStart.", "ConsoleQuickStartString", string(qs))
			continue
		}

		err = r.Client.Delete(r.ctx, &cqs)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			logger.Error(err, "Failed to delete quickstart", "Name", cqs.Name, "Namespace", cqs.Namespace)
		}

		logger.Info("Quickstart marked for deletion", "Name", cqs.Name, "Namespace", cqs.Namespace)
	}
}
