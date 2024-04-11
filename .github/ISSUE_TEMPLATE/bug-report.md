---
name: Bug report
about: Tell us about a problem you are experiencing
labels: Bug

---

**What steps did you take and what happened:**
[A clear and concise description of what the bug is, and what commands you ran.]


**What did you expect to happen:**


**The output of the following commands will help us better understand what's going on**:
(Pasting long output into a [GitHub gist](https://gist.github.com) or other [Pastebin](https://pastebin.com/) is fine.)

* `kubectl logs -f openebs-zfs-controller-f78f7467c-blr7q -n openebs -c openebs-zfs-plugin`
* `kubectl logs -f openebs-zfs-node-[xxxx] -n openebs -c openebs-zfs-plugin`
* `kubectl get pods -n openebs`
* `kubectl get zv -A -o yaml`

**Anything else you would like to add:**
[Miscellaneous information that will assist in solving the issue.]


**Environment:**
- ZFS-LocalPV version
- Kubernetes version (use `kubectl version`):
- Kubernetes installer & version:
- Cloud provider or hardware configuration:
- OS (e.g. from `/etc/os-release`):
