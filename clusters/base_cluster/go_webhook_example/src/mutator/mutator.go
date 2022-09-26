package mutator

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/wI2L/jsondiff"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

// PodMutator takes an admission request and mutates the pods within it.
// This is a just a simplified version of the more complete webhook example from Slack:
// https://github.com/slackhq/simple-kubernetes-webhook/blob/main/pkg/mutation/inject_env.go
// This version is a cattywampus minimal implementation, no thought to encapsulation or responsibilities.
type PodMutator struct {
	Request *admissionv1.AdmissionRequest
}

// Mutate
func (pm *PodMutator) Mutate() (*admissionv1.AdmissionReview, error) {
	pod, err := pm.pod()
	if err != nil {
		return nil, fmt.Errorf("could not parse pod in admission review request: %v", err)
	}

	mutatedPod, err := pm.mutatePod(pod)
	if err != nil {
		return nil, fmt.Errorf("could not mutate pod: %v", err)
	}

	patch, err := jsondiff.Compare(pod, mutatedPod)
	if err != nil {
		return nil, err
	}

	patchb, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	prr, err := patchReviewResponse(pm.Request.UID, patchb)
	if err != nil {
		return nil, err
	}

	return prr, nil
}

// Pod extracts a pod from an admission request
func (pm *PodMutator) pod() (*corev1.Pod, error) {
	if pm.Request.Kind.Kind != "Pod" {
		return nil, fmt.Errorf("kind must be pod but received " + pm.Request.Kind.Kind)
	}

	p := corev1.Pod{}
	if err := json.Unmarshal(pm.Request.Object.Raw, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

// mutatePod is where you would update the pod: add sidecars, security policies,
// resource limits, annotations, etc.
func (pm *PodMutator) mutatePod(inpod *corev1.Pod) (*corev1.Pod, error) {
	pod := inpod.DeepCopy()

	// Your code here: modify the pod (add a sidecar container, resource limits, security stuff, etc)
	// Currently this just logs a message and returns the pod without modifications.
	fmt.Println("Hit webhook! " + time.Now().String())

	return pod, nil
}

// patchReviewResponse builds an admission review with given json patch
func patchReviewResponse(uid types.UID, patch []byte) (*admissionv1.AdmissionReview, error) {
	patchType := admissionv1.PatchTypeJSONPatch

	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:       uid,
			Allowed:   true,
			PatchType: &patchType,
			Patch:     patch,
		},
	}, nil
}
