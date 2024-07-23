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
  --location=europe-west1 \
  --description="EU Docker repository"

# configure docker to use gcloud
echo y|gcloud auth configure-docker europe-west1-docker.pkg.dev

# build the docker image
docker build -t europe-west1-docker.pkg.dev/sbcwaste/eu-docker-repo/sbcwaste .

# push the docker image to the container registry
docker push europe-west1-docker.pkg.dev/sbcwaste/eu-docker-repo/sbcwaste

# deploy the docker image to cloud run
echo y|gcloud run deploy sbcwaste --image europe-west1-docker.pkg.dev/sbcwaste/eu-docker-repo/sbcwaste --platform managed --region europe-west1

# get the cloud run domain name from the deployment (not the url)
DOMAIN_NAME=$(gcloud run services describe sbcwaste --platform managed --region europe-west1 --format 'value(status.domain)')

gcloud domains list-user-verified
#gcloud domains verify sbcwaste.top
# Enable the DNS service
echo y|gcloud services enable dns.googleapis.com

# Create a managed zone
gcloud dns managed-zones create sbcwaste-zone --dns-name=sbcwaste.top. --description="DNS zone for sbcwaste.top"
gcloud dns managed-zones describe sbcwaste-zone

### Remove current DNS records
gcloud dns record-sets transaction abort --zone=sbcwaste-zone
gcloud dns record-sets transaction start --zone=sbcwaste-zone

# List A records and remove them
gcloud dns record-sets list --zone=sbcwaste-zone --name="sbcwaste.top." --type=A --format="value(name,ttl,rrdatas)" | while IFS=$'\t' read -r name ttl rrdatas; do
  IFS=';' read -ra ADDR <<< "$rrdatas" # Split rrdatas into an array based on comma
  rrdata_args=""
  for ip in "${ADDR[@]}"; do
    rrdata_args+="$ip "
  done
  gcloud dns record-sets transaction remove $rrdata_args --name="$name" --type=A --ttl="$ttl" --zone=sbcwaste-zone
done

# List AAAA records and remove them
gcloud dns record-sets list --zone=sbcwaste-zone --name="sbcwaste.top." --type=AAAA --format="value(name,ttl,rrdatas)" | while IFS=$'\t' read -r name ttl rrdatas; do
  IFS=';' read -ra ADDR <<< "$rrdatas" # Split rrdatas into an array based on comma
  rrdata_args=""
  for ip in "${ADDR[@]}"; do
    rrdata_args+="$ip "
  done
  gcloud dns record-sets transaction remove $rrdata_args --name="$name" --type=AAAA --ttl="$ttl" --zone=sbcwaste-zone
done

gcloud dns record-sets transaction execute --zone=sbcwaste-zone

### Enable domain mapping to cloud run project
gcloud beta run domain-mappings delete --domain=sbcwaste.top --platform managed --region europe-west1
output=$(gcloud beta run domain-mappings create --service=sbcwaste --domain=sbcwaste.top --platform managed --region europe-west1)

### Set DNS records to cloud run project
gcloud dns record-sets transaction start --zone=sbcwaste-zone

ip_addresses=()
while read -r ip; do
  ip_addresses+=("$ip")
done < <(echo "$output" | grep -E 'sbcwaste\s+A ' | awk '{print $3}')
gcloud dns record-sets transaction add ${ip_addresses[0]} ${ip_addresses[1]} ${ip_addresses[2]} ${ip_addresses[3]} --name=sbcwaste.top. --type=A --ttl=300 --zone=sbcwaste-zone

ip_addresses=()
while read -r ip; do
  ip_addresses+=("$ip")
done < <(echo "$output" | grep -E 'sbcwaste\s+AAAA' | awk '{print $3}')
gcloud dns record-sets transaction add ${ip_addresses[0]} ${ip_addresses[1]} ${ip_addresses[2]} ${ip_addresses[3]} --name=sbcwaste.top. --type=AAAA --ttl=300 --zone=sbcwaste-zone

gcloud dns record-sets transaction execute --zone=sbcwaste-zone

exit 0

# Enable the DNS service
echo y|gcloud services enable dns.googleapis.com

# Create a managed zone
gcloud dns managed-zones create sbcwaste-zone --dns-name=sbcwaste.top. --description="DNS zone for sbcwaste.top"
gcloud dns managed-zones describe sbcwaste-zone

gcloud dns record-sets transaction abort --zone=sbcwaste-zone
gcloud dns record-sets transaction start --zone=sbcwaste-zone --project=sbcwaste
gcloud dns record-sets transaction add google-site-verification=WHwMPTinQMVIdCnJPcloo0cqD5DG9PEm6b6bsHTDENo --name="sbcwaste.top." --type=TXT --ttl=300 --zone=sbcwaste-zone
gcloud dns record-sets transaction execute --zone=sbcwaste-zone

exit 0

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
echo y|gcloud functions delete redirect --region=europe-west1 --gen2
gcloud functions deploy redirect --source=./redirect --runtime nodejs20 --trigger-http --allow-unauthenticated --region=europe-west1 --gen2 --entry-point=redirectToWWW

