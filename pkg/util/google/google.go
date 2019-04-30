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

package google

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	goauth "golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

const (
	// adminScope is the Google Cloud Platform scope required for making calls to the Cloud SQL Admin API.
	adminScope = "https://www.googleapis.com/auth/sqlservice.admin"
)

// IsBadRequest indicates whether the specified error is the result of the Cloud SQL Admin API replying with http.StatusBadRequest.
func IsBadRequest(err error) bool {
	if err == nil {
		return false
	}
	ae, ok := err.(*googleapi.Error)
	return ok && ae.Code == http.StatusBadRequest
}

// IsConflict indicates whether the specified error is the result of the Cloud SQL Admin API replying with http.StatusConflict.
func IsConflict(err error) bool {
	if err == nil {
		return false
	}
	ae, ok := err.(*googleapi.Error)
	return ok && ae.Code == http.StatusConflict
}

// IsNotFound indicates whether the specified error is the result of the Cloud SQL Admin API replying with http.StatusNotFound.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	ae, ok := err.(*googleapi.Error)
	return ok && ae.Code == http.StatusNotFound
}

// NewCloudSQLAdminClient creates a client to the Cloud SQL Admin API that uses the specified IAM service account credentials file for authentication.
func NewCloudSQLAdminClient(keyPath string) (*sqladmin.Service, error) {
	c, err := newHTTPClient(keyPath)
	if err != nil {
		return nil, err
	}
	return sqladmin.NewService(context.Background(), option.WithHTTPClient(c))
}

// newHTTPClient returns an HTTP client that uses the specified IAM service account credentials file for authentication.
func newHTTPClient(keyPath string) (*http.Client, error) {
	if keyPath == "" {
		return nil, fmt.Errorf("the path to the \"admin\" iam service account key must be specified")
	}
	b, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	c, err := goauth.JWTConfigFromJSON(b, adminScope)
	if err != nil {
		return nil, err
	}
	return c.Client(context.Background()), nil
}
