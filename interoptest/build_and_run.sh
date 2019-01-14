#!/bin/sh
set -e

# First build
echo "Building containers.."
skaffold build --profile travis-ci

# Following is adapted from https://github.com/lawrencegripper/azurefrontdooringress/blob/master/scripts/startminikube_ci.sh

# Install
export CHANGE_MINIKUBE_NONE_USER=true

echo "--> Downloading minikube"
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.12.0/bin/linux/amd64/kubectl
chmod +x kubectl && sudo mv kubectl /usr/local/bin/
curl -Lo minikube https://storage.googleapis.com/minikube/releases/0.30.0/minikube-linux-amd64
chmod +x minikube && sudo mv minikube /usr/local/bin/

echo "--> Starting minikube"
sudo minikube start --vm-driver=none --bootstrapper=kubeadm --kubernetes-version=v1.12.0
# Fix the kubectl context, as it's often stale.
minikube update-context

echo "--> Waiting for cluster to be usable"
# Wait for Kubernetes to be up and ready.
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1; done
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl -n kube-system get pods -lcomponent=kube-addon-manager -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1;echo "waiting for kube-addon-manager to be available"; kubectl get pods --all-namespaces; done
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl -n kube-system get pods -lk8s-app=kube-dns -o jsonpath="$

echo "--> Get cluster details to check its running"
kubectl cluster-info

# Run the containers
skaffold run
sleep 60

# Run tests
cd ./src/testcontroller
sudo pip install -r requirements.txt
python run.py
