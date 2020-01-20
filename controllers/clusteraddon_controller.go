/*

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
	"github.com/mitchellh/hashstructure"
	"github.com/mmlt/operator-addons/internal/cluster"
	"github.com/mmlt/operator-addons/internal/repogit"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/mmlt/operator-addons/api/v1alpha1"
)

// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers
const finalizerName = "clusteraddon.clusterops.mmlt.nl"

// RequeueDurection is the interval with which external resources are checked for changes.
const RequeueDuration = 5 * time.Minute

// ClusterAddonReconciler reconciles a ClusterAddon object.
type ClusterAddonReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder

	// Repos maps the repository url's to repo data.
	Repos map[string]*repogit.Repo
}

// +kubebuilder:rbac:groups=clusterops.mmlt.nl,resources=clusteraddons,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusterops.mmlt.nl,resources=clusteraddons/status,verbs=get;update;patch

// Reconcile attempts to apply desired state.
func (r *ClusterAddonReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("clusteraddon", req.NamespacedName.String())
	log.V(1).Info("Reconcile start")
	defer log.V(1).Info("Reconcile end")

	// TODO add Policy checks

	// Get ClusterAddon resource.
	clusterAddon := &v1alpha1.ClusterAddon{}
	if err := r.Get(ctx, req.NamespacedName, clusterAddon); err != nil {
		log.V(1).Info("Unable to get ClusterAddon CR", "err", err)
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// Get Cluster object.
	cl, err := r.clusterFor(req.Namespace+"-"+req.Name, &clusterAddon.Spec.Target, log)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("clusterFor: %w", err)
	}

	// Handle finalizer delete.
	if !clusterAddon.ObjectMeta.DeletionTimestamp.IsZero() {
		// The CR is (being) deleted.
		if containsString(clusterAddon.ObjectMeta.Finalizers, finalizerName) {
			// Finalizer is present, proceed with delete.
			err = r.delete(cl, log)
			if err != nil {
				// Delete failed (but will be retried).
				return ctrl.Result{}, fmt.Errorf("delete (will retry): %w", err)
			}

			// Deletion succeeded, remove our finalizer.
			clusterAddon.ObjectMeta.Finalizers = removeString(clusterAddon.ObjectMeta.Finalizers, finalizerName)
			err = r.Update(context.Background(), clusterAddon)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("finalizer remove: %w", err)
			}
			log.V(1).Info("Finalizer removed")
		}
		// No point in continuing reconciliation when the CR is being deleted.
		return ctrl.Result{}, nil
	}
	if !containsString(clusterAddon.ObjectMeta.Finalizers, finalizerName) {
		// No finalizer registered yet so register it.
		clusterAddon.ObjectMeta.Finalizers = append(clusterAddon.ObjectMeta.Finalizers, finalizerName)
		err = r.Update(context.Background(), clusterAddon)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("finalizer add: %w", err)
		}
		log.V(1).Info("Finalizer added")
	}

	// Do the actual reconciliation work.
	status, err := r.createOrUpdate(cl, clusterAddon, log)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Status from createOrUpdate", "status", status)

	changed := calculateStatus(clusterAddon, status, time.Now())
	if !changed {
		return ctrl.Result{RequeueAfter: RequeueDuration}, nil
	}

	// Write updated clusterAddon/status.
	ctx = context.Background()
	err = r.Status().Update(ctx, clusterAddon)

	//TODO remove
	if err != nil {
		log.Error(err, "Status update")
	}
	log.V(2).Info("Status updated")

	//TODO don't return err==nil when condition is false
	return ctrl.Result{RequeueAfter: RequeueDuration}, err
}

// CreateOrUpdate creates or updates resources as defined in clusterAddon in the target cluster.
// Errors are mapped to status fields/conditions when possible.
// Only the errors that can't be mapped are returned.
func (r *ClusterAddonReconciler) createOrUpdate(
	cl *cluster.Cluster,
	clusterAddon *v1alpha1.ClusterAddon,
	log logr.Logger) (*v1alpha1.ClusterAddonStatus, error) {

	log.V(1).Info("CreateOrUpdate")

	status := &v1alpha1.ClusterAddonStatus{}

	// Make sure the target cluster is reachable before continuing.
	if !cl.Ping() {
		status.Conditions = append(status.Conditions, condition(v1alpha1.ClusterAddonTargetOk, false, "Ping", "Ping failed"))
		return status, nil
	}

	// Get current state from target cluster.
	currentState, err := getState(cl)
	if err != nil {
		re, m := "Error", err.Error()
		// Try to get a more specific reason and message.
		if e, ok := err.(*apierrors.StatusError); ok {
			re, m = string(e.ErrStatus.Reason), e.ErrStatus.Message
		}
		status.Conditions = append(status.Conditions, condition(v1alpha1.ClusterAddonTargetOk, false, re, m))
		return status, nil
	}
	status.Conditions = append(status.Conditions, condition(v1alpha1.ClusterAddonTargetOk, true, "", ""))

	// Iterate over Sources.
	var hasStateChange bool
	for n, src := range clusterAddon.Spec.Sources {
		log := log.WithValues("source", n)

		// Get repo.
		repo, err := r.repoFor(&src, log)
		if err != nil {
			//TODO DRY? status.Conditions = r.xxx(status.Conditions, v1alpha1.ClusterAddonSourceOk, "Get source", err)
			//TODO Keep condition, event and log together?
			// logRecordCondition(...)
			status.Conditions = append(status.Conditions, condition(v1alpha1.ClusterAddonSourceOk, false, "Error", err.Error()))
			log.Error(err, "Get source")
			r.recorder.Event(clusterAddon, corev1.EventTypeWarning, "UpdateFailed", fmt.Sprintf("Update %s failed", n))
			log.Info(fmt.Sprintf("Update %s failed", n))
			continue
		}
		status.Conditions = append(status.Conditions, condition(v1alpha1.ClusterAddonSourceOk, true, "", ""))
		log.V(2).Info("Get source")

		// Check for changes in repo or action.
		repoSHA, _ := repo.SHAlocal()
		actionHash, err := hashstructure.Hash(src.Action, nil)
		if err != nil {
			return status, err
		}
		if repoSHA == currentState.Sources[n].RepoSHA && actionHash == currentState.Sources[n].ActionHash {
			// no changes.
			continue
		}

		// Perform action.
		env := []string{"REPODIR=" + repo.Dir()} //TODO add RECONCILE=CREATE_OR_UPDATE ?
		err = cl.RunShell(src.Action.Cmd, src.Action.Values, env)
		if err != nil {
			status.Conditions = append(status.Conditions, condition(v1alpha1.ClusterAddonActionOk, false, "Error", err.Error()))
			log.Error(err, "Action")
			r.recorder.Event(clusterAddon, corev1.EventTypeWarning, "UpdateFailed", fmt.Sprintf("Update '%s' failed", n))
			log.Info(fmt.Sprintf("Update '%s' failed", n))
			continue
		}
		status.Conditions = append(status.Conditions, condition(v1alpha1.ClusterAddonActionOk, true, "", ""))

		//TODO DRY event+log combination
		r.recorder.Event(clusterAddon, corev1.EventTypeNormal, "Update", fmt.Sprintf("Update '%s' successful", n))
		log.Info(fmt.Sprintf("Update '%s' successful", n))

		// Update current state
		currentState.Sources[n] = sourceState{ActionHash: actionHash, RepoSHA: repoSHA}
		hasStateChange = true
	}

	status.Conditions = append(status.Conditions, condition(v1alpha1.ClusterAddonSynced, true, "", ""))

	if !hasStateChange {
		log.V(1).Info("Current state has not changed")
		return status, nil
	}

	// When current state has changed it needs to written back to the target cluster.
	// TODO state is only added to, old keys aren't removed.
	err = putState(cl, currentState)
	if err != nil {
		// Performed Action but failed to write the new state.
		// Return an error meaning status can not be trusted and the reconciliation should be retried soon.
		//
		// An alternative would be to log an error and return nil (meaning status should be up written to the CR)
		// In that case the next reconciliation would be some time in the future.
		// Before that time the cluster configmap will contain an outdated state.
		return status, err
	}

	return status, nil
}

// Delete is the last call before the CR is deleted.
func (r *ClusterAddonReconciler) delete(cl *cluster.Cluster, log logr.Logger) error {
	//TODO add RECONCILE=DELETE ?
	log.V(1).Info("Delete")
	return nil
}

// CalculateStatus updates clusterAddon with status and returns true when changes have been made to clusterAddon.
func calculateStatus(clusterAddon *v1alpha1.ClusterAddon, status *v1alpha1.ClusterAddonStatus, timeNow time.Time) bool {
	// Steps:
	// 1. Deduplicate status.Conditions
	// 2. Merge status.Conditions into CR Status.
	// 3. Update Status fields based on conditions.

	// Create a status condition map with deduplicated status.Conditions
	// The result is a logical AND of all instances of a type.
	// Reason is the first non-empty reason.
	// Message is a concatenation of all messages.
	scm := map[v1alpha1.ClusterAddonConditionType]v1alpha1.ClusterAddonCondition{}
	for _, c := range status.Conditions {
		if _, ok := scm[c.Type]; !ok {
			scm[c.Type] = c
			continue
		}
		org := scm[c.Type]
		if org.Status == metav1.ConditionFalse {
			continue
		}
		org.Status = c.Status
		if org.Reason == "" {
			org.Reason = c.Reason
		}
		org.Message = org.Message + ", " + c.Message
		scm[c.Type] = org
	}

	// Merge status.Conditions into CR Status.
	// When the 'old' condition matches the 'new' condition nothing changes.
	// When an 'old' condition exists but no 'new' condition the condition becomes 'unknown'.
	// In all other cases the condition becomes the 'new' condition.
	var hasChanged bool
	tn := metav1.Time{Time: timeNow}
	for _, nc := range scm {
		// Find old condition of same type as nc.
		var oc *v1alpha1.ClusterAddonCondition
		for i, c := range clusterAddon.Status.Conditions {
			if c.Type == nc.Type {
				oc = &clusterAddon.Status.Conditions[i]
				break
			}
		}

		if oc == nil {
			// No old condition to update so add new condition.
			nc.LastTransitionTime = tn
			clusterAddon.Status.Conditions = append(clusterAddon.Status.Conditions, nc)
			hasChanged = true
			continue
		}

		if oc.Status == nc.Status {
			// No transition (but reason or message might contain new information)
			if oc.Reason != nc.Reason || oc.Message != nc.Message {
				oc.Reason = nc.Reason
				oc.Message = nc.Message
				hasChanged = true
			}
			continue
		}

		// Update
		oc.Status = nc.Status
		oc.Reason = nc.Reason
		oc.Message = nc.Message
		oc.LastTransitionTime = tn
		hasChanged = true
	}

	/*TODO
	// Old condition types that have no corresponding nc get their status set to unknown.
	for i, oc := range clusterAddon.Status.Conditions {
		_, ok := scm[oc.Type]
		if !ok && oc.Status != metav1.ConditionUnknown {
			oc.Status = metav1.ConditionUnknown
			oc.Reason = ""
			oc.Message = ""
			oc.LastTransitionTime = tn
			clusterAddon.Status.Conditions[i] = oc
			hasChanged = true
		}
	}*/

	// Copy Synced condition to status.synced.
	for _, c := range clusterAddon.Status.Conditions {
		if c.Type == v1alpha1.ClusterAddonSynced {
			clusterAddon.Status.Synced = c.Status
			break
		}
	}

	return hasChanged
}

