---
title: SELinux Mount Support for LocalPV-ZFS
authors:
  - "@cjorge-graphops"
creation-date: 2024-11-01
last-updated: 2024-11-01
---

# SELinux Mount Support for LocalPV-ZFS

## Table of Contents

* [Summary](#summary)
* [Problem](#problem)
* [Proposal](#proposal)
* [Implementation Plan](#implementation-plan)
* [Test Plan](#test-plan)
* [GA Criteria](#ga-criteria)

## Summary

This design proposes adding SELinux mount support to the ZFS LocalPV CSI driver by enabling the `seLinuxMount` feature in the CSIDriver specification. This will allow Kubernetes to perform efficient SELinux relabeling for volumes containing many small files. Technical details on this and motivation are available in the [KEP](https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1710-selinux-relabeling).

## Problem

When PVCs contain millions of small files, SELinux relabeling takes an enormous amount of time. This is because the CRI (i.e CRI-O) currently has to recursively relabel all files individually. With the SELinux mount support feature introduced in Kubernetes 1.27, this process can be optimized by passing SELinux context through mount options, but requires CSI driver support.

## Proposal

Enable SELinux mount support by:

1. Setting `seLinuxMount: true` in the CSIDriver specification
2. Location: `deploy/helm/charts/templates/csidriver.yaml`

This is valid because:
- ZFS volumes are independently mounted from the Linux kernel's perspective
- ZFS supports SELinux context mount options

No additional changes are required in the CSI driver implementation as Kubernetes will:
- Only pass `-o context=<SELinux label>` when all conditions are met
- Verify SELinux OS support before attempting to use the feature
- Handle the feature gates and pod requirements

No need for further upgrade-path considerations as the CSIDriver can be updated in place

## Implementation Plan

1. Add `seLinuxMount: true` to the CSIDriver spec:
```yaml
apiVersion: storage.k8s.io/v1
kind: CSIDriver
metadata:
  name: zfs.csi.openebs.io
spec:
  seLinuxMount: true
  # ... existing fields remain unchanged
```

2. Document requirements in README:
   - Feature gates needed: `ReadWriteOncePod` and `SELinuxMountReadWriteOncePod`
   - Kubernetes version ≥ 1.27 for beta support
   - Pod must specify SELinux level
   - PV must be RWOP.

Note: I'm unsure this is necessary or makes sense to document, thoughts?

## Test Plan

1. Verify basic functionality:
   - Create PVC with millions of small files
   - Verify mount succeeds with SELinux context
   - Confirm relabeling performance improvement

Note: I explored validating that indeed Kubernetes acts as the KEP design describes by:
   - Test with feature gates disabled
   - Test with PODs that don't set SELinux level or PVCs that aren't `ReadWriteOncePod`

had no easily available non-SELinux system to test with. But these aren't really the concern of zfs-localpv as the responsibility lies elsewhere.

## GA Criteria

1. All test cases passing consistently
2. No reported issues with SELinux mount support
3. Documentation updated and complete
