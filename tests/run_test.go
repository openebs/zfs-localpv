/*
Copyright 2020 The OpenEBS Authors

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

package tests

import (
	"bytes"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	snapYAML = `apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: #snapname
spec:
  volumeSnapshotClassName: zfspv-snapclass
  source:
    persistentVolumeClaimName: #pvcname
`

	cloneYAML = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: #pvcname
spec:
  dataSource:
    name: #snapname
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: #storageclass
`
)

func execAtLocal(cmd string, input []byte, args ...string) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	command := exec.Command(cmd, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	if len(input) != 0 {
		command.Stdin = bytes.NewReader(input)
	}

	err := command.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func kubectl(args ...string) ([]byte, []byte, error) {
	return execAtLocal("kubectl", nil, args...)
}

func kubectlWithInput(input []byte, args ...string) ([]byte, []byte, error) {
	return execAtLocal("kubectl", input, args...)
}

func verifySnapshotCreated(snapName string) bool {
	Eventually(func() bool {
		stdout, stderr, err := kubectl("get", "volumesnapshots.snapshot", snapName, "-n", OpenEBSNamespace, "-o=template", "--template={{.status.readyToUse}}")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		return strings.TrimSpace(string(stdout)) == "true"
	}, 240, 5).Should(BeTrue())
	return true
}

func createSnapshot(pvcName, snapName string) {
	By("creating snapshot for a pvc " + pvcName)

	tyaml := strings.Replace(snapYAML, "#pvcname", pvcName, -1)
	yaml := strings.Replace(tyaml, "#snapname", snapName, -1)

	stdout, stderr, err := kubectlWithInput([]byte(yaml), "apply", "-n", OpenEBSNamespace, "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}

func createClone(clonepvc, snapname, storageclass string) {
	By("creating clone volume from snapshot " + snapname)

	syaml := strings.Replace(cloneYAML, "#snapname", snapname, -1)
	cyaml := strings.Replace(syaml, "#pvcname", clonepvc, -1)
	yaml := strings.Replace(cyaml, "#storageclass", storageclass, -1)

	stdout, stderr, err := kubectlWithInput([]byte(yaml), "apply", "-n", OpenEBSNamespace, "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}

func deleteSnapshot(pvcName, snapName string) {
	By("deleting the snapshot " + snapName)

	tyaml := strings.Replace(snapYAML, "#pvcname", pvcName, -1)
	yaml := strings.Replace(tyaml, "#snapname", snapName, -1)

	stdout, stderr, err := kubectlWithInput([]byte(yaml), "delete", "-n", OpenEBSNamespace, "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}
