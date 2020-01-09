### Setup helm

This step uses helm the kubernetes package manager. If you not setup the helm then do the below the configuration, otherwise move to next step.

```
$ helm version
Client: &version.Version{SemVer:"v2.16.1", GitCommit:"bbdfe5e7803a12bbdf97e94cd847859890cf4050", GitTreeState:"clean"}
Server: &version.Version{SemVer:"v2.16.1", GitCommit:"bbdfe5e7803a12bbdf97e94cd847859890cf4050", GitTreeState:"clean"}

$ helm init
Tiller (the Helm server-side component) has been installed into your Kubernetes Cluster.

Please note: by default, Tiller is deployed with an insecure 'allow unauthenticated users' policy.
To prevent this, run `helm init` with the --tiller-tls-verify flag.
For more information on securing your installation see: https://docs.helm.sh/using_helm/#securing-your-helm-installation

$ kubectl create serviceaccount --namespace kube-system tiller
serviceaccount/tiller created

$ kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
clusterrolebinding.rbac.authorization.k8s.io/tiller-cluster-rule created

$ kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'
deployment.extensions/tiller-deploy patched
```

### Install Prometheus Operator

Once the helm is ready and related titler pods is up and running , use the Prometheus chart from the helm repository

```
$ helm install stable/prometheus-operator --name prometheus-operator
NAME:   prometheus-operator
LAST DEPLOYED: Thu Jan  9 12:50:03 2020
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1/Alertmanager
NAME                              AGE
prometheus-operator-alertmanager  54s

==> v1/ClusterRole
NAME                                              AGE
prometheus-operator-alertmanager                  54s
prometheus-operator-grafana-clusterrole           54s
prometheus-operator-operator                      54s
prometheus-operator-operator-psp                  54s
prometheus-operator-prometheus                    54s
prometheus-operator-prometheus-psp                54s
psp-prometheus-operator-kube-state-metrics        54s
psp-prometheus-operator-prometheus-node-exporter  54s

==> v1/ClusterRoleBinding
NAME                                              AGE
prometheus-operator-alertmanager                  54s
prometheus-operator-grafana-clusterrolebinding    54s
prometheus-operator-operator                      54s
prometheus-operator-operator-psp                  54s
prometheus-operator-prometheus                    54s
prometheus-operator-prometheus-psp                54s
psp-prometheus-operator-kube-state-metrics        54s
psp-prometheus-operator-prometheus-node-exporter  54s

==> v1/ConfigMap
NAME                                                   AGE
prometheus-operator-apiserver                          54s
prometheus-operator-cluster-total                      54s
prometheus-operator-controller-manager                 54s
prometheus-operator-etcd                               54s
prometheus-operator-grafana                            54s
prometheus-operator-grafana-config-dashboards          54s
prometheus-operator-grafana-datasource                 54s
prometheus-operator-grafana-test                       54s
prometheus-operator-k8s-resources-cluster              54s
prometheus-operator-k8s-resources-namespace            54s
prometheus-operator-k8s-resources-node                 54s
prometheus-operator-k8s-resources-pod                  54s
prometheus-operator-k8s-resources-workload             54s
prometheus-operator-k8s-resources-workloads-namespace  54s
prometheus-operator-kubelet                            54s
prometheus-operator-namespace-by-pod                   54s
prometheus-operator-namespace-by-workload              54s
prometheus-operator-node-cluster-rsrc-use              54s
prometheus-operator-node-rsrc-use                      54s
prometheus-operator-nodes                              54s
prometheus-operator-persistentvolumesusage             54s
prometheus-operator-pod-total                          54s
prometheus-operator-pods                               54s
prometheus-operator-prometheus                         54s
prometheus-operator-proxy                              54s
prometheus-operator-scheduler                          54s
prometheus-operator-statefulset                        54s
prometheus-operator-workload-total                     54s

==> v1/DaemonSet
NAME                                          AGE
prometheus-operator-prometheus-node-exporter  54s

==> v1/Deployment
NAME                                    AGE
prometheus-operator-grafana             54s
prometheus-operator-kube-state-metrics  54s
prometheus-operator-operator            54s

==> v1/Pod(related)
NAME                                                     AGE
prometheus-operator-grafana-85bb5d49d-bffdg              54s
prometheus-operator-kube-state-metrics-5d46566c59-p8k6s  54s
prometheus-operator-operator-64844759f7-rpwws            54s
prometheus-operator-prometheus-node-exporter-p9rl8       54s

==> v1/Prometheus
NAME                            AGE
prometheus-operator-prometheus  54s

==> v1/PrometheusRule
NAME                                                      AGE
prometheus-operator-alertmanager.rules                    54s
prometheus-operator-etcd                                  54s
prometheus-operator-general.rules                         54s
prometheus-operator-k8s.rules                             54s
prometheus-operator-kube-apiserver-error                  54s
prometheus-operator-kube-apiserver.rules                  54s
prometheus-operator-kube-prometheus-node-recording.rules  54s
prometheus-operator-kube-scheduler.rules                  54s
prometheus-operator-kubernetes-absent                     54s
prometheus-operator-kubernetes-apps                       54s
prometheus-operator-kubernetes-resources                  54s
prometheus-operator-kubernetes-storage                    54s
prometheus-operator-kubernetes-system                     54s
prometheus-operator-kubernetes-system-apiserver           54s
prometheus-operator-kubernetes-system-controller-manager  54s
prometheus-operator-kubernetes-system-kubelet             54s
prometheus-operator-kubernetes-system-scheduler           54s
prometheus-operator-node-exporter                         54s
prometheus-operator-node-exporter.rules                   54s
prometheus-operator-node-network                          54s
prometheus-operator-node-time                             54s
prometheus-operator-node.rules                            54s
prometheus-operator-prometheus                            54s
prometheus-operator-prometheus-operator                   54s

==> v1/Role
NAME                              AGE
prometheus-operator-grafana-test  54s

==> v1/RoleBinding
NAME                              AGE
prometheus-operator-grafana-test  54s

==> v1/Secret
NAME                                           AGE
alertmanager-prometheus-operator-alertmanager  54s
prometheus-operator-grafana                    54s

==> v1/Service
NAME                                          AGE
prometheus-operator-alertmanager              54s
prometheus-operator-coredns                   54s
prometheus-operator-grafana                   54s
prometheus-operator-kube-controller-manager   54s
prometheus-operator-kube-etcd                 54s
prometheus-operator-kube-proxy                54s
prometheus-operator-kube-scheduler            54s
prometheus-operator-kube-state-metrics        54s
prometheus-operator-operator                  54s
prometheus-operator-prometheus                54s
prometheus-operator-prometheus-node-exporter  54s

==> v1/ServiceAccount
NAME                                          AGE
prometheus-operator-alertmanager              54s
prometheus-operator-grafana                   54s
prometheus-operator-grafana-test              54s
prometheus-operator-kube-state-metrics        54s
prometheus-operator-operator                  54s
prometheus-operator-prometheus                54s
prometheus-operator-prometheus-node-exporter  54s

==> v1/ServiceMonitor
NAME                                         AGE
prometheus-operator-alertmanager             53s
prometheus-operator-apiserver                53s
prometheus-operator-coredns                  53s
prometheus-operator-grafana                  53s
prometheus-operator-kube-controller-manager  53s
prometheus-operator-kube-etcd                53s
prometheus-operator-kube-proxy               53s
prometheus-operator-kube-scheduler           53s
prometheus-operator-kube-state-metrics       53s
prometheus-operator-kubelet                  53s
prometheus-operator-node-exporter            53s
prometheus-operator-operator                 53s
prometheus-operator-prometheus               53s

==> v1beta1/ClusterRole
NAME                                    AGE
prometheus-operator-kube-state-metrics  54s

==> v1beta1/ClusterRoleBinding
NAME                                    AGE
prometheus-operator-kube-state-metrics  54s

==> v1beta1/MutatingWebhookConfiguration
NAME                           AGE
prometheus-operator-admission  54s

==> v1beta1/PodSecurityPolicy
NAME                                          AGE
prometheus-operator-alertmanager              54s
prometheus-operator-grafana                   54s
prometheus-operator-grafana-test              54s
prometheus-operator-kube-state-metrics        54s
prometheus-operator-operator                  54s
prometheus-operator-prometheus                54s
prometheus-operator-prometheus-node-exporter  54s

==> v1beta1/Role
NAME                         AGE
prometheus-operator-grafana  54s

==> v1beta1/RoleBinding
NAME                         AGE
prometheus-operator-grafana  54s

==> v1beta1/ValidatingWebhookConfiguration
NAME                           AGE
prometheus-operator-admission  53s


NOTES:
The Prometheus Operator has been installed. Check its status by running:
  kubectl --namespace default get pods -l "release=prometheus-operator"

  Visit https://github.com/coreos/prometheus-operator for instructions on how
  to create & configure Alertmanager and Prometheus instances using the Operator.
```

