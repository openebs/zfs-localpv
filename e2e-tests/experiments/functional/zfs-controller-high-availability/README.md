## About this experiment

This functional experiment scale up the zfs-controller statefulset replicas to use it in high availability mode and then verify the zfs-localpv behaviour when one of the replicas go down. This experiment checks the initial number of replicas of zfs-controller statefulset and scale it by one if a free node is present which should be able to schedule the pods. Default value for zfs-controller statefulset replica is one.

## Supported platforms:

K8s : 1.18+

OS : Ubuntu, CentOS

ZFS : 0.7, 0.8

## Entry-Criteria

- k8s cluster should be in healthy state including all desired worker nodes in ready state.
- zfs-localpv driver should be deployed and zfs-controller and csi node-agent daemonset pods should be in running state.
- one spare schedulable node should be present in the cluster so that after scaling up the zfs-controller replica by one, new replica gets scheduled on that node. These replicas will follow the anti-affinity rules so that replica pods will be present on different nodes only.

## Exit-Criteria

- zfs-controller statefulset should be scaled up by one replica.
- All the replias should be in running state.
- zfs-localpv volumes should be healthy and data after scaling up controller should not be impacted.
- This experiment makes one of the zfs-controller statefulset replica to go down, as a result active/master replica of zfs-controller prior to the experiment will be changed to some other remaining replica after the experiment completes. This happens because of the lease mechanism, which is being used to decide which replica will be serving as master. At a time only one replica will be master.
- Volumes provisioning / deprovisioning should not be impacted if any one replica goes down.

## Steps performed

- Get the no of zfs-controller statefulset replica count.
- Scale down the controller replicas to zero, wait until controller pods gets terminated successfully and then try to provision a volume to use by busybox application.
- Due to zero active replicas of zfs-controller, pvc should remain in Pending state.
- If no. of schedulable nodes are greater or equal to the previous replica count + 1, then zfs-controller will be scaled up by +1 replica. Doing this will Bound the pvc and application pod will come in Running state.
- Now taint all the nodes with `NoSchedule` so that when we delete the master replica of zfs-controller it doesn't come back to running state and at that time lease should be given to some other replica and now that replica will work as master.
- Now deprovision the application. This time deprovision will be done by that master replica which is active at present. So here we validated that provision and deprovisioning was successully done by two different replica of zfs-controller. And remove the taints before exiting the test execution. And then we check running statue of all the replicas and csi-node agent pods.
- If no. of schedulable nodes are not present for scheduling updated no. of replicas then this test will fail at the task of scaling up replicas and then it will skip further tasks. Before exiting it will scale up the down replicas with same no of replica count which was present at starting of this experiment. Doing this will Bound the pvc and application pod will come in running state. This test execution will end after deleting that pvc and application pod.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of zfs-localpv controller high availability, clone openens/zfs-localpv[https://github.com/openebs/zfs-localpv] repo and then first apply rbac and crds for e2e-framework.
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
kubectl logs -f <zfs-controller-high-availability-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er zfs-controller-high-availability -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er zfs-controller-high-availability -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```