// ClusterFor get or creates a new cluster object for a ClusterAddonTarget.
func (r *ClusterAddonReconciler) clusterFor(name string, target *v1alpha1.ClusterAddonTarget, log logr.Logger) (*cluster.Cluster, error) {
	cl, err := cluster.New(name, log)
	if err != nil {
		return nil, err
	}

	err = cl.SetServerCoordinates(
		target.URL,
		target.CACert,
		target.User,
		target.Password,
		target.ClientCert,
		target.ClientKey)

	return cl, err
}

// RepoFor gets or creates a new Repo object from a ClusterAddon.spec.source item.
func (r *ClusterAddonReconciler) repoFor(src *v1alpha1.ClusterAddonSource, log logr.Logger) (*repogit.Repo, error) {
	log.V(1).Info("Get repo", "url", src.URL, "branch", src.Branch)

	name := repogit.Hashed(src.URL, src.Branch)

	if re, ok := r.Repos[name]; ok {
		return re, nil
	}

	re, err := repogit.New(src.URL, src.Branch, src.Token, log)
	if err != nil {
		return nil, err
	}
	r.Repos[name] = re

	err = re.Update()

	return re, err
}

// IgnoreNotFound makes NotFound errors disappear.
// We generally want to ignore (not requeue) NotFound errors, since we'll get a
// reconciliation request once the object exists, and requeuing in the meantime
// won't help.
func ignoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// SetupWithManager initializes the receiver and adds it to mgr.
func (r *ClusterAddonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Repos == nil {
		r.Repos = make(map[string]*repogit.Repo)
	}

	r.recorder = mgr.GetEventRecorderFor("op-addons") //TODO use same name for metrics

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterAddon{}).
		Complete(r)
}
