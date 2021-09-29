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
	"os"
)

var (
	DefaultValMap = map[string]string{
		"OPERATOR_NAMESPACE": "openshift-storage",

		"NOOBAA_SUBSCRIPTION_NAME":                    "noobaa-operator",
		"NOOBAA_SUBSCRIPTION_PACKAGE":                 "noobaa-operator",
		"NOOBAA_SUBSCRIPTION_CHANNEL":                 "alpha",
		"NOOBAA_SUBSCRIPTION_STARTINGCSV":             "noobaa-operator.v5.9.0",
		"NOOBAA_SUBSCRIPTION_CATALOGSOURCE":           "odf-catalogsource",
		"NOOBAA_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE": "openshift-storage",

		"OCS_SUBSCRIPTION_NAME":                    "ocs-operator",
		"OCS_SUBSCRIPTION_PACKAGE":                 "ocs-operator",
		"OCS_SUBSCRIPTION_CHANNEL":                 "alpha",
		"OCS_SUBSCRIPTION_STARTINGCSV":             "ocs-operator.v4.9.0",
		"OCS_SUBSCRIPTION_CATALOGSOURCE":           "odf-catalogsource",
		"OCS_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE": "openshift-storage",

		"IBM_SUBSCRIPTION_NAME":                    "ibm-storage-odf-operator",
		"IBM_SUBSCRIPTION_PACKAGE":                 "ibm-storage-odf-operator",
		"IBM_SUBSCRIPTION_CHANNEL":                 "stable-v1",
		"IBM_SUBSCRIPTION_STARTINGCSV":             "ibm-storage-odf-operator.v0.2.0",
		"IBM_SUBSCRIPTION_CATALOGSOURCE":           "odf-catalogsource",
		"IBM_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE": "openshift-storage",
	}

	OperatorNamespace = GetEnvOrDefault("OPERATOR_NAMESPACE")

	OcsSubscriptionName                   = GetEnvOrDefault("OCS_SUBSCRIPTION_NAME")
	OcsSubscriptionPackage                = GetEnvOrDefault("OCS_SUBSCRIPTION_PACKAGE")
	OcsSubscriptionChannel                = GetEnvOrDefault("OCS_SUBSCRIPTION_CHANNEL")
	OcsSubscriptionStartingCSV            = GetEnvOrDefault("OCS_SUBSCRIPTION_STARTINGCSV")
	OcsSubscriptionCatalogSource          = GetEnvOrDefault("OCS_SUBSCRIPTION_CATALOGSOURCE")
	OcsSubscriptionCatalogSourceNamespace = GetEnvOrDefault("OCS_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE")

	NoobaaSubscriptionName                   = GetEnvOrDefault("NOOBAA_SUBSCRIPTION_NAME")
	NoobaaSubscriptionPackage                = GetEnvOrDefault("NOOBAA_SUBSCRIPTION_PACKAGE")
	NoobaaSubscriptionChannel                = GetEnvOrDefault("NOOBAA_SUBSCRIPTION_CHANNEL")
	NoobaaSubscriptionStartingCSV            = GetEnvOrDefault("NOOBAA_SUBSCRIPTION_STARTINGCSV")
	NoobaaSubscriptionCatalogSource          = GetEnvOrDefault("NOOBAA_SUBSCRIPTION_CATALOGSOURCE")
	NoobaaSubscriptionCatalogSourceNamespace = GetEnvOrDefault("NOOBAA_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE")

	IbmSubscriptionName                   = GetEnvOrDefault("IBM_SUBSCRIPTION_NAME")
	IbmSubscriptionPackage                = GetEnvOrDefault("IBM_SUBSCRIPTION_PACKAGE")
	IbmSubscriptionChannel                = GetEnvOrDefault("IBM_SUBSCRIPTION_CHANNEL")
	IbmSubscriptionStartingCSV            = GetEnvOrDefault("IBM_SUBSCRIPTION_STARTINGCSV")
	IbmSubscriptionCatalogSource          = GetEnvOrDefault("IBM_SUBSCRIPTION_CATALOGSOURCE")
	IbmSubscriptionCatalogSourceNamespace = GetEnvOrDefault("IBM_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE")
)

func GetEnvOrDefault(env string) string {
	if val := os.Getenv(env); val != "" {
		return val
	}

	return DefaultValMap[env]
}
