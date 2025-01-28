## OpenEBS - LocalPV-ZFS CSI Driver
[![Build Status](https://github.com/openebs/zfs-localpv/actions/workflows/build.yml/badge.svg)](https://github.com/openebs/zfs-localpv/actions/workflows/build.yml)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fzfs-localpv.svg?type=shield&issueType=license)](https://app.fossa.com/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fzfs-localpv?ref=badge_shield&issueType=license)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/3523/badge)](https://www.bestpractices.dev/projects/3523)
[![Slack](https://img.shields.io/badge/chat-slack-ff1493.svg?style=flat-square)](https://kubernetes.slack.com/messages/openebs/)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://us05web.zoom.us/j/87535654586?pwd=CigbXigJPn38USc6Vuzt7qSVFoO79X.1)
[![Go Report](https://goreportcard.com/badge/github.com/openebs/zfs-localpv)](https://goreportcard.com/report/github.com/openebs/zfs-localpv)

## Overview

### What is OpenEBS ZFS LocalPV?
OpenEBS ZFS LocalPV is a [CSI](https://github.com/container-storage-interface/spec) plugin for implementation of [ZFS](https://en.wikipedia.org/wiki/ZFS) backed persistent volumes for Kubernetes. It is a local storage solution, which means the device, volume and the application are on the same host. It doesn't contain any dataplane, i.e only its simply a control-plane for the kernel zfs volumes. It mainly comprises of two components which are implemented in accordance to the CSI Specs:

1. CSI Controller - Frontends the incoming requests and initiates the operation.
2. CSI Node Plugin - Serves the requests by performing the operations and making the volume available for the initiator.

### Why OpenEBS ZFS LocalPV?
1. Light weight, easy to setup storage provisoner for provisioning node local volumes in k8s ecosystem.
2. Makes ZFS stack available to k8s, allowing end users to use the ZFS functionalites like snapshot, restore, clone, thin provisioning, resize, encryption, compression, dedup, etc for their Persistent Volumes.
3. Cloud native, i.e based on CSI spec, so certified to run on K8s.

### Architecture

LocalPV refers to storage that is directly attached to a specific node in the Kubernetes cluster. It uses locally available disks (e.g., SSDs, HDDs) on the node.

<b>Use Case</b>: Ideal for workloads that require low-latency access to storage or when data locality is critical (e.g., databases, caching systems).

#### Characteristics:
- <b>Node-bound</b>: The volume is tied to the node where the disk is physically located.
- <b>No replication</b>: Data is not replicated across nodes, so if the node fails, the data may become inaccessible.
- <b>High performance</b>: Since the storage is local, it typically offers lower latency compared to network-attached storage.

```mermaid
graph TD;
  subgraph Node1["Node 1"]
    subgraph K8S NODE 1[" "]
      NODE_1_PV1["PV 1"] --> NODE_1_APP1["APP 1"]
      NODE_1_PV2["PV 2"] --> NODE_1_APP2["APP 2"]
    end
    subgraph ZFS Stack 2["ZFS Stack"]
      ZPOOL_1_1 --> ZVOL1_1["ZVOL 1"]
      ZPOOL_1_1 --> ZVOL3_1["ZVOL 2"]
      ZVOL1_1["ZVOL 1"] --> NODE_1_PV1 
      ZVOL3_1["ZVOL 2"] --> NODE_1_PV2
    end
    subgraph Blockdevices1[" "]
      NODE_1_DISK_1["/dev/sdc"] --> ZPOOL_1_1["ZPOOL"]
      NODE_1_DISK_2["/dev/sdb"] --> ZPOOL_1_1["ZPOOL"]
    end
  end

  subgraph Node2["Node 2"]
    subgraph K8S NODE 2[" "]
      NODE_2_PV1["PV 1"] --> NODE_2_APP1["APP 1"]
    end
    subgraph ZFS Stack 1["ZFS Stack"]
      ZPOOL_2_2 --> ZVOL_2_2["ZVOL 1"]
      ZVOL_2_2["ZVOL 1"] --> NODE_2_PV1 
    end
    subgraph Blockdevices2[" "]
      NODE_2_DISK["/dev/sdb"] --> ZPOOL_2_2["ZPOOL"]
    end
  end
```

### Supported System

> | Name | Version |
> | :--- | :--- |
> | K8S | 1.23+ |
> | Distro | Alpine, Arch, CentOS, Debian, Fedora, NixOS, SUSE, RHEL, Ubuntu |
> | Kenel | oldest supported kernel is 2.6.32 |
> | ZFS | 0.7, 0.8, 2.2.3 |
> | Memory | ECC Memory is highly recommended |
> | RAM | 8GiB for best perf with Dedup enabled. (Will work with 2GiB or less without Dedup) |

Check the [features](./docs/features.md) supported for each k8s version.

### Documents

- [Prerequisites](./docs/quickstart.md#prerequisites)
- [Quickstart](./docs/quickstart.md#setup)
- [Developer Setup](./docs/developer-setup.md#development-workflow)
- [Testing](./docs/developer-setup.md#testing)
- [Contibuting Guidelines](./CONTRIBUTING.md)
- [Governance](./GOVERNANCE.md)
- [Changelog](./CHANGELOG.md)
- [Release Process](./RELEASE.md)

## Features

- [x] Access Modes
    - [x] ReadWriteOnce
    - ~~ReadOnlyMany~~
    - ~~ReadWriteMany~~
- [x] Volume modes
    - [x] `Filesystem` mode
    - [x] `Block` mode
- [x] Supports fsTypes: `ext4`, `btrfs`, `xfs`, `zfs`
- [x] Volume metrics
- [x] [Snapshot]
    - [x] [Create](docs/snapshot.md)
    - [x] [Restore](docs/clone.md#create-clone-from-snapshot)
- [x] [Clone](docs/clone.md#create-clone-from-volume)
- [x] [Volume Resize](docs/resize.md)
- [x] [Raw Block Volume](docs/raw-block-volume.md)
- [x] [Backup/Restore](docs/backup-restore.md)
- [ ] Ephemeral inline volume

## Dev Activity dashboard
![Alt](https://repobeats.axiom.co/api/embed/d990adda232a580d4c0fd9b98d6557079bb3bf4a.svg "Repobeats analytics image")

## License compliance
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fzfs-localpv.svg?type=large&issueType=license)](https://app.fossa.com/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fzfs-localpv?ref=badge_large&issueType=license)
