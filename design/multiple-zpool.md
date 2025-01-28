---
title: Multiple ZPool Support for ZFS  Local PV 
authors:
  - "@pawanpraka1"
creation-date: 2022-06-16
last-updated: 2022-06-16
---

# Multiple ZPool Support for ZFS  Local PV

## Table of Contents

* [Table of Contents](#table-of-contents)
* [Summary](#summary)
* [Design Constraints/Assumptions](#design-constraintsassumptions)
* [Proposal](#proposal)
    * [User Stories](#user-stories)
      * [Create Zpool](#create-zpool)
    * [Implementation Details](#implementation-details)
      * [Setup Storageclass](#setup-storageclass)
      * [Zpool Selection Workflow](#volume-selection-workflow)
    * [High Level Design](#high-level-design)
* [Implementation Plan](#implementation-plan)

## Summary

This is a design proposal to support multiple Zpools in a single stoorage class for ZFS Local PV. This design describes how we can provide multiple ZFS pools in a storageclass, how the ZFS Local PV driver will pick the ZFS Pool. 

Using the design/solution described in this document, users will be able to provide multiple ZFS Pools in single storage class.

## Design Constraints/Assumptions

- Ubuntu 18.04
- Kubernetes 1.14+
- Node are installed with ZFS 0.7 or 0.8
- ZPOOLs are pre-created by the administrator on the nodes. Zpools on all the nodes will have the same name.
- StorageClass Topology specification will be used to restrict the Volumes to be provisioned on the nodes where the ZPOOLs are available.

## Proposal

### User Stories

#### Create Zpool
I should be able to provide a multiple zpools in a storageclass that can be used to provision the volume. This volume should get created dynamically during application creation time and the provision should happen from the ZFS pools which is mentioned in the storageclass.

### Implementation Details

#### Setup Storageclass

User can create the storageclass with the poolpattern paramaeter. In order to provide multiple Zpools, users can use regular expression in the poolpattern paramater.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: openebs-zfspv
allowVolumeExpansion: true
parameters:
 poolpattern: "zfspv-pool*"
provisioner: zfs.csi.openebs.io
```


#### Zpool Selection Workflow

- CSI driver will handle CSI request for volume create
- CSI driver will read the request parameters and create ZFSVolume resources:
- ZFSVolume will be watched by the LocalPV-ZFS node agent and will check the poolpattern parameter.
- The node agent will find all the zpool matching the poolpattern parameter regx and will pick the one which has highest free space available.
- now the node agent will go ahead and create the volume in the selected zfs pool
- the node agent will also update the poolname parameter with the selected poolname

#### High Level Design
- user will setup all the node and setup the ZFS pool on each of those nodes.
- user will deploy below sample storage class where we get all the needed zfs properties for creating the volume. The storage class will have poolpattern parameter which will help us pick the correcponding zfs pool matching the poolpattern 

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: openebs-zfspv
provisioner: zfs.csi.openebs.io
parameters:
  poolpattern: "zfspv-pool*"
```

- user will deploy a PVC using above storage class

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: demo-zfspv-vol-claim
spec:
  storageClassName: openebs-zfspv
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5G
```
## Workflow

### 1. CSI create volume
At CSI, when we get a Create Volume request, it will first try to find a node where it can create the PV object. The driver will trigger the scheduler which will return a node where the PV should be created.
In CreateVolume call, we will have the list of nodes where the ZFS pools are present and the volume should be created in any one of the node present in the list.

### 2. Agent Volume Controller
LocalPV-ZFS driver will create the PV object on scheduled node so that the applcation using that PV always comes to the same node and also it creates the ZFSVolume object for that volume in order to manage the creation of the ZFS dataset. There will be a watcher at each node which will be watching for the ZFSVolume resource which is aimed for them. The watcher is inbuilt into ZFS node-agent. As soon as ZFSVolume object is created for a node, the corresponding watcher will get the add event and it will check for the zpools present on that node whcih satisfies the poolpattern parameter. It will create a list of all the zfs pools matching the poolpattern regx and will pick the pool which has highest free space available which can accomodate the volume creation request. Once volume is created successfully, the node agent will update the ZFSVolume CR with poolname it has selected to create the volume


## Implementation Plan

### Phase 1
1. add support for multiple zpools in storageclass
2. add unit test cases

### Phase 2
1. BDD for multiple zpool support.
2. CI  pipelines setup to validate the software.
