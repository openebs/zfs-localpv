package zfs

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

// maxEventMessageLen is the Kubernetes API server hard limit for Event.Message.
// Messages longer than this are rejected with a validation error.
const maxEventMessageLen = 1024

func truncateEventMessage(msg string) string {
	if len(msg) <= maxEventMessageLen {
		return msg
	}
	return msg[:maxEventMessageLen-3] + "..."
}

// EmitFailureEvent records a Warning event on obj. The event message is
// err.Error(), which for *ZFSError includes the verbatim ZFS stderr so
// the real failure cause is visible in kubectl describe output.
func EmitFailureEvent(recorder record.EventRecorder, obj runtime.Object, reason string, err error) {
	if recorder == nil || err == nil || obj == nil {
		return
	}
	recorder.Event(obj, corev1.EventTypeWarning, reason, truncateEventMessage(err.Error()))
}

// EmitSuccessEvent records a Normal event on obj for terminal
// user-visible transitions (provisioned, destroyed, etc.).
func EmitSuccessEvent(recorder record.EventRecorder, obj runtime.Object, reason, message string) {
	if recorder == nil || obj == nil {
		return
	}
	recorder.Event(obj, corev1.EventTypeNormal, reason, truncateEventMessage(message))
}
