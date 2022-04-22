#!/bin/bash
set -ex
cp ~/.kube/config ./kube.config
kubectl create configmap k8s-client-config --from-file=./kube.config