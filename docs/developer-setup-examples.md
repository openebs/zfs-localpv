# kubernetes and os setup examples

This documents contains several examples to show working setups for a development environment

## tested on ubuntu 18.04.4 with minikube 1.9.2 and zfs 0.7 and 0.8

### Prerequisites
* You have a zfs pool created on localhost, with pool name `zfspv-pool`; that is you have installed
zfs-dkms (or zfs-fuse), zfs-zed and zfsutils-linux .deb packages. And then created the pool with `zpool create` command
* You have followed the prerequisites stated in [Development Workflow](developer-setup.md)

```
wget https://github.com/kubernetes/minikube/releases/download/v1.9.2/minikube_1.9.2-0_amd64.deb
sudo dpkg -i minikube_1.9.2-0_amd64.deb
sudo minikube start --driver=none
sudo chown -R $USER $HOME/.kube $HOME/.minikube

kubectl apply -f https://raw.githubusercontent.com/openebs/zfs-localpv/master/deploy/zfs-operator.yaml
kubectl get pods -n kube-system -l role=openebs-zfs

export OPENEBS_NAMESPACE=openebs
export KUBECONFIG=$HOME/.kube/config

cd ~/go/src/github.com/openebs/zfs-localpv/tests

ginkgo -v
```

All the tests should pass now.

## Example output of a successful integration test on development environment

```sh
Running Suite: Test ZFSPV volume provisioning
=============================================
Random Seed: 1586718777
Will run 1 of 1 specs

[zfspv] TEST VOLUME PROVISIONING App is deployed with zfs driver 
  Running zfs volume Creation Test
  /home/filippo/go/src/github.com/openebs/zfs-localpv/tests/provision_test.go:25
STEP: Running dataset creation test
STEP: Creating zfs storage class
STEP: building a zfs storage class
STEP: creating and verifying PVC bound status
STEP: building a pvc
STEP: creating above pvc
STEP: verifying pvc status as bound
STEP: Creating and deploying app pod
STEP: creating and deploying app pod
STEP: building a busybox app pod deployment using above zfs volume
STEP: verifying app pod is running
STEP: verifying ZFSVolume object
STEP: fetching zfs volume
STEP: verifying zfs volume
STEP: Resizing the PVC
STEP: updating the pvc with new size
STEP: verifying pvc size to be updated
STEP: verifying ZFSVolume property change
STEP: verifying compression property update
STEP: fetching zfs volume for setting compression=on
STEP: fetching zfs volume for setting compression=off
STEP: verifying dedup property update
STEP: fetching zfs volume for setting dedup=on
STEP: fetching zfs volume for setting dedup=off
STEP: verifying recordsize property update
STEP: fetching zfs volume for setting the recordsize
STEP: Deleting application deployment
STEP: Deleting pvc
STEP: verifying deleted pvc
STEP: Deleting storage class
STEP: Running zvol creation test
STEP: Creating ext4 storage class
STEP: building a ext4 storage class
STEP: creating and verifying PVC bound status
STEP: building a pvc
STEP: creating above pvc
STEP: verifying pvc status as bound
STEP: verifying ZFSVolume object
STEP: fetching zfs volume
STEP: verifying zfs volume
STEP: verifying ZFSVolume property change
STEP: verifying compression property update
STEP: fetching zfs volume for setting compression=on
STEP: fetching zfs volume for setting compression=off
STEP: verifying dedup property update
STEP: fetching zfs volume for setting dedup=on
STEP: fetching zfs volume for setting dedup=off
STEP: verifying blocksize property update
STEP: fetching zfs volume for setting the blocksize
STEP: Deleting pvc
STEP: verifying deleted pvc
STEP: Deleting storage class

• [SLOW TEST:210.965 seconds]
[zfspv] TEST VOLUME PROVISIONING
/home/filippo/go/src/github.com/openebs/zfs-localpv/tests/provision_test.go:23
  App is deployed with zfs driver
  /home/filippo/go/src/github.com/openebs/zfs-localpv/tests/provision_test.go:24
    Running zfs volume Creation Test
    /home/filippo/go/src/github.com/openebs/zfs-localpv/tests/provision_test.go:25
------------------------------

Ran 1 of 1 Specs in 210.966 seconds
SUCCESS! -- 1 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS
```
