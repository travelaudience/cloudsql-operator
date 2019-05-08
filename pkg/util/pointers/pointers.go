/*
Copyright 2019 The cloudsql-postgres-operator Authors.

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

package pointers

// NewBool returns a pointer to the specified boolean value.
func NewBool(v bool) *bool {
	return &v
}

// NewInt32 returns a pointer to the specified integer value.
func NewInt32(v int32) *int32 {
	return &v
}

// NewInt64 returns a pointer to the specified integer value.
func NewInt64(v int64) *int64 {
	return &v
}

// NewString returns a pointer to the specified string value.
func NewString(v string) *string {
	return &v
}
