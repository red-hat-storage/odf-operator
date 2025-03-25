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

	"github.com/go-logr/logr"
	consolev1 "github.com/openshift/api/console/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

// ensureQuickStarts create or update the quickstarts
func ensureQuickStarts(ctx context.Context, cli client.Client, logger logr.Logger) error {
	for _, qs := range AllQuickStarts {
		desiredCQS := &consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, desiredCQS)
		if err != nil {
			logger.Error(err, "failed to unmarshal ConsoleQuickStart", "ConsoleQuickStartString", string(qs))
			continue
		}
		cqs := &consolev1.ConsoleQuickStart{}
		cqs.ObjectMeta = desiredCQS.ObjectMeta
		_, err = controllerutil.CreateOrUpdate(ctx, cli, cqs, func() error {
			cqs.Spec = desiredCQS.Spec
			return nil
		})
		if err != nil {
			logger.Error(err, "failed to create or update quickstart", "Name", desiredCQS.Name, "Namespace", desiredCQS.Namespace)
			return nil
		}
		logger.Info("updating quickstarts", "Name", desiredCQS.Name, "Namespace", desiredCQS.Namespace)
	}
	return nil
}

// deleteQuickStarts deletes the quickstarts
// TODO: This function is not used, a call for this function need to be introduced whenever we resolve ODF uninstallation techdebt
func deleteQuickStarts(ctx context.Context, cli client.Client, logger logr.Logger) { //nolint:unused

	for _, qs := range AllQuickStarts {
		cqs := consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, &cqs)
		if err != nil {
			logger.Error(err, "failed to unmarshal ConsoleQuickStart.", "ConsoleQuickStartString", string(qs))
			continue
		}

		err = cli.Delete(ctx, &cqs)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			logger.Error(err, "failed to delete quickstart", "Name", cqs.Name, "Namespace", cqs.Namespace)
		}

		logger.Info("quickstart marked for deletion", "Name", cqs.Name, "Namespace", cqs.Namespace)
	}
}
