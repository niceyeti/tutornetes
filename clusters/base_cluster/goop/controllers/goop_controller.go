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
	"encoding/json"
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
	"github.com/go-logr/logr"
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

// fetchGoop retrieves the Goop object from the k8s api using its namespaced name.
func (r *GoopReconciler) fetchGoop(
	ctx context.Context,
	namespacedName types.NamespacedName,
) (*goopv1alpha1.Goop, error) {
	goop := &goopv1alpha1.Goop{}
	if err := r.Get(ctx, namespacedName, goop); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation.
			return nil, nil
		}
		// Error reading the object - caller should requeue the request.
		return nil, err
	}
	return goop, nil
}

// Updates the goop status with the passed status and info.
// We re-fetch the goop Custom Resource after updating the status so that
// we have the latest state of the resource on the cluster and avoid
// raising "the object has been modified, please apply your changes to the
// latest version and try again" which would re-trigger the reconciliation
// if we try to update it again in the following operation.
func (r *GoopReconciler) setStatusCondition(
	ctx context.Context,
	goop *goopv1alpha1.Goop,
	condition metav1.Condition,
) (*goopv1alpha1.Goop, error) {
	meta.SetStatusCondition(&goop.Status.Conditions, condition)
	if err := r.Status().Update(ctx, goop); err != nil {
		return nil, err
	}

	namespacedName := types.NamespacedName{
		Namespace: goop.Namespace,
		Name:      goop.Name,
	}
	updated := &goopv1alpha1.Goop{}
	if err := r.Get(ctx, namespacedName, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

// TODO: I'm still working out these handler signatures. The only issue is deciding
// how the handlers can update the goop object which gets passed to subsequent
// handlers, since many handlers will modify and subsequently request the latest
// copy of the goop object. I'm going to take a walk for lunch and think about
// a clean way to do this; code is getting kludgy. The 'simple' will fall in place
// soon; time to hammer the delete key.

// HandlerFunc is a handler with the following signature.
// If error is non-nil, then
// If *ctrl.Result is non-nil, it means no error occurred, but Reconciliation should stop.
type HandlerFunc func(context.Context, *logr.Logger, *goopv1alpha1.Goop) (*ctrl.Result, error)

// StateHandler defines its own internal handling and state, then calls the passed
// handlers, whose success is presumably dependent on it and any predecessor handlers.
// The expected implementation pattern is that for each HandlerFunc, if ctrl.Result or error are
// non-nil, then return these immediately and abort subsequent handlers. This allows defining
// handler dependencies in a sequence at the caller level:
//
//	SomeHandler( GetFoo, CreateFoo, UpdateFoo, DeleteFoo)
type StateHandler func(...HandlerFunc) HandlerFunc

func (r *GoopReconciler) HandleGoop(req ctrl.Request, handlers ...HandlerFunc) HandlerFunc {
	return func(ctx context.Context, log *logr.Logger, _ *goopv1alpha1.Goop) (*ctrl.Result, error) {
		goop, err := r.fetchGoop(ctx, req.NamespacedName)
		if err != nil {
			// Error reading the object - requeue the request.
			log.Error(err, "Failed to get goop")
			return &ctrl.Result{}, err
		}
		if goop == nil {
			// If the custom resource is not found then, it usually means that it
			// was deleted or not created, so stop reconciliation.
			log.Info("goop resource not found. Ignoring since object must be deleted")
			return &ctrl.Result{}, nil
		}

		obj, _ := json.MarshalIndent(goop, "", " ")
		log.Info("\nGoop:\n>>>" + string(obj) + "<<<\n")

		return handleAll(ctx, log, goop, handlers...)
	}
}

func handleAll(
	ctx context.Context,
	log *logr.Logger,
	goop *goopv1alpha1.Goop,
	handlers ...HandlerFunc,
) (result *ctrl.Result, err error) {
	for i, _ := range handlers {
		result, err = handlers[i](ctx, log, goop)
		if err != nil || result != nil {
			return
		}
	}
	return
}

func (r *GoopReconciler) EnsureInitialStatus(handlers ...HandlerFunc) HandlerFunc {
	return func(ctx context.Context, log *logr.Logger, goop *goopv1alpha1.Goop) (*ctrl.Result, error) {
		// Set the status as Unknown when no status are available
		if len(goop.Status.Conditions) == 0 {
			goop, err := r.setStatusCondition(
				ctx,
				goop,
				metav1.Condition{
					Type:    typeAvailableGoop,
					Status:  metav1.ConditionUnknown,
					Reason:  "Reconciling",
					Message: "Starting reconciliation"})
			if err != nil {
				log.Error(err, "Failed to update goop status")
				return &ctrl.Result{}, err
			}
		}

		return nil, nil
	}
}

// TODO:
// - fix state pattern and double creation
// - webhook would be cool

// These markers cause the appropriate RBAC resources to be created for the controller,
// such as clusterroles and clusterrolebindings. Add as needed to interact with other
// kubernetes resources within the controller. For example, I added the daemonset
// roles so the controller can create and manipulate daemonsets.
// See 'RBAC markers': https://book.kubebuilder.io/cronjob-tutorial/controller-overview.html
//+kubebuilder:rbac:groups=goop.example.com,resources=goops,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=goop.example.com,resources=goops/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=goop.example.com,resources=goops/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
// Returns: see the Result{} definition.
func (r *GoopReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// TODO: I worked through the requirements of Reconcile just to understand some
	// of the (extremely underdocumented!) requirements. But on reflection I think
	// this code could be rewritten as chained handler functions, much like other
	// Golang pipeline patterns whereby functions may either continue to handle
	// a request for which they can make progress, or abort and return an error.
	// Titmus' Cloud Native go has examples; see the CircuitBreaker pattern, et al.
	// There a few repetititve patterns:
	// 1) modify/update the Goop fields, then re-get it to avoid raising the error "the object
	//    has been modified, please apply your changes to the latest version and try again".
	//    These are low-level utilities returning errors.
	// 2) The high-level pattern of chaining together mutators which either mutate
	//    state and call the next handler, or abort handling. These are correlated
	//    (return?) with the (ctrl.Result{},err) types below (and logging?).
	// (2) could be written multiple ways, and I am unsure what would be the most readable,
	// vs. mere procedural code like below. This is a good problem for code exercise
	// since it encompasses a bunch of reqs for readability and testing. Should probably
	// look at other controller implementations for best-practices.

	// Fetch the Goop instance
	// In part, the purpose is to check if the Custom Resource for the Kind Goop
	// is applied on the cluster, and if not we return nil to stop the reconciliation.

	//type Handler func(context.Context, *logr.Logger, *goopv1alpha1.Goop) (ctrl.Result, error)
	result, err := r.HandleGoop(
		req,
		r.EnsureInitialStatus,
		r.EnsureFinalizer,
		r.HandleDeletion,
		r.HandleCreation(
			r.MonitorCompletion,
		),
		// r.HandleUpdate
	)(ctx, &log, nil)
	return *result, err

	// Adds a finalizer, then we can define some operations that should occur
	// before the custom resource deletion.
	// TODO (Jesse): could add finalizer reqs for exercise. Finalizers are for custom behavior/state;
	// but note that the daemonsets created for a Goop job deleted automatically
	// since the api-server knows that they are owned by the Goop object. Thus currently
	// there is nothing else to clean up. Best practice is probably to rely on ownership
	// and avoid finalizer code bloat.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	if !controllerutil.ContainsFinalizer(goop, goopFinalizer) {
		log.Info("Adding Finalizer for goop")
		if ok := controllerutil.AddFinalizer(goop, goopFinalizer); !ok {
			log.Error(err, "Failed to add finalizer into the custom resource")
			return ctrl.Result{Requeue: true}, nil
		}

		if err = r.Update(ctx, goop); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}

		// Re-fetch the goop Custom Resource after updating the status so that
		// we have the latest state of the resource on the cluster and we will avoid
		// raising "the object has been modified, please apply your changes to the
		// latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations.
		if err := r.Get(ctx, req.NamespacedName, goop); err != nil {
			log.Error(err, "Failed to re-fetch goop")
			return ctrl.Result{}, err
		}
	}

	// Check if the Goop instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isGoopMarkedToBeDeleted := goop.GetDeletionTimestamp() != nil
	if isGoopMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(goop, goopFinalizer) {
			log.Info("Performing Finalizer Operations for goop before deleting")

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

			log.Info("Removing Finalizer for goop after successfully perform the operations")
			if ok := controllerutil.RemoveFinalizer(goop, goopFinalizer); !ok {
				log.Error(err, "Failed to remove finalizer for Goop")
				return ctrl.Result{Requeue: true}, nil
			}

			if err := r.Update(ctx, goop); err != nil {
				log.Error(err, "Failed to remove finalizer for goop")
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
			meta.SetStatusCondition(
				&goop.Status.Conditions,
				metav1.Condition{
					Type:    typeAvailableGoop,
					Status:  metav1.ConditionFalse,
					Reason:  "Reconciling",
					Message: fmt.Sprintf("Failed to create Daemonset for the custom resource (%s): (%s)", goop.Name, err)})

			if err := r.Status().Update(ctx, goop); err != nil {
				log.Error(err, "Failed to update goop status")
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

	// TODO: [Jesse] The Goop operator currently has no update logic, primarily
	// because its features are not complete (for demo purposes). Otherwise this
	// would need to be done here by diffing goop.Spec and found.Spec.

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

func (r *GoopReconciler) doFinalizerOperationsForGoop(cr *goopv1alpha1.Goop) {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.

	// Note: It is not recommended to use finalizers to delete resources which are
	// created and managed in the reconciliation. These ones, such as the Daemonset created on this reconcile,
	// are defined as depended of the custom resource. See the method ctrl.SetControllerReference.
	// to set the ownerRef which means that the Deployment will be deleted by the Kubernetes API.
	// More info: https://kubernetes.io/docs/tasks/administer-cluster/use-cascading-deletion/
}

// Returns the daemonset that implements jobs using init-containers.
// Deploying Jobs to nodes such that a Job runs independently on every node
// can be done using a Daemonset that runs the job logic in init containers.
// Its very likely a more modern pattern exists, perhaps using Jobs with topology
// spread constraints. I have no idea if the daemonset pattern obtains all desired
// lifecycle requirements for job-like behavior, or if this is a hack.
// This yaml pattern comes from the github issue reply here: https://github.com/kubernetes/kubernetes/issues/36601
func (r *GoopReconciler) daemonsetForGoop(
	goop *goopv1alpha1.Goop,
) (*appsv1.DaemonSet, error) {
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
						Name:            "goop-job", // Note: name can only be alphanumeric and '-'
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
								"memory": resource.MustParse("128Mi"),
							},
							Requests: corev1.ResourceList{
								"cpu":    resource.MustParse("10m"),
								"memory": resource.MustParse("64Mi"),
							},
						},
						// TODO: define this command for goop: 'goop.Spec.Command' or something
						Command: []string{"sleep", "11"},
					}},
					Containers: []corev1.Container{{
						Image:           "k3d-devregistry:5000/gcr.io/google_containers/pause:latest",
						Name:            "pause",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    resource.MustParse("50m"),
								"memory": resource.MustParse("128Mi"),
							},
							Requests: corev1.ResourceList{
								"cpu":    resource.MustParse("10m"),
								"memory": resource.MustParse("64Mi"),
							},
						},
						// TODO: [Jesse] None of these requirements are fully fleshed-out for a full-fledged
						// distributed-job CRD, which would depend completely on its features: node access?
						// root access? et cetera.
						// Ensure restrictive context for the container
						// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
						SecurityContext: &corev1.SecurityContext{
							// WARNING: Ensure that the image used defines an UserID in the Dockerfile
							// otherwise the Pod will not run and will fail with "container has runAsNonRoot and image has non-numeric user"".
							// If you want your workloads admitted in namespaces enforced with the restricted mode in OpenShift/OKD vendors
							// then, you MUST ensure that the Dockerfile defines a User ID OR you MUST leave the "RunAsNonRoot" and
							// "RunAsUser" fields empty.
							RunAsNonRoot: &[]bool{true}[0],
							// The memcached image did not use a non-zero numeric user as the default user.
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
					}},
				},
			},
		},
	}

	// Set the ownerRef for the Daemonset
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
	// TODO: this is not a valid way to parse the image tag for image names prefixed by a registry
	// address and no image tag: k3d-devregistry:5000/busybox -> "5000/busybox".
	tokens := strings.Split(image, ":")
	if err == nil && len(tokens) > 1 {
		imageTag = tokens[len(tokens)-1]
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
