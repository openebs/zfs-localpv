/*
Copyright 2024 The OpenEBS Authors

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

package zfs

import (
	"context"
	"fmt"
	"sync"
	"time"

	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

var (
	eventClientOnce sync.Once
	eventClient     kubernetes.Interface
	eventClientErr  error
)

// getEventClient lazily constructs (and caches) a kubernetes clientset
// used by the CSI controller to read events that the node agent emits
// on the ZFS CRs. We only need CoreV1 Events here, but we keep a full
// kubernetes.Interface for simplicity — it matches the existing
// initialization pattern used elsewhere in the package.
func getEventClient() (kubernetes.Interface, error) {
	eventClientOnce.Do(func() {
		cfg, err := k8sapi.Config().Get()
		if err != nil {
			eventClientErr = fmt.Errorf("build kubeconfig: %w", err)
			return
		}
		c, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			eventClientErr = fmt.Errorf("build k8s clientset: %w", err)
			return
		}
		eventClient = c
	})
	return eventClient, eventClientErr
}

// LatestWarningEvent returns the reason and message of the most recent
// Warning event recorded on the resource identified by (kind, name,
// namespace, uid). It returns empty strings and a nil error if no
// matching event exists — callers are expected to fall back to a
// generic message in that case.
//
// The lookup uses a field selector against core/v1.Event rather than an
// informer: CSI controller waits are short-lived per RPC, the event
// namespace is the openebs namespace (small), and avoiding an informer
// keeps the controller's startup path and memory footprint unchanged.
func LatestWarningEvent(ctx context.Context, kind, name, namespace string, uid types.UID) (reason, message string, err error) {
	cli, err := getEventClient()
	if err != nil {
		return "", "", err
	}

	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.kind", kind),
		fields.OneTermEqualSelector("involvedObject.name", name),
		fields.OneTermEqualSelector("involvedObject.uid", string(uid)),
		fields.OneTermEqualSelector("type", corev1.EventTypeWarning),
	).String()

	list, err := cli.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: selector,
		Limit:         50,
	})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("list events for %s/%s: %w", namespace, name, err)
	}

	latest := latestEvent(list.Items)
	if latest == nil {
		return "", "", nil
	}
	return latest.Reason, latest.Message, nil
}

// latestEvent returns the event with the most recent observation time
// from the given slice, preferring LastTimestamp, then EventTime, then
// FirstTimestamp. Returns nil if the slice is empty.
func latestEvent(items []corev1.Event) *corev1.Event {
	var best *corev1.Event
	for i := range items {
		e := &items[i]
		if best == nil || eventTime(e).After(eventTime(best)) {
			best = e
		}
	}
	return best
}

func eventTime(e *corev1.Event) time.Time {
	if !e.LastTimestamp.IsZero() {
		return e.LastTimestamp.Time
	}
	if !e.EventTime.IsZero() {
		return e.EventTime.Time
	}
	return e.FirstTimestamp.Time
}
