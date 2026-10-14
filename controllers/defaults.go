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
	"fmt"
	"os"
	"time"
)

const (
	OdfSubscriptionPackage      = "odf-operator"
	OdfDepsSubscriptionPackage  = "odf-dependencies"
	CnsaDepsSubscriptionPackage = "cnsa-dependencies"
	TLSProfileName              = "ocs-tls-profile"
)

const (
	// The internal ConfigMap stores the last accepted odf-operator CSV and its
	// build date. During an odf-operator upgrade, the current CSV build date is
	// compared with the stored build date before dependent operator subscription
	// channels are updated.

	// If the current CSV is older than the allowed threshold, dependencies upgrade
	// are blocked and the ConfigMap is updated with the blocked state and message.
	// Users can bypass this validation by setting the
	// odf.openshift.io/skip-build-date-check annotation to "true".

	// apiVersion: v1
	// kind: ConfigMap
	// metadata:
	//   name: odf-operator-internal-config
	//   namespace: openshift-storage
	// data:
	//   LAST_ACCEPTED_CSV: odf-operator.v4.21.12-rhodf
	//   LAST_ACCEPTED_CSV_BUILD_DATE: "2026-04-22T09:54:14Z"
	//   DEPENDENCIES_UPGRADE_BLOCKED: "true"
	//   DEPENDENCIES_UPGRADE_MESSAGE: >-
	//     the target version may not include the bug fixes that are present
	//     in the currently running version

	OdfOperatorInternalConfigMapName = "odf-operator-internal-config"

	skipBuildDateCheckAnnotation = "odf.openshift.io/skip-build-date-check"

	lastAcceptedCSVKey          = "LAST_ACCEPTED_CSV"
	lastAcceptedCSVBuildDateKey = "LAST_ACCEPTED_CSV_BUILD_DATE"

	dependenciesUpgradeBlockedKey = "DEPENDENCIES_UPGRADE_BLOCKED"
	dependenciesUpgradeMessageKey = "DEPENDENCIES_UPGRADE_MESSAGE"

	csvBuildDateAnnotation = "createdAt"

	allowedThreshold = 7 * 24 * time.Hour

	dependenciesUpgradeBlockedErrMsg = "upgrade is paused because the target version may be missing bug fixes present in the current version"
)

var (
	DepsSubscriptionPackageNames = []string{
		OdfDepsSubscriptionPackage,
		CnsaDepsSubscriptionPackage,
	}
)

var (
	OperatorNamespace        string
	odfOperatorConfigMapName string
)

func init() {
	OperatorNamespace = GetEnvOrPanic("OPERATOR_NAMESPACE")
	odfOperatorConfigMapName = GetEnvOrPanic("PKGS_CONFIG_MAP_NAME")
}

func GetEnvOrPanic(env string) string {
	if val := os.Getenv(env); val != "" {
		return val
	}
	panic(fmt.Sprintf("Environment variable %s is not set", env))
}
