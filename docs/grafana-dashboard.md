# Install Prometheus Operator via Helm 3.2.1



## Setup Helm 3

### Install Helm 3
```
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod +x get_helm.sh
./get_helm.sh
```

### Add Stable Charts
```
helm repo add stable https://kubernetes-charts.storage.googleapis.com/
```

## Setup Prometheus and Grafana via Helm

### Install Prometheus Operator

In my case, the Kubernetes cluster is behind a firewall. I am configuring the Granfana to be accessible via NodePort, as I need to access the Grafana UI using ssh tunnel.

```
kubectl create namespace prometheus-operator
helm install prometheus-operator stable/prometheus-operator -n prometheus-operator --set prometheusOperator.createCustomResource=false,grafana.service.type=NodePort
```

Note: Prometheus Operator (by default) installs Grafana and adds Prometheus as the data source.

### Verify

```
kubectl get pods -n prometheus-operator
```

The above commands should show that all promtheus operator, prometheus, node exporter and grafana pods are running.

```
NAME                                                     READY   STATUS    RESTARTS   AGE
alertmanager-prometheus-operator-alertmanager-0          2/2     Running   0          30m
prometheus-operator-grafana-cf6954699-5rcgl              2/2     Running   0          30m
prometheus-operator-kube-state-metrics-5fdcd78bc-sckjv   1/1     Running   0          30m
prometheus-operator-operator-5dd8f8f568-52qk8            2/2     Running   0          30m
prometheus-operator-prometheus-node-exporter-p8pm8       1/1     Running   0          30m
prometheus-operator-prometheus-node-exporter-trlhp       1/1     Running   0          30m
prometheus-operator-prometheus-node-exporter-wsm4n       1/1     Running   0          30m
prometheus-prometheus-operator-prometheus-0              3/3     Running   1          30m
```

```
kubectl get svc -n prometheus-operator
```

Note that Grafana alone is running on NodePort

```
NAME                                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
alertmanager-operated                          ClusterIP   None            <none>        9093/TCP,9094/TCP,9094/UDP   31m
prometheus-operated                            ClusterIP   None            <none>        9090/TCP                     30m
prometheus-operator-alertmanager               ClusterIP   10.102.104.48   <none>        9093/TCP                     31m
prometheus-operator-grafana                    NodePort    10.96.160.172   <none>        80:31409/TCP                 31m
prometheus-operator-kube-state-metrics         ClusterIP   10.105.92.154   <none>        8080/TCP                     31m
prometheus-operator-operator                   ClusterIP   10.99.15.245    <none>        8080/TCP,443/TCP             31m
prometheus-operator-prometheus                 ClusterIP   10.109.75.138   <none>        9090/TCP                     31m
prometheus-operator-prometheus-node-exporter   ClusterIP   10.98.128.115   <none>        9100/TCP                     31m
```

## Accessing Grafana UI over SSH Tunnel

### Windows using PuTTY

- Get the Kubernetes Worker Node IP and the Grafana Node Port.
- Get the SSH server using which, Kubernetes Worker Node IP is accessible. Say this is Landing IP.
- Configure the PuTTY as follows:
  - Create a new Session with Landing IP, Landing Port
  - Create a Connection -> SSH -> Tunnels
    - Source Port = Grafana NodePort
    - Destination = Kubernetes Worker Node IP:Grafana Node Port
  - Open the PuTTY session. Enter SSH user name and passowrd for the Landing IP.
- Now you can access Grafana UI at http://localhost:<Grafana-Node-Port>/. Default login and password ( admin/prom-operator )


### Linux using SSH

- Get the Kubernetes Worker Node IP and the Grafana Node Port.
- Get the SSH server using which, Kubernetes Worker Node IP is accessible. Say this is Landing IP.
- Open SSH tunnel using command.
  ```
  ssh -NL <Grafana-Node-Port>:<k8s-worker-node-IP>:<Grafana-Node-Port> <landing-machine-user>@<landing-machine-ip> -p <landing-machine-ssh-port>
  ```
- Now you can access Grafana UI at http://localhost:<Grafana-Node-Port>/. Default login and password ( admin/prom-operator )

## Verify Granafa Dashboard

- Login to Granfa UI
- Click on Settings -> Data Source. You must see a Default Prometheus data source for `http://prometheus-operator-prometheus:9090/`
- Click on Dashboards -> Manage Dashboards. You must see a list of dashboards. Click on any of them like: `kubernetes-compute-resources-cluster`

## Add ZFS-LocalPV Grafana Dashboard

- Login to Granfa UI
- Click on Create Dashboard -> Import dashboard
- Paste the below json and Click on Load

  ```
  https://raw.githubusercontent.com/openebs/zfs-localpv/master/deploy/sample/grafana-dashboard.json
  ```
- Select datasource as Prometheus and Import it.

- Now you should see a dashboard with name as "ZFS-LocalPV"

This dashboard exposes below metrics

- Volume Capacity (Used space percentage)
- ZPOOL Read/Write time
- ZPOOL Read/Write IOs
- ARC Size, Hits, Misses
- L2ARC Size, Hits, Misses

The "ZFS-LocalPV" dashboard will look like this :-

![Grafana](https://github.com/openebs/zfs-localpv/blob/master/deploy/sample/vol-stats.png)

## References:
- https://helm.sh/docs/intro/install/
- https://github.com/helm/charts/issues/19452
- https://www.linode.com/docs/networking/ssh/using-ssh-on-windows/#ssh-tunneling-port-forwarding
- https://github.com/helm/charts/tree/master/stable/prometheus-operator
- https://github.com/helm/charts/tree/master/stable/grafana#configuration
