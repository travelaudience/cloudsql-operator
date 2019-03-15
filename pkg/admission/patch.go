/*
Copyright 2019 The cloudsql-operator Authors.

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

package admission

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/appscode/jsonpatch"
	"k8s.io/apimachinery/pkg/runtime"
)

// CreateRFC6902Patch creates an RFC6902 patch that captures the difference between the specified objects.
func CreateRFC6902Patch(oldObj, newObj runtime.Object) ([]byte, error) {
	// Make sure we're dealing with resources of the same GVK.
	oldGVK := oldObj.GetObjectKind().GroupVersionKind()
	newGVK := newObj.GetObjectKind().GroupVersionKind()
	if !reflect.DeepEqual(oldGVK, newGVK) {
		return nil, fmt.Errorf("gvk mismatch (expected %v, got %v)", oldGVK, newGVK)
	}
	// Marshal the old object.
	oldBytes, err := json.Marshal(oldObj)
	if err != nil {
		return nil, err
	}
	// Marshal the new object.
	newBytes, err := json.Marshal(newObj)
	if err != nil {
		return nil, err
	}
	// Create the RFC6902 patch based on the representations of the old and new objects.
	r, err := jsonpatch.CreatePatch(oldBytes, newBytes)
	if err != nil {
		return nil, err
	}
	// Return a byte array containing the patch.
	return json.Marshal(r)
}
