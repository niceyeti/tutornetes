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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/env"

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

// No goop finalization is needed, I kept this for demo purposes.
const goopFinalizer = "goop.example.com/finalizer"

// Definitions to manage status conditions
const (
	// typeAvailableGoop represents the status of the Deployment reconciliation
	typeAvailableGoop = "Available"
	// typeDegradedGoop represents the status used when the custom resource is deleted and the finalizer operations must to occur.
	typeDegradedGoop = "Degraded"
)

const (
	// The first state of the Goop object.
	initialized = "Initialized"
	// The daemonset 'job' is deployed.
	deployed = "Deployed"
	// The daemonset 'job' has completed.
	completed = "Completed"
	// The daemonset is being finalized and deleted.
	finalized = "Finalized"
)

// GoopReconciler reconciles a Goop object
type GoopReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var logLevel int

const (
	// The logging level.
	ENV_LOG_LEVEL = "LOG_LEVEL"
	// Noisier logs: 1 = verbose. 0 = logr package default.
	verbose = 1
)

func init() {
	logLevel = verbose
	if level, err := env.GetInt(ENV_LOG_LEVEL, verbose); err == nil {
		logLevel = level
	}
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

// Updates the goop status with the passed status-condition and info.
// We re-fetch the goop Custom Resource after updating the status so that
// we have the latest state of the resource on the cluster and avoid
// raising "the object has been modified, please apply your changes to the
// latest version and try again" which would re-trigger the reconciliation
// if we try to update it again in the following operation.
// API notes: there is much historical discussion per the patterns and
// practices for using Conditions. To clarify what I've learned:
//   - use Phase for high-level info, such as merely "Running" or "Failed".
//   - some use a "Ready" condition type and store the actual state within
//     the Condition's Reason field.
//   - I'm going to use Condition's Reason to define state, unlike Pods.
//     The Type will always be Available or Degraded, simply because this
//     is the convention of the generated controller code.
//     Conditions have some feature bloat covering requirements I don't have,
//     such as substates (true, false, unknown), don't be distracted by these.
//
// The number of states and substatuses one attempts to encode into Conditions
// can be seen as a code smell, and overall a distraction of state machine development.
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

// HandlerFunc has the same return signature as Reconcile for setting up chained handlers.
// I am not incredibly satisfied with this pattern, in retrospect, and since
// implementing it I have seen cleaner patterns like the IBM operator: https://github.com/IBM/operator-sample-go
// The downside is merely readability. The handler chain pass-go/no-go style
// code can actually be somewhat confusing when implementing and debugging the
// operator, but worth considering/reworking.
type HandlerFunc func(context.Context, *logr.Logger, *goopRequest) (ctrl.Result, error)

// goopRequest merely encapsulates the request to allow
// handlers to update these fields before passing them along.
type goopRequest struct {
	req  *ctrl.Request
	goop *goopv1alpha1.Goop
}

func isInitialState(goop *goopv1alpha1.Goop) bool {
	return len(goop.Status.Conditions) == 0
}

func isState(goop *goopv1alpha1.Goop, status string) bool {
	return len(goop.Status.Conditions) > 0 &&
		goop.Status.Conditions[len(goop.Status.Conditions)-1].Reason == status
}

func isCompleteState(goop *goopv1alpha1.Goop, status string) bool {
	return isState(goop, status) &&
		goop.Status.Conditions[len(goop.Status.Conditions)-1].Status == metav1.ConditionTrue
}

func isIncompleteState(goop *goopv1alpha1.Goop, status string) bool {
	return isState(goop, status) &&
		goop.Status.Conditions[len(goop.Status.Conditions)-1].Status != metav1.ConditionTrue
}

func (r *GoopReconciler) handleGoop(req *ctrl.Request, next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, log *logr.Logger, _ *goopRequest) (ctrl.Result, error) {
		gr := &goopRequest{req: req, goop: nil}
		goop, err := r.fetchGoop(ctx, req.NamespacedName)
		if err != nil {
			log.Error(err, "Failed to get goop")
			return ctrl.Result{}, err
		}
		if goop == nil {
			// If the custom resource is not found then, it usually means that it
			// was deleted or not created, so stop reconciliation.
			log.Info("goop resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		gr.goop = goop

		obj, _ := json.MarshalIndent(goop, "", " ")
		log.Info("\nGoop:\n>>>" + string(obj) + "<<<\n")

		return next(ctx, log, gr)
	}
}

func (r *GoopReconciler) ensureInitialization(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, log *logr.Logger, gr *goopRequest) (ctrl.Result, error) {
		// Set the status as Unknown when no status are available
		if isInitialState(gr.goop) {
			goop, err := r.setStatusCondition(
				ctx,
				gr.goop,
				metav1.Condition{
					Type:    initialized,
					Status:  metav1.ConditionUnknown,
					Reason:  initialized,
					Message: "Starting reconciliation"})
			if err != nil {
				log.Error(err, "Failed to update goop status")
				return ctrl.Result{}, err
			}
			// Ensure post-conditions for this state: first condition set and retrieved
			if len(goop.Status.Conditions) == 0 {
				log.Info("Requeueing until first condition is initialized")
				return ctrl.Result{Requeue: true}, nil
			}
			gr.goop = goop
		}

		log.Info("Calling next() from ensureInitialization")
		return next(ctx, log, gr)
	}
}

func (r *GoopReconciler) ensureFinalizer(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, log *logr.Logger, gr *goopRequest) (ctrl.Result, error) {
		// Adds a finalizer, then we can define some operations that should occur
		// before the custom resource deletion.
		// TODO (Jesse): could add finalizer reqs for exercise. Finalizers are for custom behavior/state;
		// but note that the daemonsets created for a Goop job are deleted automatically
		// since the api-server knows that they are owned by the Goop object. Thus currently
		// there is nothing else to clean up. Best practice is probably to rely on ownership
		// where possible and avoid finalizer code bloat.
		// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
		if isState(gr.goop, initialized) && !controllerutil.ContainsFinalizer(gr.goop, goopFinalizer) {
			log.Info("Adding Finalizer for goop")
			if ok := controllerutil.AddFinalizer(gr.goop, goopFinalizer); !ok {
				log.Error(errors.New("failed to update finalizer"), "Failed to add finalizer into the custom resource")
				return ctrl.Result{Requeue: true}, nil
			}

			if err := r.Update(ctx, gr.goop); err != nil {
				log.Error(err, "Failed to update custom resource to add finalizer")
				return ctrl.Result{}, err
			}

			// Re-fetch the goop Custom Resource after updating the status so that
			// we have the latest state of the resource on the cluster and we will avoid
			// raising "the object has been modified, please apply your changes to the
			// latest version and try again" which would re-trigger the reconciliation
			// if we try to update it again in the following operations.
			if err := r.Get(ctx, gr.req.NamespacedName, gr.goop); err != nil {
				log.Error(err, "Failed to re-fetch goop")
				return ctrl.Result{}, err
			}
		}

		// After every initial thing is complete, most of which does not involve
		// cluster modification, mark initialization as completed.
		if isIncompleteState(gr.goop, initialized) {
			// The goop is now fully initialized, ready for creation.
			goop, err := r.setStatusCondition(
				ctx,
				gr.goop,
				metav1.Condition{
					Type:    initialized,
					Status:  metav1.ConditionTrue,
					Reason:  initialized,
					Message: "Initialization completed"})
			if err != nil {
				log.Error(err, "Failed to update goop status")
				return ctrl.Result{}, err
			}
			gr.goop = goop
		}

		log.Info("Calling next() from ensureFinalizer")
		return next(ctx, log, gr)
	}
}

