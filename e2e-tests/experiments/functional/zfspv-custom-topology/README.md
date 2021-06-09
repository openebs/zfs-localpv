## About the experiment

- After zfs-driver:v0.7.x user can label the nodes with the required topology, the zfs-localpv driver will support all the node labels as topology keys. This experiment verifies this custom-topology support for zfs-localpv. Volume should be provisioned on only such nodes which have been labeled with the keys set via the storage-class.
- In this experiment we cover two scenarios such as one with immediate volume binding and other with late binding (i.e. WaitForFirstConsumer). If we add a label to node after zfs-localpv driver deployment and using late binding mode, then a restart of all the node agents are required so that the driver can pick the labels and add them as supported topology key. Restart is not required in case of immediate volumebinding irrespective of if we add labels after zfs-driver deployment or before.

## Supported platforms:
K8s : 1.18+

OS : Ubuntu, CentOS

ZFS : 0.7, 0.8

ZFS-LocalPV version: 0.7+

## Entry-Criteria

- K8s cluster should be in healthy state including all desired nodes in ready state.
- zfs-controller and node-agent daemonset pods should be in running state.

## Steps performed

- select any of the two nodes randomly from the k8s cluster and label them with some key.
- deploy five applications using the pvc, provisioned by storage class in which volume binding mode is immediate.
- verify that pvc is bound and application pod is in running state.
- verify that volume is provisioned on only those nodes which was labeled prior to the provisioning.
- after that deploy five more applications, using the pvc provisioned by storage class in which volume binding mode is waitforfirstconsumer.
- check that pvc remains in pending state.
- restart the csi node-agent pods on all nodes.
- verify that new topology keys are now present in csi-nodes.
- now pvc should come into Bound state and application should be in running state.
- verify that volume is provisioned on only those nodes which was labeled.
- At end of test, remove the node labels and restart csi nodes so that custom-labels will be removed from csi node.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of zfspv custom topology, first clone openens/zfs-localpv[https://github.com/openebs/zfs-localpv] repo and then apply rbac and crds for e2e-framework.
```
kubectl apply -f zfs-localpv/e2e-tests/hack/rbac.yaml
kubectl apply -f zfs-localpv/e2e-tests/hack/crds.yaml
```
then update the needed test specific values in run_e2e_test.yml file and create the kubernetes job.
```
kubectl create -f run_e2e_test.yml
```
All the env variables description is provided with the comments in the same file.
After creating kubernetes job, when the job’s pod is instantiated, we can see the logs of that pod which is executing the test-case.

```
kubectl get pods -n e2e
kubectl logs -f <zfspv-custom-topology-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er zfspv-custom-topology -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er zfspv-custom-topology -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```