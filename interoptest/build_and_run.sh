#!/bin/sh
set -e

# First build
echo "Building containers.."
skaffold build --profile travis-ci

# Run the containers
skaffold run
sleep 60

# Run tests
cd ./src/testcontroller
sudo pip install -r requirements.txt
python run.py