// handleDeletion checks if the Goop object is to be deleted, and if so,
// performs deletion tasks and aborts other handlers.
func (r *GoopReconciler) handleDeletion(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, log *logr.Logger, gr *goopRequest) (ctrl.Result, error) {
		// FUTURE: aborting when the object is in the completed finalized state is a
		// perfunctory sink, but this omits other possible reasons or causes for reconciliation
		// after finalization.
		if isCompleteState(gr.goop, finalized) {
			log.Info("Object is already finalized, aborting handling.")
			// Stop reconciliation as the item is being deleted
			return ctrl.Result{}, nil
		}

		// Check if the Goop instance is marked to be deleted,
		// indicated by the deletion timestamp being set.
		isGoopMarkedToBeDeleted := gr.goop.GetDeletionTimestamp() != nil
		if isGoopMarkedToBeDeleted {
			if controllerutil.ContainsFinalizer(gr.goop, goopFinalizer) {
				// Perform all operations required before remove the finalizer and allow
				// the Kubernetes API to remove the custom resource.
				// TODO: I need to implement this
				r.doFinalizerOperationsForGoop(gr.goop)

				// TODO(user): If you add operations to the doFinalizerOperationsForGoop method
				// then you need to ensure that all worked fine before deleting and updating the Downgrade status
				// otherwise, you should requeue here.

				log.Info("Removing Finalizer for goop after successfully performing operations")
				if ok := controllerutil.RemoveFinalizer(gr.goop, goopFinalizer); !ok {
					log.Error(errors.New("failed to remove finalizer"), "Failed to remove finalizer for Goop")
					return ctrl.Result{Requeue: true}, nil
				}

				if err := r.Update(ctx, gr.goop); err != nil {
					log.Error(err, "Failed to re-fetch goop")
					return ctrl.Result{}, err
				}

				return ctrl.Result{Requeue: true}, nil
			}

			// FUTURE: I met with a ton of api nastiness when attempting to mark the goop object's
			// state as 'finalized' without raising 'object has been modified' errors, not matter
			// whether or not the view of the goop object was up to date or not, nor the correct
			// sequence of the above finalizer-removal code. The 'finalized' state may not be necessary,
			// since the object will simply be deleted (and unqueryable) after the finalizer is removed
			// above, making this unreachable code.
			_, err := r.setStatusCondition(
				ctx,
				gr.goop,
				metav1.Condition{
					Type:    finalized,
					Status:  metav1.ConditionTrue,
					Reason:  finalized,
					Message: fmt.Sprintf("Finalization for goop %s completed", gr.goop.Name)})
			if err != nil {
				log.Error(err, "Failed to update goop status")
				return ctrl.Result{}, err
			}

			// TODO: smooth out the corners of deletion. If deletion/finalization is completed,
			// then this should possibly return nil,nil, with the requirement that the daemonset
			// is completely deleted (and its result perhaps recorded in the Goop object). This
			// may be accomplished simply, by requeueing after a few seconds and verifying
			// daemonset deletion before completing deletion of the goop object.
			log.Info("Aborting from delete with: ctrl.Result{}, nil. There should be no further calls.")
			return ctrl.Result{}, nil
		}

		log.Info("Calling next() from handleDeletion")
		return next(ctx, log, gr)
	}
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

