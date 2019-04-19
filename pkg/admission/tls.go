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

package admission

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/cert"

	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
)

const (
	// tlsSecretName is the name of the secret that contains the TLS material used to register and serve the admission webhook.
	tlsSecretName = "cloudsql-postgres-operator-tls"
)

// ensureTLSSecret generates a self-signed certificate and a private key to be used for registering and serving the admission webhook.
// It then creates a secret containing them so they can be used by all running instances of cloudsql-postgres-operator.
// In case such secret already exists, it is read and returned instead.
func (w *Webhook) ensureTLSSecret() (*corev1.Secret, error) {
	// Generate a self-signed certificate valid for "<service-name>.<namespace>.svc".
	svc := fmt.Sprintf("%s.%s.svc", cloudsqlPostgresOperatorServiceName, w.namespace)
	now := time.Now()
	crt := x509.Certificate{
		Subject: pkix.Name{
			CommonName: svc,
		},
		NotBefore:    now,
		NotAfter:     now.Add(10 * 365 * 24 * time.Hour),
		SerialNumber: big.NewInt(now.Unix()),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		IsCA:                  true,
		BasicConstraintsValid: true,
		DNSNames: []string{
			svc,
		},
	}
	// Generate a private key.
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}
	// PEM-encode the private key.
	keyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  cert.RSAPrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	// Self-sign the certificate using the private key.
	sig, err := x509.CreateCertificate(rand.Reader, &crt, &crt, key.Public(), key)
	if err != nil {
		return nil, fmt.Errorf("failed to self-sign certificate: %v", err)
	}
	// PEM-encode the signed certificate
	sigBytes := pem.EncodeToMemory(&pem.Block{
		Type:  cert.CertificateBlockType,
		Bytes: sig,
	})

	// Create a secret containing the generated self-signed certificate and private key.
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				constants.LabelAppKey: constants.ApplicationName,
			},
			Name:      tlsSecretName,
			Namespace: w.namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       sigBytes,
			corev1.TLSPrivateKeyKey: keyBytes,
		},
	}
	// Try to actually create the secret in the Kubernetes API.
	sec, err = w.kubeClient.CoreV1().Secrets(w.namespace).Create(sec)
	if err == nil {
		// Creation was successful, so there's nothing left to do.
		return sec, nil
	}
	if errors.IsAlreadyExists(err) {
		// The secret already exists, so we should read and reuse it.
		return w.kubeClient.CoreV1().Secrets(w.namespace).Get(tlsSecretName, metav1.GetOptions{})
	}
	// The secret doesn't exist, but we couldn't create it and hence should fail.
	return nil, fmt.Errorf("failed to create the admission webhook's tls secret: %v", err)
}