Lookup all the required pods are up and running

```
$ kubectl get pods -l "release=prometheus-operator"
NAME                                                 READY   STATUS    RESTARTS   AGE
prometheus-operator-grafana-85bb5d49d-bffdg          2/2     Running   0          2m21s
prometheus-operator-operator-64844759f7-rpwws        2/2     Running   0          2m21s
prometheus-operator-prometheus-node-exporter-p9rl8   1/1     Running   0          2m21s
```

### Setup alert rule

Please check the rules there in the system :-

```
$ kubectl get PrometheusRule
NAME                                                       AGE
prometheus-operator-alertmanager.rules                     4m21s
prometheus-operator-etcd                                   4m21s
prometheus-operator-general.rules                          4m21s
prometheus-operator-k8s.rules                              4m21s
prometheus-operator-kube-apiserver-error                   4m21s
prometheus-operator-kube-apiserver.rules                   4m21s
prometheus-operator-kube-prometheus-node-recording.rules   4m21s
prometheus-operator-kube-scheduler.rules                   4m21s
prometheus-operator-kubernetes-absent                      4m21s
prometheus-operator-kubernetes-apps                        4m21s
prometheus-operator-kubernetes-resources                   4m21s
prometheus-operator-kubernetes-storage                     4m21s
prometheus-operator-kubernetes-system                      4m21s
prometheus-operator-kubernetes-system-apiserver            4m21s
prometheus-operator-kubernetes-system-controller-manager   4m21s
prometheus-operator-kubernetes-system-kubelet              4m21s
prometheus-operator-kubernetes-system-scheduler            4m21s
prometheus-operator-node-exporter                          4m21s
prometheus-operator-node-exporter.rules                    4m21s
prometheus-operator-node-network                           4m21s
prometheus-operator-node-time                              4m21s
prometheus-operator-node.rules                             4m21s
prometheus-operator-prometheus                             4m21s
prometheus-operator-prometheus-operator                    4m21s
```