func (r *GoopReconciler) handleCreation(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, log *logr.Logger, gr *goopRequest) (ctrl.Result, error) {
		// Only create daemonset after completing initialization
		if isCompleteState(gr.goop, initialized) {
			log.Info("Checking daemonset in creation")
			// Check if the daemonset already exists, if not create a new one.
			// NOTE: querying things like daemonsets requires RBAC permission by the
			// service-account to do so. Failing to do so gives errors such as:
			/// 'system:serviceaccount:goop-system:goop-controller-manager" cannot list resource "daemonsets" in API group "apps".
			// The solution is simply to add the appropriate permissions via roles and rolebindings.
			ds := &appsv1.DaemonSet{}
			err := r.Get(ctx, gr.req.NamespacedName, ds)
			if err != nil && !apierrors.IsNotFound(err) {
				log.Error(err, fmt.Sprintf("Unknown error getting Daemonset for goop %s", gr.req.NamespacedName))
				return ctrl.Result{}, err
			}

			if err != nil && apierrors.IsNotFound(err) {
				// Define a new deployment
				dsSpec, err := r.daemonsetForGoop(gr.goop)
				if err != nil {
					log.Error(err, "Failed to define new Daemonset resource for Goop")
					return ctrl.Result{}, err
				}

				log.Info("Creating a new job Daemonset",
					"Daemonset.Namespace", dsSpec.Namespace, "Daemonset.Name", dsSpec.Name)

				if err := r.Create(ctx, dsSpec); err != nil {
					log.Error(err, "Failed to create new Daemonset",
						"Daemonset.Namespace", dsSpec.Namespace, "Daemonset.Name", dsSpec.Name)
					return ctrl.Result{}, err
				}

				meta.SetStatusCondition(
					&gr.goop.Status.Conditions,
					metav1.Condition{
						Type:    deployed,
						Status:  metav1.ConditionTrue,
						Reason:  deployed,
						Message: fmt.Sprintf("Daemonset created for goop: %s", gr.goop.Name)})

				if err := r.Status().Update(ctx, gr.goop); err != nil {
					log.Error(err, "Failed to update goop status in creation")
					return ctrl.Result{}, err
				}

				log.Info("Requeueing to check for daemonset completion in 3 seconds")
				return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
			}
		}

		log.Info("Calling next from handleCreation")
		return next(ctx, log, gr)
	}
}

