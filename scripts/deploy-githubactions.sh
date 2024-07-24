#!/bin/bash

PROJECT_ID="sbcwaste"
GITHUB_ORG="kolonuk"      # github username, or your org name if part of an org
REPO_NAME="kolonuk/sbcwaste.top"

echo y|gcloud services enable artifactregistry.googleapis.com
echo y|gcloud services enable run.googleapis.com
echo y|gcloud services enable secretmanager.googleapis.com
echo y|gcloud services enable iamcredentials.googleapis.com

gcloud iam service-accounts create "${PROJECT_ID}" \
    --description="${PROJECT_ID}_sa" \
    --project="${PROJECT_ID}"
    
gcloud iam workload-identity-pools create "github" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --display-name="GitHub Actions Pool"

WIPOOL=$(gcloud iam workload-identity-pools describe "github" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --format="value(name)")

gcloud iam workload-identity-pools providers create-oidc "my-repo" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --workload-identity-pool="github" \
  --display-name="My GitHub repo Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository,attribute.repository_owner=assertion.repository_owner" \
  --attribute-condition="assertion.repository_owner == '${GITHUB_ORG}'" \
  --issuer-uri="https://token.actions.githubusercontent.com"

# TODO: replace ${PROJECT_ID} with your value below.

WIPROVIDER=$(gcloud iam workload-identity-pools providers describe "my-repo" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --workload-identity-pool="github" \
  --format="value(name)")

# Grant roles/run.admin
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --role="roles/run.admin" \
  --member="principalSet://iam.googleapis.com/${WIPOOL}/attribute.repository/${REPO_NAME}"

# Grant roles/iam.serviceAccountUser
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --role="roles/iam.serviceAccountUser" \
  --member="principalSet://iam.googleapis.com/${WIPOOL}/attribute.repository/${REPO_NAME}"

# Grant roles/artifactregistry.admin
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --role="roles/artifactregistry.admin" \
  --member="principalSet://iam.googleapis.com/${WIPOOL}/attribute.repository/${REPO_NAME}"

# Grant roles/artifactregistry.admin
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --role="roles/iam.serviceAccountTokenCreator" \
  --member="principalSet://iam.googleapis.com/${WIPOOL}/attribute.repository/${REPO_NAME}"

#Create GitHub secrets for WIF_PROVIDER and WIF_SERVICE_ACCOUNT
echo In your Github project, create the following secrets:
echo
echo WIF_PROVIDER: $WIPROVIDER
echo WIF_SERVICE_ACCOUNT: $PROJECT_ID@$PROJECT_ID.iam.gserviceaccount.com
echo
gcloud iam service-accounts list --project="${PROJECT_ID}"



