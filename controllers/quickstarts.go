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
	"context"

	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	consolev1 "github.com/openshift/api/console/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ensureQuickStarts(ctx context.Context, cli client.Client, log logr.Logger) error {
	if len(AllQuickStarts) == 0 {
		log.Info("No quickstarts found")
		return nil
	}
	for _, qs := range AllQuickStarts {
		cqs := consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, &cqs)
		if err != nil {
			log.Error(err, "Failed to unmarshal ConsoleQuickStart", "ConsoleQuickStartString", string(qs))
			continue
		}
		found := consolev1.ConsoleQuickStart{}
		err = cli.Get(ctx, types.NamespacedName{Name: cqs.Name, Namespace: cqs.Namespace}, &found)
		if err != nil {
			if errors.IsNotFound(err) {
				err = cli.Create(ctx, &cqs)
				if err != nil {
					log.Error(err, "Failed to create quickstart", "Name", cqs.Name, "Namespace", cqs.Namespace)
					return nil
				}
				log.Info("Creating quickstarts", "Name", cqs.Name, "Namespace", cqs.Namespace)
				continue
			}
			log.Error(err, "Error has occurred when fetching quickstarts")
			return nil
		}
		found.Spec = cqs.Spec
		err = cli.Update(ctx, &found)
		if err != nil {
			log.Error(err, "Failed to update quickstart", "Name", cqs.Name, "Namespace", cqs.Namespace)
			return nil
		}
		log.Info("Updating quickstarts", "Name", cqs.Name, "Namespace", cqs.Namespace)
	}
	return nil
}

func deleteQuickStarts(ctx context.Context, cli client.Client, log logr.Logger) { //nolint:unused
	if len(AllQuickStarts) == 0 {
		log.Info("No quickstarts found.")
	}

	for _, qs := range AllQuickStarts {
		cqs := consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, &cqs)
		if err != nil {
			log.Error(err, "Failed to unmarshal ConsoleQuickStart.", "ConsoleQuickStartString", string(qs))
			continue
		}

		err = cli.Delete(ctx, &cqs)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			log.Error(err, "Failed to delete quickstart", "Name", cqs.Name, "Namespace", cqs.Namespace)
		}

		log.Info("Quickstart marked for deletion", "Name", cqs.Name, "Namespace", cqs.Namespace)
	}
}
