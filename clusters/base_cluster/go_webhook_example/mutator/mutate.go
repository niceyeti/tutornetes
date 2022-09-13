package mutator

import (
	k8s "k8s.io/api/core/v1"
)

type PodMutator struct {}


// Mutate returns a new mutated pod according to set env rules
func (pm *PodMutator) Mutate(inpod *corev1.Pod) (*corev1.Pod, error) {
	pod := inpod.DeepCopy()

	if pod. 

	return mpod, nil
}