// FUTURE: this is just a perfunctory example of checking the properties of
// some k8s object: daemonset, job, etc. The internals of this implementation
// are one's cluster-wide requirements; here, I merely check that the daemonset is
// fully available and ready, which isn't a robust check of whether or not jobs
// completed, aka their tool processes returned 0, which would have to be done by
// checking the exit codes of all the init containers. This returns true before
// the daemonset containers are even running, which would require a separate query.
func isJobCompleted(ds *appsv1.DaemonSet) bool {
	return ds.Status.DesiredNumberScheduled == ds.Status.NumberReady &&
		ds.Status.DesiredNumberScheduled == ds.Status.NumberAvailable
}

func (r *GoopReconciler) handleCompletion(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, log *logr.Logger, gr *goopRequest) (ctrl.Result, error) {
		if isCompleteState(gr.goop, deployed) {
			log.Info("Checking for completion")

			ds := &appsv1.DaemonSet{}
			err := r.Get(ctx, gr.req.NamespacedName, ds)
			if err != nil && !apierrors.IsNotFound(err) {
				log.Error(err, "Failed to get Daemonset")
				return ctrl.Result{}, err
			}

			if !isJobCompleted(ds) {
				// Abort handlers and requeue to monitor job-completion.
				log.Info("Requeueing to monitor for completion in 3 seconds")
				return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
			}

			goop, err := r.setStatusCondition(
				ctx,
				gr.goop,
				metav1.Condition{
					Type:    completed,
					Status:  metav1.ConditionTrue,
					Reason:  completed,
					Message: fmt.Sprintf("Daemonset for goop completed successfully: %s", gr.goop.Name)})
			if err != nil {
				log.Error(err, "Failed to update goop in completion")
				return ctrl.Result{}, err
			}

			gr.goop = goop
		}

		log.Info("Calling next from handleCompletion")
		return next(ctx, log, gr)
	}
}

// TODO:
// - goop webhook would be cool

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
	log := log.FromContext(ctx).V(logLevel)

	// TODO: [Jesse] The Goop operator currently has no update logic, primarily
	// because its features are not complete (for demo purposes). Otherwise this
	// would need to be done here by diffing goop.Spec and found.Spec.

	// FUTURE: this handler pattern could be reworked. This is a simple sequential
	// set of handlers, each modifying the passed-along parameters or aborting the
	// chain, forming a linear chain of responsibility. Other patterns are possible,
	// such as a functional graph of chains, rather than a straight line sequence of
	// handlers. No need here though. There are a lot of code smells with recursive
	// handlers, in terms of readability. It just seems over engineered. This
	// pattern was implemented merely as a first-pass.
	return r.handleGoop(&req,
		r.ensureInitialization(
			r.ensureFinalizer(
				r.handleDeletion(
					r.handleCreation(
						r.handleCompletion(nilHandler))))),
	)(ctx, &log, nil)
}

func nilHandler(ctx context.Context, log *logr.Logger, gr *goopRequest) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// Returns the daemonset that implements distributed jobs as init-containers.
// Deploying Jobs to nodes such that a Job runs independently on every node
// can be done using a Daemonset that runs the job logic in init containers.
// Its very likely a more modern pattern exists, perhaps using Jobs with topology
// spread constraints. I have no idea if the daemonset pattern obtains all desired
// lifecycle requirements for job-like behavior, or if this is a hack. This is
// for demo purposes.
// This yaml pattern comes from the github issue reply:
// https://github.com/kubernetes/kubernetes/issues/36601
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
