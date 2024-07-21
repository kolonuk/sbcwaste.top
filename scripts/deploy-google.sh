#!/bin/bash

# tidy up the go mod file
go mod tidy

# set the configuration for the project
gcloud config set project sbcwaste

# enable the artifact registry
echo y|gcloud services enable artifactregistry.googleapis.com

# create a docker repository
gcloud artifacts repositories create eu-docker-repo \
  --repository-format=docker \
  --location=europe-west2 \
  --description="EU Docker repository"

# configure docker to use gcloud
echo y|gcloud auth configure-docker europe-west2-docker.pkg.dev

# build the docker image
docker build -t europe-west2-docker.pkg.dev/sbcwaste/eu-docker-repo/sbcwaste .

# push the docker image to the container registry
docker push europe-west2-docker.pkg.dev/sbcwaste/eu-docker-repo/sbcwaste

# deploy the docker image to cloud run
echo y|gcloud run deploy sbcwaste --image europe-west2-docker.pkg.dev/sbcwaste/eu-docker-repo/sbcwaste --platform managed --region europe-west2

# get the domain name from the deployment ( not the url)
DOMAIN_NAME=$(gcloud run services describe sbcwaste --platform managed --region europe-west2 --format 'value(status.domain)')

# Enable the DNS service
echo y|gcloud services enable dns.googleapis.com

# Create a managed zone
gcloud dns managed-zones create sbcwaste-zone --dns-name=sbcwaste.top. --description="DNS zone for sbcwaste.top"
gcloud dns managed-zones describe sbcwaste-zone

# redirect the domain to the service URL (TLD)
gcloud dns record-sets transaction abort --zone=sbcwaste-zone
gcloud dns record-sets transaction start --zone=sbcwaste-zone --project=sbcwaste
# List and remove all CNAME records for www.sbcwaste.top.
gcloud dns record-sets list --zone=sbcwaste-zone --name="www.sbcwaste.top." --type=CNAME --format="value(name,ttl,rrdatas[0])" | while read -r name ttl rrdata; do
    gcloud dns record-sets transaction remove --name="$name" --type=CNAME --ttl="$ttl" --zone=sbcwaste-zone "$rrdata"
done
gcloud dns record-sets transaction add $DOMAIN_NAME --name="www.sbcwaste.top." --type=CNAME --ttl=300 --zone=sbcwaste-zone
gcloud dns record-sets transaction execute --zone=sbcwaste-zone

# abandon the transaction
#gcloud dns record-sets transaction abort --zone=sbcwaste-zone

# do the TLD redirect to the cname www sub domain
echo y|gcloud services enable cloudbuild.googleapis.com
echo y|gcloud functions delete redirect --region=europe-west2 --gen2
gcloud functions deploy redirect --source=./redirect --runtime nodejs20 --trigger-http --allow-unauthenticated --region=europe-west2 --gen2 --entry-point=redirectToWWW

