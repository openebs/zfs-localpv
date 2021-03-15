#!/bin/bash
  
set -e

mkdir app_yamls_immediate

for i in $(seq 1 5)
do
        sed "s/pvc-custom-topology/pvc-custom-topology-$i/g" busybox.yml > app_yamls_immediate/busybox-$i.yml
        sed -i "s/busybox-deploy-custom-topology/busybox-deploy-custom-topology-$i/g" app_yamls_immediate/busybox-$i.yml
        sed -i "s/storageClassName: zfspv-custom-topology/storageClassName: zfspv-custom-topology-immediate/g" app_yamls_immediate/busybox-$i.yml
done