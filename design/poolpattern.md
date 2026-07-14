---
title: Pool Pattern based Volume Provisioning for LocalPV-ZFS
authors:
  - "@krishna-gajabi"
creation-date: 2026-07-13
last-updated: 2026-07-13
status: proposed
---

# Pool Pattern (`poolpattern`) based Volume Provisioning

## Table of Contents

* [Summary](#summary)
* [Motivation](#motivation)
* [Current Behaviour](#current-behaviour)
* [Proposal](#proposal)
  * [StorageClass API](#storageclass-api)
  * [Pool Resolution and Scheduling](#pool-resolution-and-scheduling)
  * [Clone / Snapshot Paths](#clone--snapshot-paths)
  * [Validation](#validation)
* [Implementation Plan](#implementation-plan)
* [Test Plan](#test-plan)
* [Docs](#docs)
* [Resolved Decisions](#resolved-decisions)

## Summary

This proposal adds a `poolpattern` StorageClass parameter to LocalPV-ZFS. It is a
regular expression that selects the ZFS pool to provision into at CreateVolume
time, as an alternative to the existing fixed `poolname` parameter. This mirrors
the `vgpattern` feature in [lvm-localpv](https://github.com/openebs/lvm-localpv),
where `poolname`/`poolpattern` are the ZFS analogs of `volgroup`/`vgpattern`
(a ZFS **pool** is the analog of an LVM **volume group**).

## Motivation

Today an administrator must name one exact pool per StorageClass. When pools are
named inconsistently across nodes (e.g. `zfspv-pool-a`, `zfspv-pool-b`), or when
new pools are added, the StorageClass cannot target them without editing. A regex
lets one StorageClass span a family of pools and lets the driver choose among the
matching pools per node.

## Current Behaviour

`CreateZFSVolume` ([pkg/driver/controller.go](../pkg/driver/controller.go)):

1. Reads `poolname` from the (case-insensitive) SC parameters.
2. `getNodeMap(scheduler, pool)` ([pkg/driver/schd_helper.go](../pkg/driver/schd_helper.go))
   builds a `node -> weight` map for that single pool by scanning all `ZFSVolume`
   CRs — `VolumeWeighted` counts volumes, `CapacityWeighted` sums their capacity.
3. `schd.Scheduler(req, nmap)` (`github.com/openebs/lib-csi/pkg/scheduler`) filters
   nodes by topology and returns them ordered least-loaded-first. It only orders
   **nodes**; the pool is uniform across the returned list.
4. A single `ZFSVolume` is built with the fixed `poolname` and provisioning is tried
   on each node in order. The resolved pool is stored in `ZFSVolume.Spec.PoolName`
   and returned to CSI in the volume context (`zfs.PoolNameKey`).

Key facts the design relies on:

* `ZFSNode` CRs ([pkg/apis/openebs.io/zfs/v1/zfsnode.go](../pkg/apis/openebs.io/zfs/v1/zfsnode.go))
  already advertise every pool on each node with `Free`/`Used` capacity — this is
  the authoritative list of pools to match a pattern against (a matching pool may
  have zero volumes yet, so `ZFSVolume` list alone is insufficient).
* No CRD changes are needed: StorageClass parameters are free-form, and the resolved
  concrete pool is already persisted per-volume in `ZFSVolume.Spec.PoolName`.
* The node plugin reads the pool from the `ZFSVolume` CR, so no node-side changes.

## Proposal

### StorageClass API

Add one parameter, `poolpattern`, a Go regular expression (RE2, unanchored — users
anchor with `^...$` if desired). Precedence mirrors lvm-localpv:

* Exactly one of `poolname` / `poolpattern` should be set.
* If both are set, `poolname` wins and `poolpattern` is ignored.
* If neither is set, CreateVolume fails with `InvalidArgument`.

```yaml
parameters:
  poolpattern: "zfspv-pool.*"   # regex; pool chosen by the scheduler among matches
```

### Pool Resolution and Scheduling

Selection policy (per the decision on this feature): **reuse the existing node
scheduler's weight metric to also pick among matching pools**, rather than a simple
"largest free capacity" pick. Free capacity is used only as a *fit* filter (a pool
whose `Free` cannot hold the requested size is excluded), not as the selection key.

`poolname` mode is unchanged. For `poolpattern` mode, extend `schd_helper.go`:

1. Compile the regex once (fail fast on error).
2. List `ZFSNode` CRs → `node -> []Pool`. List `ZFSVolume` CRs → per-`(node,pool)`
   weight using the selected algorithm (volume count or summed capacity), exactly
   as the existing per-pool weight is computed today.
3. For each node, take the pools whose name matches the regex **and** whose `Free`
   can fit the requested size. Among those, pick the pool with the **minimum weight**
   (fewest volumes / least provisioned capacity) — this is "reuse the node scheduler"
   applied one level down, to pools. Skip nodes with no matching+fitting pool.
4. Produce two maps: `nmap[node] = weight-of-chosen-pool` (fed to
   `schd.Scheduler`, unchanged, for node ordering) and `poolMap[node] = chosenPool`.
5. In the provisioning loop, build the `ZFSVolume` per node with
   `WithPoolName(poolMap[node])`; skip any returned node absent from `poolMap`.

`lib-csi` is not modified — it continues to order nodes; pool selection stays in
this repo.

Proposed helper shape:

```go
// returns node->weight for the scheduler and node->concretePool for provisioning.
func getNodeMapForPattern(schd, pattern string, size uint64) (map[string]int64, map[string]string, error)
```

`CreateZFSVolume` returns the resolved pool (new return value) so `CreateVolume`
can set the response context and log line correctly in pattern mode.

### Clone / Snapshot Paths

`CreateVolClone` / `CreateSnapClone` ([controller.go](../pkg/driver/controller.go))
currently require `source.Spec.PoolName == poolname` (a ZFS clone must live in the
source's pool). Change: if `poolname` is set keep the exact check; if `poolpattern`
is set instead, require the **source pool to match the regex**. The resolved pool for
a clone is always the source's pool, returned up to `CreateVolume` unchanged.

### Validation

In `validateVolumeCreateReq` ([controller.go:1189](../pkg/driver/controller.go#L1189)):
reject when both `poolname` and `poolpattern` are empty, and reject a `poolpattern`
that fails to compile — both as `codes.InvalidArgument`.

## Implementation Plan

1. `schd_helper.go` — add `getNodeMapForPattern` (pool matching + per-pool weighting
   + fit filter) and a small pool-weight helper; reuse `VolumeWeighted`/
   `CapacityWeighted` constants.
2. `controller.go` — read `poolpattern`; branch in `CreateZFSVolume`; thread the
   resolved pool through `CreateZFSVolume` / `CreateVolClone` / `CreateSnapClone`
   return values into `CreateVolume`'s response context; update clone/snap pool checks.
3. `validateVolumeCreateReq` — presence + regex-compile validation.
4. Docs (see below) and samples.

## Test Plan

* Unit: pattern matching, precedence (`poolname` wins), no-match → empty node list,
  fit filtering, min-weight pool selection under both algorithms, invalid regex.
* BDD (`tests/`, `ci/ci-test.sh`): provision with `poolpattern` across nodes with
  differently-named pools; clone/snapshot of a pattern-provisioned volume.

## Docs

Update [docs/storageclasses.md](../docs/storageclasses.md) and
[docs/scheduler.md](../docs/scheduler.md); add a sample StorageClass.

## Resolved Decisions

* **Fit filter**: **in scope, but only for thick volumes.** For volumes that are
  not explicitly thin (`thinprovision != "yes"` — i.e. `"no"` or unset, both of
  which get a ZFS reservation), matching pools are filtered by `Free >= size` using
  the `ZFSNode`-advertised free capacity; pools (and nodes) that cannot hold the
  reservation are excluded before scheduling. For thin volumes (`thinprovision:
  "yes"`, created with `zfs create -s`) only the `Free`-based exclusion is
  **skipped** — pattern matching, weighted pool selection, and concrete-pool
  resolution all still apply, so thin volumes are fully supported in pattern mode.
  The exclusion is dropped because overcommit is the whole point: the create
  succeeds regardless of free space, and gating on `Free` would wrongly leave thin
  PVCs Pending. This makes pattern mode
  capacity-aware for thick volumes (the fixed-`poolname` path is not) while staying
  consistent with the fixed path for thin volumes, which defer the capacity
  decision to `zfs create` at the agent.
* **Anchoring**: **unanchored RE2**, matching lvm-localpv's `vgpattern`
  (`regexp.MatchString` semantics). Users anchor with `^...$` when they need a full
  match. Chosen so migration between lvm-localpv and zfs-localpv behaves identically.
* **CapacityWeighted metric for pools**: **summed capacity of driver-provisioned
  `ZFSVolume`s**, identical to the existing `getCapacityWeightedMap`. Keeps
  `poolname` and `poolpattern` modes consistent; non-driver data in the pool does
  not affect the weight (but does affect the fit filter above, via `Free`).
