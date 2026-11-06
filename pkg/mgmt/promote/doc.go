/*
Copyright 2026 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package promote is the node agent's ZFSPromote controller.

A ZFSPromote asks for `zfs promote` of one ZFSVolume. After it the volume
owns the snapshot it was cloned from and every earlier snapshot of its
origin, and the origin becomes a clone of the volume. The CI cache use case:
a candidate cloned from the trunk's snapshot passes and becomes the new
trunk, and the old trunk's lineage can then be deleted from under it.

The flow:

- the caller creates a ZFSPromote named after the volume, in the driver's
  namespace, with the volume's ownerNodeID and status Init;

- the agent on that node sets InProgress, runs `zfs promote` (skipped when
  the volume has no origin, as after an earlier promote), then relabels
  every ZFSSnapshot record whose snapshot now lives under the volume, and
  sets Done;

- on an error it sets Failed and records a Warning event with the reason
  (for example ENOSPC under a `quota` that cannot take the moved snapshots);

- the caller reads the result and deletes the ZFSPromote. There is no
  finalizer.

Every step is safe to rerun, so an agent restart re-handles Init and
InProgress resources from the start. Relabelling moves the snapshot record
off the old origin; when the old origin is marked for deletion and that was
its last snapshot, the controller's marked-volume reaper finishes deleting
it.
*/

package promote
