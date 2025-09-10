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
)

const (
	OdfSubscriptionPackage      = "odf-operator"
	OdfDepsSubscriptionPackage  = "odf-dependencies"
	CnsaDepsSubscriptionPackage = "cnsa-dependencies"
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
