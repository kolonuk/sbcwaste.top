#!/bin/bash

PROJECT_ID="sbcwaste"
GITHUB_ORG="kolonuk"      # github username, or your org name if part of an org
REPO_NAME="kolonuk/sbcwaste.top"
PROJECT_NUMBER=$(gcloud projects describe ${PROJECT_ID} --format='value(projectNumber)')
WIP_NAME="github"

echo y|gcloud services enable artifactregistry.googleapis.com
echo y|gcloud services enable run.googleapis.com
echo y|gcloud services enable secretmanager.googleapis.com
echo y|gcloud services enable iamcredentials.googleapis.com

gcloud iam workload-identity-pools create "${WIP_NAME}" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --display-name="GitHub Actions Pool"

WIPOOL=$(gcloud iam workload-identity-pools describe "${WIP_NAME}" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --format="value(name)")

gcloud iam workload-identity-pools providers create-oidc "my-repo" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --workload-identity-pool="${WIP_NAME}" \
  --display-name="My GitHub repo Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository,attribute.repository_owner=assertion.repository_owner" \
  --attribute-condition="assertion.repository_owner == '${GITHUB_ORG}'" \
  --issuer-uri="https://token.actions.githubusercontent.com"

WIPROVIDER=$(gcloud iam workload-identity-pools providers describe "my-repo" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --workload-identity-pool="${WIP_NAME}" \
  --format="value(name)")

## Create service account and assign roles
gcloud iam service-accounts create "${PROJECT_ID}" \
    --description="${PROJECT_ID}_sa" \
    --project="${PROJECT_ID}"

gcloud iam service-accounts add-iam-policy-binding ${PROJECT_ID}@${PROJECT_ID}.iam.gserviceaccount.com \
    --role=roles/iam.serviceAccountTokenCreator \
    --member="principalSet://iam.googleapis.com/${WIPOOL}/attribute.repository/${REPO_NAME}" \
    --project=${PROJECT_ID} > /dev/null

gcloud iam service-accounts add-iam-policy-binding ${PROJECT_ID}@${PROJECT_ID}.iam.gserviceaccount.com \
    --role=roles/iam.workloadIdentityUser \
    --member="principalSet://iam.googleapis.com/${WIPOOL}/attribute.repository/${REPO_NAME}" \
    --project=${PROJECT_ID} > /dev/null

# gcloud artifacts repositories add-iam-policy-binding ${PROJECT_ID} \
#   --location=europe-west1 \
#   --member="serviceAccount:${PROJECT_ID}@${PROJECT_ID}.iam.gserviceaccount.com" \
#   --role="roles/artifactregistry.writer" \
#   --project=sbcwaste

## Grant roles to project IAM
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --role="roles/run.admin" \
  --member="principalSet://iam.googleapis.com/${WIPOOL}/attribute.repository/${REPO_NAME}" \
  --quiet > /dev/null

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --role="roles/iam.serviceAccountUser" \
  --member="principalSet://iam.googleapis.com/${WIPOOL}/attribute.repository/${REPO_NAME}" \
  --quiet > /dev/null

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --role="roles/artifactregistry.admin" \
  --member="principalSet://iam.googleapis.com/${WIPOOL}/attribute.repository/${REPO_NAME}" \
  --quiet > /dev/null

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --role="roles/iam.serviceAccountTokenCreator" \
  --member="principalSet://iam.googleapis.com/${WIPOOL}/attribute.repository/${REPO_NAME}" \
  --quiet > /dev/null

## Display secrets required for Github actions
echo
echo
echo In your Github project, create the following secrets:
echo
echo WIF_PROVIDER: $WIPROVIDER
echo WIF_SERVICE_ACCOUNT: $PROJECT_ID@$PROJECT_ID.iam.gserviceaccount.com
echo
#gcloud iam service-accounts list --filter="email:${PROJECT_ID}@${PROJECT_ID}.iam.gserviceaccount.com"
#gcloud projects get-iam-policy "${PROJECT_ID}" --format=json | jq '.bindings[] | select(.members[] | contains("serviceAccount:$PROJECT_ID@$PROJECT_ID.iam.gserviceaccount.com"))'

