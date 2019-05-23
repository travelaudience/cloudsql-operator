#!/bin/bash

set -e

# base64 needs different args depending on the flavor of the tool that is installed.
base64w () {
    (base64 --version >/dev/null 2>&1 && base64 -w 0) || base64 --break 0
}

# sed needs different args to -i depending on the flavor of the tool that is installed.
sedi () {
    (sed --version >/dev/null 2>&1 && sed -i "$@") || sed -i "" "$@"
}

# NAMESPACE is the namespace where to deploy "cloudsql-postgres-operator".
NAMESPACE=cloudsql-postgres-operator
# ROOT_DIR is the absolute path to the root of the repository.
ROOT_DIR="$(git rev-parse --show-toplevel)"
# TMP_DIR is the path (relative to ROOT_DIR) where to copy manifest templates to.
TMP_DIR="tmp/skaffold/operator"

# ADMIN_KEY_JSON_FILE is the path to the file containing the credentials of the "admin" IAM service account.
ADMIN_KEY_JSON_FILE="${ADMIN_KEY_JSON_FILE:-${ROOT_DIR}/admin-key.json}"
# CLIENT_KEY_JSON_FILE is the path to the file containing the credentials of the "client" IAM service account.
CLIENT_KEY_JSON_FILE="${CLIENT_KEY_JSON_FILE:-${ROOT_DIR}/client-key.json}"
# PROFILE is the skaffold profile to use.
PROFILE=${PROFILE:-local}
# PROJECT_ID is the ID of the Google Cloud Platform project where cloudsql-postgres-operator should manage CSQLP instances.
PROJECT_ID=${PROJECT_ID:-cloudsql-operator}

# Switch directories to "ROOT_DIR".
pushd "${ROOT_DIR}" > /dev/null

# Create the temporary directory if it does not exist.
mkdir -p "${TMP_DIR}"
# Copy manifest templates to the temporary directory.
cp -r "${ROOT_DIR}/docs/deployment/02-service-account.yaml" "${TMP_DIR}/"
cp -r "${ROOT_DIR}/hack/skaffold/operator/"* "${TMP_DIR}/"

# Replace the "__TMP_DIR__" placeholder.
sedi -e "s|__TMP_DIR__|${TMP_DIR}|" "${TMP_DIR}/"*.yaml
# Replace the "__PROJECT_ID__" placeholder.
sedi -e "s|__PROJECT_ID__|${PROJECT_ID}|g" "${TMP_DIR}/"*.yaml
# Replace the "__BASE64_ENCODED_ADMIN_KEY_JSON__" placeholder.
BASE64_ENCODED_ADMIN_KEY_JSON="$(base64w < "${ADMIN_KEY_JSON_FILE}")"
sedi -e "s|__BASE64_ENCODED_ADMIN_KEY_JSON__|${BASE64_ENCODED_ADMIN_KEY_JSON}|g" "${TMP_DIR}/"*.yaml
# Replace the "__BASE64_ENCODED_CLIENT_KEY_JSON__" placeholder.
BASE64_ENCODED_CLIENT_KEY_JSON="$(base64w < "${CLIENT_KEY_JSON_FILE}")"
sedi -e "s|__BASE64_ENCODED_CLIENT_KEY_JSON__|${BASE64_ENCODED_CLIENT_KEY_JSON}|g" "${TMP_DIR}/"*.yaml

# Build the cloudsql-postgres-operator binary.
make -C "${ROOT_DIR}" "build"

# Validate the profile and configure skaffold as required.
case "${PROFILE}" in
    "gke")
        # Override IMAGE_PULL_POLICY.
        IMAGE_PULL_POLICY="IfNotPresent"
        ;;
    "kind")
        # Let skaffold know that the current context is local.
        # NOTE: This assumes that the current context does indeed point at a kind cluster.
        skaffold config set --kube-context "$(kubectl config current-context)" local-cluster true > /dev/null 2>&1
        # Kind requires local Docker images to be manually loaded using "kind load docker-image".
        # Hence, we must build the image beforehand and manually load it before deploying.
        kind load docker-image "$(skaffold build -f "${TMP_DIR}/skaffold.yaml" -n "${NAMESPACE}" -p "${PROFILE}" --quiet --output "{{(index .Builds 0).Tag}}")"
        # Override IMAGE_PULL_POLICY.
        IMAGE_PULL_POLICY="Never"
        ;;
    "minikube")
        # Override IMAGE_PULL_POLICY.
        IMAGE_PULL_POLICY="Never"
        ;;
    *)
        echo "unsupported profile: \"${PROFILE}\"" && exit 1
esac

# Replace the "__IMAGE_PULL_POLICY__" placeholder.
sedi -e "s|__IMAGE_PULL_POLICY__|${IMAGE_PULL_POLICY}|g" "${TMP_DIR}/"*.yaml

# Make sure the target namespace exists.
kubectl get namespace "${NAMESPACE}" > /dev/null 2>&1 || kubectl create namespace "${NAMESPACE}"

# Run skaffold in development mode.
skaffold run -f "${TMP_DIR}/skaffold.yaml" -n "${NAMESPACE}" -p "${PROFILE}" --tail
