/*
Copyright 2026 Data Foundation.

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

package controller

import (
	"context"
	"maps"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ensureNamespaces(
	ctx context.Context, cli client.Client, logger logr.Logger,
	olmPkgRecords []*OlmPkgRecord, operatorNamespace, publisherName string) error {

	for _, record := range olmPkgRecords {

		if record.Namespace != operatorNamespace && publisherName == PublisherNameIBM {
			return createOrUpdateNamespace(ctx, cli, logger, record.Namespace)
		}
	}

	return nil
}

func createOrUpdateNamespace(ctx context.Context, cli client.Client, logger logr.Logger, namespace string) error {

	desiredNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				ManagedByKey: ManagedByValOdfOperator,
			},
		},
	}

	ns := &corev1.Namespace{}
	ns.Name = namespace

	op, err := controllerutil.CreateOrUpdate(ctx, cli, ns, func() error {

		if ns.Annotations == nil {
			ns.Annotations = map[string]string{}
		}

		if ns.Labels == nil {
			ns.Labels = map[string]string{}
		}

		maps.Copy(ns.Annotations, desiredNs.Annotations)
		maps.Copy(ns.Labels, desiredNs.Labels)
		return nil
	})

	if err != nil {
		logger.Error(err, "failed to create or update Namespace", "Namespace", namespace)
		return err
	}

	logger.Info("Namespace reconciled successfully", "Namespace", namespace, "operation", op)

	return nil
}
