/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	goopv1alpha1 "github.com/example/goop/api/v1alpha1"
)

// TODO (Jesse): revisit finalizer. See other TODO for link to docs.
const goopFinalizer = "goop.example.com/finalizer"

// Definitions to manage status conditions
const (
	// typeAvailableGoop represents the status of the Deployment reconciliation
	typeAvailableGoop = "Available"
	// typeDegradedGoop represents the status used when the custom resource is deleted and the finalizer operations are must to occur.
	typeDegradedGoop = "Degraded"
)

// GoopReconciler reconciles a Goop object
type GoopReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=goop.example.com,resources=goops,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=goop.example.com,resources=goops/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=goop.example.com,resources=goops/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Goop object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
// Returns: see the Result{} definition.
func (r *GoopReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Goop instance
	// The purpose is to check if the Custom Resource for the Kind Goop
	// is applied on the cluster, and if not we return nil to stop the reconciliation
	goop := &goopv1alpha1.Goop{}
	err := r.Get(ctx, req.NamespacedName, goop)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation.
			log.Info("goop resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get goop")
		return ctrl.Result{}, err
	}

	// Let's just set the status as Unknown when no status are available
	if goop.Status.Conditions == nil || len(goop.Status.Conditions) == 0 {
		meta.SetStatusCondition(
			&goop.Status.Conditions,
			metav1.Condition{
				Type:    typeAvailableGoop,
				Status:  metav1.ConditionUnknown,
				Reason:  "Reconciling",
				Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, goop); err != nil {
			log.Error(err, "Failed to update Memcached status")
			return ctrl.Result{}, err
		}

		// Re-fetch the goop Custom Resource after update the status so that
		// we have the latest state of the resource on the cluster and we will avoid
		// raising "the object has been modified, please apply your changes to the
		// latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations
		if err := r.Get(ctx, req.NamespacedName, goop); err != nil {
			log.Error(err, "Failed to re-fetch goop")
			return ctrl.Result{}, err
		}
	}

	// Adds a finalizer, then we can define some operations that should occur
	// before the custom resource deletetion.
	// TODO (Jesse): figure out finalizer reqs
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	if !controllerutil.ContainsFinalizer(goop, goopFinalizer) {
		log.Info("Adding Finalizer for Memcached")
		if ok := controllerutil.AddFinalizer(goop, goopFinalizer); !ok {
			log.Error(err, "Failed to add finalizer into the custom resource")
			return ctrl.Result{Requeue: true}, nil
		}

		if err = r.Update(ctx, goop); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Check if the Goop instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isGoopMarkedToBeDeleted := goop.GetDeletionTimestamp() != nil
	if isGoopMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(goop, goopFinalizer) {
			log.Info("Performing Finalizer Operations for Goop before deleting CR")

			// Add a status "Downgrade" to define that this resource begins its process to be terminated.
			meta.SetStatusCondition(
				&goop.Status.Conditions,
				metav1.Condition{
					Type:    typeDegradedGoop,
					Status:  metav1.ConditionUnknown,
					Reason:  "Finalizing",
					Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", goop.Name)})

			if err := r.Status().Update(ctx, goop); err != nil {
				log.Error(err, "Failed to update Memcached status")
				return ctrl.Result{}, err
			}

			// Perform all operations required before remove the finalizer and allow
			// the Kubernetes API to remove the custom resource.
			// TODO: I need to implement this
			r.doFinalizerOperationsForGoop(goop)

			// TODO(user): If you add operations to the doFinalizerOperationsForGoop method
			// then you need to ensure that all worked fine before deleting and updating the Downgrade status
			// otherwise, you should requeue here.

			// Re-fetch the goop Custom Resource before updating its status,
			// such that we have the latest state of the resource on the cluster and avoid
			// raising "the object has been modified, please apply your changes to the
			// latest version and try again" which would re-trigger the reconciliation
			if err := r.Get(ctx, req.NamespacedName, goop); err != nil {
				log.Error(err, "Failed to re-fetch goop")
				return ctrl.Result{}, err
			}

			meta.SetStatusCondition(
				&goop.Status.Conditions,
				metav1.Condition{
					Type:    typeDegradedGoop,
					Status:  metav1.ConditionTrue,
					Reason:  "Finalizing",
					Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", goop.Name)})

			if err := r.Status().Update(ctx, goop); err != nil {
				log.Error(err, "Failed to update Memcached status")
				return ctrl.Result{}, err
			}

			log.Info("Removing Finalizer for Goop after successfully perform the operations")
			if ok := controllerutil.RemoveFinalizer(goop, goopFinalizer); !ok {
				log.Error(err, "Failed to remove finalizer for Goop")
				return ctrl.Result{Requeue: true}, nil
			}

			if err := r.Update(ctx, goop); err != nil {
				log.Error(err, "Failed to remove finalizer for Goop")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Check if the daemonset already exists, if not create a new one.
	// NOTE: querying things like daemonsets requires RBAC permission by the
	// service-account to do so. Failing to do so gives errors such as:
	/// 'system:serviceaccount:goop-system:goop-controller-manager" cannot list resource "daemonsets" in API group "apps".
	// The solution is simply to add the appropriate permissions via roles and rolebindings.
	found := &appsv1.DaemonSet{}
	err = r.Get(
		ctx,
		types.NamespacedName{
			Name:      goop.Name,
			Namespace: goop.Namespace,
		},
		found)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new deployment
		ds, err := r.daemonsetForGoop(goop)
		if err != nil {
			log.Error(err, "Failed to define new Daemonset resource for Goop")

			// The following implementation will update the status
			meta.SetStatusCondition(
				&goop.Status.Conditions,
				metav1.Condition{
					Type:    typeAvailableGoop,
					Status:  metav1.ConditionFalse,
					Reason:  "Reconciling",
					Message: fmt.Sprintf("Failed to create Deployment for the custom resource (%s): (%s)", goop.Name, err)})

			if err := r.Status().Update(ctx, goop); err != nil {
				log.Error(err, "Failed to update Memcached status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		log.Info("Creating a new job Daemonset",
			"Daemonset.Namespace", ds.Namespace, "Daemonset.Name", ds.Name)
		if err = r.Create(ctx, ds); err != nil {
			log.Error(err, "Failed to create new Daemonset",
				"Daemonset.Namespace", ds.Namespace, "Daemonset.Name", ds.Name)
			return ctrl.Result{}, err
		}

		// DS created successfully
		// We will requeue the reconciliation so that we can ensure the state
		// and move forward for the next operations
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Daemonset")
		// Let's return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// The CRD API is defining that the Memcached type, have a MemcachedSpec.Size field
	// to set the quantity of Deployment instances is the desired state on the cluster.
	// Therefore, the following code will ensure the Deployment size is the same as defined
	// via the Size spec of the Custom Resource which we are reconciling.
	/*  // TODO: this handled replica size change in the memcache example. I should likewise
	    // implement Update logic, but will do so after getting the controller to partially work.
	size := goop.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		if err = r.Update(ctx, found); err != nil {
			log.Error(err, "Failed to update Deployment",
				"Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)

			// Re-fetch the memcached Custom Resource before updating the status
			// so that we have the latest state of the resource on the cluster and we will avoid
			// raise the issue "the object has been modified, please apply
			// your changes to the latest version and try again" which would re-trigger the reconciliation
			if err := r.Get(ctx, req.NamespacedName, goop); err != nil {
				log.Error(err, "Failed to re-fetch goop")
				return ctrl.Result{}, err
			}

			// The following implementation will update the status
			meta.SetStatusCondition(&goop.Status.Conditions, metav1.Condition{
				Type:    typeAvailableGoop,
				Status:  metav1.ConditionFalse,
				Reason:  "Resizing",
				Message: fmt.Sprintf("Failed to update the size for the custom resource (%s): (%s)", goop.Name, err)})

			if err := r.Status().Update(ctx, goop); err != nil {
				log.Error(err, "Failed to update Memcached status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		// TODO: evaluate and translate comment. I need to understand the context for requeuing.
		// Now, that we update the size we want to requeue the reconciliation
		// so that we can ensure that we have the latest state of the resource before
		// update. Also, it will help ensure the desired state on the cluster
		return ctrl.Result{Requeue: true}, nil
	}
	*/

	// The following implementation will update the status
	meta.SetStatusCondition(
		&goop.Status.Conditions,
		metav1.Condition{
			Type:    typeAvailableGoop,
			Status:  metav1.ConditionTrue,
			Reason:  "Reconciling",
			Message: fmt.Sprintf("Daemonset for custom resource (%s) created successfully", goop.Name)})

	if err := r.Status().Update(ctx, goop); err != nil {
		log.Error(err, "Failed to update Goop status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// finalizeMemcached performs the required operations before deleting the CR.
func (r *GoopReconciler) doFinalizerOperationsForGoop(cr *goopv1alpha1.Goop) {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.

	// Note: It is not recommended to use finalizers with the purpose of delete resources which are
	// created and managed in the reconciliation. These ones, such as the Deployment created on this reconcile,
	// are defined as depended of the custom resource. See that we use the method ctrl.SetControllerReference.
	// to set the ownerRef which means that the Deployment will be deleted by the Kubernetes API.
	// More info: https://kubernetes.io/docs/tasks/administer-cluster/use-cascading-deletion/

	// The following implementation will raise an event
	// TODO: I need to implement this, but my generated code is missing a Reconciler (?)
	//r.Recorder.Event(
	//	cr,
	//	"Warning",
	//	"Deleting",
	//	fmt.Sprintf("Custom Resource %s is being deleted from the namespace %s",
	//		cr.Name,
	//		cr.Namespace))
}

/*
	Deploying Jobs to nodes such that a Job runs independently on every node
	can be done using a Daemonset that runs the job logic in init containers.
	Its very likely a more modern pattern exists, perhaps using Jobs with topology
	spread constraints. I have no idea if the daemonset pattern obtains all desired
	lifecycle requirements for job-like behavior, or if this is a hack.
	This yaml pattern comes from the github issue reply here: https://github.com/kubernetes/kubernetes/issues/36601

apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: auto-pull-images
  namespace: default
  labels:
    k8s-app: auto-pull-images
spec:
  selector:
    matchLabels:
      name: auto-pull-images
  template:
    metadata:
      labels:
        name: auto-pull-images
    spec:
      initContainers:
        - name: serverless-template-container
          image: unfor19/serverless-template
          resources:
            limits:
              cpu: 100m
              memory: 100Mi
            requests:
              cpu: 100m
              memory: 100Mi
      containers:
        - name: pause
          image: gcr.io/google_containers/pause
          resources:
            limits:
              cpu: 50m
              memory: 50Mi
            requests:
              cpu: 50m
              memory: 50Mi
*/

func (r *GoopReconciler) daemonsetForGoop(
	goop *goopv1alpha1.Goop) (*appsv1.DaemonSet, error) {
	ls := labelsForGoop(goop.Name)

	// Get the Operand image
	image, err := imageForGoop()
	if err != nil {
		return nil, err
	}

	dep := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      goop.Name,
			Namespace: goop.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.OnDeleteDaemonSetStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &[]bool{true}[0],
						// IMPORTANT: seccomProfile was introduced with Kubernetes 1.19
						// If you are looking for to produce solutions to be supported
						// on lower versions you must remove this option.
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					InitContainers: []corev1.Container{{
						Image:           image,
						Name:            "goop",
						ImagePullPolicy: corev1.PullIfNotPresent,
						// Ensure restrictive context for the container
						// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
						SecurityContext: &corev1.SecurityContext{
							// WARNING: Ensure that the image used defines an UserID in the Dockerfile
							// otherwise the Pod will not run and will fail with "container has runAsNonRoot and image has non-numeric user"".
							// If you want your workloads admitted in namespaces enforced with the restricted mode in OpenShift/OKD vendors
							// then, you MUST ensure that the Dockerfile defines a User ID OR you MUST leave the "RunAsNonRoot" and
							// "RunAsUser" fields empty.
							RunAsNonRoot: &[]bool{true}[0],
							// The memcached image does not use a non-zero numeric user as the default user.
							// Due to RunAsNonRoot field being set to true, we need to force the user in the
							// container to a non-zero numeric user. We do this using the RunAsUser field.
							// However, if you are looking to provide solution for K8s vendors like OpenShift
							// be aware that you cannot run under its restricted-v2 SCC if you set this value.
							RunAsUser:                &[]int64{1001}[0],
							AllowPrivilegeEscalation: &[]bool{false}[0],
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
						},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    resource.MustParse("100m"),
								"memory": resource.MustParse("100m"),
							},
							Requests: corev1.ResourceList{
								"cpu":    resource.MustParse("100m"),
								"memory": resource.MustParse("100m"),
							},
						},
						// TODO: ports probably not needed
						Ports: []corev1.ContainerPort{
							/*{
								ContainerPort: goop.Spec.ContainerPort,
								Name:          "goop",
							}*/
						},
						// TODO: define this command for goop: 'goop.Spec.Command' or something
						Command: []string{"echo \"GOOP!\""},
					}},
					Containers: []corev1.Container{{
						Image:           "gcr.io/google_containers/pause",
						Name:            "pause",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    resource.MustParse("50m"),
								"memory": resource.MustParse("50m"),
							},
							Requests: corev1.ResourceList{
								"cpu":    resource.MustParse("50m"),
								"memory": resource.MustParse("50m"),
							},
						},
						// Ensure restrictive context for the container
						// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
						SecurityContext: &corev1.SecurityContext{
							// WARNING: Ensure that the image used defines an UserID in the Dockerfile
							// otherwise the Pod will not run and will fail with "container has runAsNonRoot and image has non-numeric user"".
							// If you want your workloads admitted in namespaces enforced with the restricted mode in OpenShift/OKD vendors
							// then, you MUST ensure that the Dockerfile defines a User ID OR you MUST leave the "RunAsNonRoot" and
							// "RunAsUser" fields empty.
							RunAsNonRoot: &[]bool{true}[0],
							// The memcached image does not use a non-zero numeric user as the default user.
							// Due to RunAsNonRoot field being set to true, we need to force the user in the
							// container to a non-zero numeric user. We do this using the RunAsUser field.
							// However, if you are looking to provide solution for K8s vendors like OpenShift
							// be aware that you cannot run under its restricted-v2 SCC if you set this value.
							RunAsUser:                &[]int64{1001}[0],
							AllowPrivilegeEscalation: &[]bool{false}[0],
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
						},
						// TODO: ports probably not needed
						//Ports: []corev1.ContainerPort{},
					}},
				},
			},
		},
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(goop, dep, r.Scheme); err != nil {
		return nil, err
	}
	return dep, nil
}

// labelsForGoop returns the labels for selecting the resources
// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
func labelsForGoop(name string) map[string]string {
	imageTag := ""
	image, err := imageForGoop()
	if err == nil && len(strings.Split(image, ":")) > 1 {
		imageTag = strings.Split(image, ":")[1]
	}
	return map[string]string{
		"app.kubernetes.io/name":       "Goop",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/version":    imageTag,
		"app.kubernetes.io/part-of":    "goop-operator",
		"app.kubernetes.io/created-by": "controller-manager",
	}
}

// imageForGoop gets the Operand image which is managed by this controller
// from the GOOP_IMAGE environment variable defined in the config/manager/manager.yaml
func imageForGoop() (string, error) {
	var imageEnvVar = "GOOP_IMAGE"
	image, found := os.LookupEnv(imageEnvVar)
	if !found {
		return "", fmt.Errorf("Unable to find %s environment variable with the image", imageEnvVar)
	}
	return image, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GoopReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&goopv1alpha1.Goop{}).
		Complete(r)
}
