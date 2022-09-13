package mutator

import (
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	k8s "k8s.io/api/core/v1"
)

// PodMutator takes an admission request and mutates the pods within it.
// This is a just a simplified version of the more complete webhook example from Slack:
// https://github.com/slackhq/simple-kubernetes-webhook/blob/main/pkg/mutation/inject_env.go
type PodMutator struct {
	Request *admissionv1.AdmissionRequest
}

// Mutate is where you would update the pod: add sidecars, security policies,
// resource limits, annotations, etc.
func (pm *PodMutator) Mutate(inpod *k8s.Pod) (*k8s.Pod, error) {
	pod := inpod.DeepCopy()

	// Your code here: modify the pod (add a sidecar container, resource limits, security stuff, etc)
	// Currently this just echoes the pod without modifications.

	// Confirm the hook has been called by printing the logs for the hook container
	fmt.Println("Hit ")

	return pod, nil
}
