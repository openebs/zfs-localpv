### Prerequisite

We need to have k8s version 1.15+ to access the CSI volume metrics.

### Setup helm

This step uses helm the kubernetes package manager. If you have not setup the helm then do the below configuration, otherwise move to the next step.

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

Once the helm is ready and related tiller pods is up and running , use the Prometheus chart from the helm repository

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

Check all the required pods are up and running

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

You can edit any of the default rules or setup the new rule to get the alerts. Here is the sample rule to start firing the alerts if available storage space is less than 10% :-

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

### ZFS metrics exposed by Prometheus

We can create the rule for ZFS metrics also. Here is the list of ZFS metrics exposed by prometheus :-

```
node_zfs_abd_linear_cnt
node_zfs_abd_linear_data_size
node_zfs_abd_scatter_chunk_waste
node_zfs_abd_scatter_cnt
node_zfs_abd_scatter_data_size
node_zfs_abd_scatter_order_0
node_zfs_abd_scatter_order_1
node_zfs_abd_scatter_order_10
node_zfs_abd_scatter_order_2
node_zfs_abd_scatter_order_3
node_zfs_abd_scatter_order_4
node_zfs_abd_scatter_order_5
node_zfs_abd_scatter_order_6
node_zfs_abd_scatter_order_7
node_zfs_abd_scatter_order_8
node_zfs_abd_scatter_order_9
node_zfs_abd_scatter_page_alloc_retry
node_zfs_abd_scatter_page_multi_chunk
node_zfs_abd_scatter_page_multi_zone
node_zfs_abd_scatter_sg_table_retry
node_zfs_abd_struct_size
node_zfs_arc_access_skip
node_zfs_arc_anon_evictable_data
node_zfs_arc_anon_evictable_metadata
node_zfs_arc_anon_size
node_zfs_arc_arc_dnode_limit
node_zfs_arc_arc_loaned_bytes
node_zfs_arc_arc_meta_limit
node_zfs_arc_arc_meta_max
node_zfs_arc_arc_meta_min
node_zfs_arc_arc_meta_used
node_zfs_arc_arc_need_free
node_zfs_arc_arc_no_grow
node_zfs_arc_arc_prune
node_zfs_arc_arc_sys_free
node_zfs_arc_arc_tempreserve
node_zfs_arc_bonus_size
node_zfs_arc_c
node_zfs_arc_c_max
node_zfs_arc_c_min
node_zfs_arc_compressed_size
node_zfs_arc_data_size
node_zfs_arc_dbuf_size
node_zfs_arc_deleted
node_zfs_arc_demand_data_hits
node_zfs_arc_demand_data_misses
node_zfs_arc_demand_hit_predictive_prefetch
node_zfs_arc_demand_metadata_hits
node_zfs_arc_demand_metadata_misses
node_zfs_arc_dnode_size
node_zfs_arc_evict_l2_cached
node_zfs_arc_evict_l2_eligible
node_zfs_arc_evict_l2_ineligible
node_zfs_arc_evict_l2_skip
node_zfs_arc_evict_not_enough
node_zfs_arc_evict_skip
node_zfs_arc_hash_chain_max
node_zfs_arc_hash_chains
node_zfs_arc_hash_collisions
node_zfs_arc_hash_elements
node_zfs_arc_hash_elements_max
node_zfs_arc_hdr_size
node_zfs_arc_hits
node_zfs_arc_l2_abort_lowmem
node_zfs_arc_l2_asize
node_zfs_arc_l2_cksum_bad
node_zfs_arc_l2_evict_l1cached
node_zfs_arc_l2_evict_lock_retry
node_zfs_arc_l2_evict_reading
node_zfs_arc_l2_feeds
node_zfs_arc_l2_free_on_write
node_zfs_arc_l2_hdr_size
node_zfs_arc_l2_hits
node_zfs_arc_l2_io_error
node_zfs_arc_l2_misses
node_zfs_arc_l2_read_bytes
node_zfs_arc_l2_rw_clash
node_zfs_arc_l2_size
node_zfs_arc_l2_write_bytes
node_zfs_arc_l2_writes_done
node_zfs_arc_l2_writes_error
node_zfs_arc_l2_writes_lock_retry
node_zfs_arc_l2_writes_sent
node_zfs_arc_memory_all_bytes
node_zfs_arc_memory_direct_count
node_zfs_arc_memory_free_bytes
node_zfs_arc_memory_indirect_count
node_zfs_arc_memory_throttle_count
node_zfs_arc_metadata_size
node_zfs_arc_mfu_evictable_data
node_zfs_arc_mfu_evictable_metadata
node_zfs_arc_mfu_ghost_evictable_data
node_zfs_arc_mfu_ghost_evictable_metadata
node_zfs_arc_mfu_ghost_hits
node_zfs_arc_mfu_ghost_size
node_zfs_arc_mfu_hits
node_zfs_arc_mfu_size
node_zfs_arc_misses
node_zfs_arc_mru_evictable_data
node_zfs_arc_mru_evictable_metadata
node_zfs_arc_mru_ghost_evictable_data
node_zfs_arc_mru_ghost_evictable_metadata
node_zfs_arc_mru_ghost_hits
node_zfs_arc_mru_ghost_size
node_zfs_arc_mru_hits
node_zfs_arc_mru_size
node_zfs_arc_mutex_miss
node_zfs_arc_overhead_size
node_zfs_arc_p
node_zfs_arc_prefetch_data_hits
node_zfs_arc_prefetch_data_misses
node_zfs_arc_prefetch_metadata_hits
node_zfs_arc_prefetch_metadata_misses
node_zfs_arc_size
node_zfs_arc_sync_wait_for_async
node_zfs_arc_uncompressed_size
node_zfs_dmu_tx_dmu_tx_assigned
node_zfs_dmu_tx_dmu_tx_delay
node_zfs_dmu_tx_dmu_tx_dirty_delay
node_zfs_dmu_tx_dmu_tx_dirty_over_max
node_zfs_dmu_tx_dmu_tx_dirty_throttle
node_zfs_dmu_tx_dmu_tx_error
node_zfs_dmu_tx_dmu_tx_group
node_zfs_dmu_tx_dmu_tx_memory_reclaim
node_zfs_dmu_tx_dmu_tx_memory_reserve
node_zfs_dmu_tx_dmu_tx_quota
node_zfs_dmu_tx_dmu_tx_suspended
node_zfs_dnode_dnode_alloc_next_block
node_zfs_dnode_dnode_alloc_next_chunk
node_zfs_dnode_dnode_alloc_race
node_zfs_dnode_dnode_allocate
node_zfs_dnode_dnode_buf_evict
node_zfs_dnode_dnode_free_interior_lock_retry
node_zfs_dnode_dnode_hold_alloc_hits
node_zfs_dnode_dnode_hold_alloc_interior
node_zfs_dnode_dnode_hold_alloc_lock_misses
node_zfs_dnode_dnode_hold_alloc_lock_retry
node_zfs_dnode_dnode_hold_alloc_misses
node_zfs_dnode_dnode_hold_alloc_type_none
node_zfs_dnode_dnode_hold_dbuf_hold
node_zfs_dnode_dnode_hold_dbuf_read
node_zfs_dnode_dnode_hold_free_hits
node_zfs_dnode_dnode_hold_free_lock_misses
node_zfs_dnode_dnode_hold_free_lock_retry
node_zfs_dnode_dnode_hold_free_misses
node_zfs_dnode_dnode_hold_free_overflow
node_zfs_dnode_dnode_hold_free_refcount
node_zfs_dnode_dnode_hold_free_txg
node_zfs_dnode_dnode_move_active
node_zfs_dnode_dnode_move_handle
node_zfs_dnode_dnode_move_invalid
node_zfs_dnode_dnode_move_recheck1
node_zfs_dnode_dnode_move_recheck2
node_zfs_dnode_dnode_move_rwlock
node_zfs_dnode_dnode_move_special
node_zfs_dnode_dnode_reallocate
node_zfs_fm_erpt_dropped
node_zfs_fm_erpt_set_failed
node_zfs_fm_fmri_set_failed
node_zfs_fm_payload_set_failed
node_zfs_vdev_cache_delegations
node_zfs_vdev_cache_hits
node_zfs_vdev_cache_misses
node_zfs_xuio_onloan_read_buf
node_zfs_xuio_onloan_write_buf
node_zfs_xuio_read_buf_copied
node_zfs_xuio_read_buf_nocopy
node_zfs_xuio_write_buf_copied
node_zfs_xuio_write_buf_nocopy
node_zfs_zfetch_hits
node_zfs_zfetch_max_streams
node_zfs_zfetch_misses
node_zfs_zil_zil_commit_count
node_zfs_zil_zil_commit_writer_count
node_zfs_zil_zil_itx_copied_bytes
node_zfs_zil_zil_itx_copied_count
node_zfs_zil_zil_itx_count
node_zfs_zil_zil_itx_indirect_bytes
node_zfs_zil_zil_itx_indirect_count
node_zfs_zil_zil_itx_metaslab_normal_bytes
node_zfs_zil_zil_itx_metaslab_normal_count
node_zfs_zil_zil_itx_metaslab_slog_bytes
node_zfs_zil_zil_itx_metaslab_slog_count
node_zfs_zil_zil_itx_needcopy_bytes
node_zfs_zil_zil_itx_needcopy_count
node_zfs_zpool_nread
node_zfs_zpool_nwritten
node_zfs_zpool_rcnt
node_zfs_zpool_reads
node_zfs_zpool_rlentime
node_zfs_zpool_rtime
node_zfs_zpool_rupdate
node_zfs_zpool_wcnt
node_zfs_zpool_wlentime
node_zfs_zpool_writes
node_zfs_zpool_wtime
node_zfs_zpool_wupdate
```
