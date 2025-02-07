## About this experiment

This experiment validates the stability and fault-tolerance of application pod consuming zfs-localpv storage. In this test chaos is induced on the container of application pod using pumba chaos utils. Basically it is used as a disruptive test, to cause loss of access to storage by failing the application pod and later it tests the recovery workflow of the application pod.

## Supported platforms:

K8s : 1.18+

OS : Ubuntu, CentOS

ZFS : 0.7, 0.8

## Entry-Criteria

- One application should be deployed consuming zfs-localpv storage.
- Application services are accessible & pods are healthy
- Application writes are successful
- zfs-controller and csi node-agent daemonset pods should be in running state.

## Exit-Criteria

- Application services are accessible & pods are healthy
- Data written prior to chaos is successfully retrieved/read
- Data consistency is maintained as per integrity check utils
- Storage target pods are healthy

## Steps performed

- Get the application pod name and check its Running status
- Dump some dummy data into the application mount point to check data consistency after chaos injection.
- Create a daemonset of pumba utils and get the name of the pod scheduled on the same node as of application pod. Utils used in this test is located at `e2e-tests/chaoslib/pumba` directory.
- Now using SIGKILL command via pumba pod disrupt the access of application container to the storage. And now in recovery process container restarts.
- Check the container restart count to validate successful chaos injection.
- Validate the data consistency by checking the md5sum of test data.
- Delete the pumba daemonset.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of application pod failure, clone openens/zfs-localpv[https://github.com/openebs/zfs-localpv] repo and then first apply rbac and crds for e2e-framework.
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
kubectl logs -f <application-pod-failure-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er application-pod-failure -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er application-pod-failure -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```