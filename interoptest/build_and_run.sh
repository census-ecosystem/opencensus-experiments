#!/bin/sh
set -e

echo "Building containers.."
skaffold build --profile travis-ci

./run.sh
