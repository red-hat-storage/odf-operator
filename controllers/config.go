/*
Copyright 2025 Red Hat OpenShift Data Foundation.

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

type OlmPkgRecord struct {
	/* example
	   channel: alpha
	   csv: ocs-operator.v4.18.0
	   pkg: ocs-operator
	*/
	Channel string `yaml:"channel"`
	Csv     string `yaml:"csv"`
	Pkg     string `yaml:"pkg"`
}
