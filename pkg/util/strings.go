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

package util

// FindInSlice find given string in given slice
func FindInSlice(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// RemoveFromSlice removes given string from given slice
func RemoveFromSlice(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// Find returns the first entry matching the function "f" or else return nil
func Find[T any](list []T, f func(item *T) bool) *T {
	for idx := range list {
		ele := &list[idx]
		if f(ele) {
			return ele
		}
	}
	return nil
}
