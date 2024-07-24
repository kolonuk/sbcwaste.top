#!/bin/bash

# Recommended tip: install go air for live-reloads: https://github.com/air-verse/air
# This script attempts to configure a both a dev environment which allows local testing, and requirements for remote deployment

# Update OS and apt db
sudo apt-get update && apt-get upgrade -y

# Install Go
sudo apt-get install -y golang-go

# setup requirements for chromdp
sudo apt-get install -y \
    ca-certificates \
    fonts-liberation \
    libappindicator3-1 \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcups2 \
    libdbus-1-3 \
    libgdk-pixbuf2.0-0 \
    libnspr4 \
    libnss3 \
    libx11-xcb1 \
    lsb-release \
    wget \
    xdg-utils \
    jq \
    chromium

# setup google cloud sdk
sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
    gnupg \
    curl
sudo rm -f /usr/share/keyrings/cloud.google.gpg
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg
echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee /etc/apt/sources.list.d/google-cloud-sdk.list
sudo apt-get update && sudo apt-get install -y google-cloud-cli

# configure google cloud sdk
gcloud init
gcloud auth list