You can edit any of the default rule or setup the new rule to get the alerts. Here is the sample alert rule if available storage space is less than 10% then start throwing the alert :-

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  labels:
    app: prometheus-operator
    chart: prometheus-operator-8.5.4
    heritage: Tiller
    release: prometheus-operator
  name: prometheus-operator-zfs-alertmanager.rules
  namespace: default
spec:
  groups:
  - name: zfsalertmanager.rules
    rules:
    - alert: ZFSVolumeUsageCritical
      annotations:
        message: The PersistentVolume claimed by {{ $labels.persistentvolumeclaim
          }} in Namespace {{ $labels.namespace }} is only {{ printf "%0.2f" $value
          }}% free.
      expr: |
        100 * kubelet_volume_stats_available_bytes{job="kubelet"}
          /
        kubelet_volume_stats_capacity_bytes{job="kubelet"}
          < 10
      for: 1m
      labels:
        severity: critical
```

Apply the above yaml so that Prometheus can fire the alert when available space is less than 10%


### Check the Prometheus alert

To be able to view the Prometheus web UI, expose it through a Service. A simple way to do this is to use a Service of type NodePort

```
$ cat prometheus-service.yaml

apiVersion: v1
kind: Service
metadata:
  name: prometheus-service
spec:
  type: NodePort
  ports:
  - name: web
    nodePort: 30090
    port: 9090
    protocol: TCP
    targetPort: web
  selector:
    prometheus: prometheus-operator-prometheus
```

apply the above yaml

```
$ kubectl apply -f prometheus-service.yaml
service/prometheus-service created
```

Now you can access the alert manager UI via "nodes-external-ip:30090"

```
$ kubectl get nodes -owide
NAME                                         STATUS   ROLES    AGE    VERSION          INTERNAL-IP   EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION   CONTAINER-RUNTIME
gke-zfspv-pawan-default-pool-3e407350-xvzp   Ready    <none>   103m   v1.15.4-gke.22   10.168.0.45   34.94.3.140   Ubuntu 18.04.3 LTS   5.0.0-1022-gke   docker://19.3.2
```

In this case we can access the alert manager via url http://34.94.3.140:30090/

### Check the Alert Manager

To be able to view the Alert Manager web UI, expose it through a Service of type NodePort

```
$ cat alertmanager-service.yaml

apiVersion: v1
kind: Service
metadata:
  name: alertmanager-service
spec:
  type: NodePort
  ports:
  - name: web
    nodePort: 30093
    port: 9093
    protocol: TCP
    targetPort: web
  selector:
    alertmanager: prometheus-operator-alertmanager
```

apply the above yaml

```
$ kubectl apply -f alertmanager-service.yaml
service/alertmanager-service created
```

Now you can access the alert manager UI via "nodes-external-ip:30093"

```
$ kubectl get nodes -owide
NAME                                         STATUS   ROLES    AGE    VERSION          INTERNAL-IP   EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION   CONTAINER-RUNTIME
gke-zfspv-pawan-default-pool-3e407350-xvzp   Ready    <none>   103m   v1.15.4-gke.22   10.168.0.45   34.94.3.140   Ubuntu 18.04.3 LTS   5.0.0-1022-gke   docker://19.3.2
```

In this case we can access the alert manager via url http://34.94.3.140:30093/
