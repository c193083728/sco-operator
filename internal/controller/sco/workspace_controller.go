/*
Copyright 2023.

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

package sco

import (
	"context"
	"sort"

	wsApi "github.com/sco1237896/sco-operator/api/sco/v1alpha1"
	"go.uber.org/multierr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/sco1237896/sco-operator/pkg/controller"

	"github.com/go-logr/logr"
	client "github.com/sco1237896/sco-operator/pkg/controller/client"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func NewKWorkspaceReconciler(manager ctrl.Manager) (*WorkspaceReconciler, error) {
	c, err := client.NewClient(manager.GetConfig(), manager.GetScheme(), manager.GetClient())
	if err != nil {
		return nil, err
	}

	rec := WorkspaceReconciler{
		Client:      c,
		Scheme:      manager.GetScheme(),
		ClusterType: controller.ClusterTypeVanilla,
		actions:     make([]controller.Action[wsApi.Workspace], 0),
		l:           ctrl.Log.WithName("controller"),
	}
	rec.actions = append(rec.actions, NewDeployAction(rec.l))

	isOpenshift, err := c.IsOpenShift()
	if err != nil {
		return nil, err
	}
	if isOpenshift {
		rec.ClusterType = controller.ClusterTypeOpenShift
	}

	return &rec, nil
}

type WorkspaceReconciler struct {
	*client.Client

	Scheme      *runtime.Scheme
	ClusterType controller.ClusterType
	actions     []controller.Action[wsApi.Workspace]
	l           logr.Logger
}

// +kubebuilder:rbac:groups=sco.sco1237896.github.com,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sco.sco1237896.github.com,resources=workspaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=sco.sco1237896.github.com,resources=workspaces/finalizers,verbs=update
// +kubebuilder:rbac:groups=camel.apache.org,resources=kameletbindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=camel.apache.org,resources=kamelets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=camel.apache.org,resources=integrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="route.openshift.io",resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.l.Info("Reconciling", "resource", req.NamespacedName.String())

	rr := controller.ReconciliationRequest[wsApi.Workspace]{
		Client: r.Client,
		NamespacedName: types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		ClusterType: r.ClusterType,
		Resource:    &wsApi.Workspace{},
	}

	err := r.Get(ctx, req.NamespacedName, rr.Resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// no CR found
			return ctrl.Result{}, nil
		}
	}

	reconcileCondition := metav1.Condition{
		Type:               "Reconcile",
		Status:             metav1.ConditionTrue,
		Reason:             "Reconciled",
		Message:            "Reconciled",
		ObservedGeneration: rr.Resource.Generation,
	}
	var allErrors error

	for i := range r.actions {
		if err := r.actions[i].Apply(ctx, &rr); err != nil {
			allErrors = multierr.Append(allErrors, err)
		}
	}

	if allErrors != nil {
		reconcileCondition.Status = metav1.ConditionFalse
		reconcileCondition.Reason = "Failure"
		reconcileCondition.Message = "Failure"

		rr.Resource.Status.Phase = "Error"
	} else {
		rr.Resource.Status.ObservedGeneration = rr.Resource.Generation
		rr.Resource.Status.Phase = "Ready"
	}

	meta.SetStatusCondition(&rr.Resource.Status.Conditions, reconcileCondition)

	sort.SliceStable(rr.Resource.Status.Conditions, func(i, j int) bool {
		return rr.Resource.Status.Conditions[i].Type < rr.Resource.Status.Conditions[j].Type
	})

	// Update status
	err = r.Status().Update(ctx, rr.Resource)
	if err != nil && k8serrors.IsConflict(err) {
		r.l.Info(err.Error())
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		allErrors = multierr.Append(allErrors, err)
	}

	return ctrl.Result{}, allErrors
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkspaceReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr)

	c = c.For(&wsApi.Workspace{}, builder.WithPredicates(
		predicate.Or(
			predicate.GenerationChangedPredicate{},
		)))

	for i := range r.actions {
		b, err := r.actions[i].Configure(ctx, r.Client, c)
		if err != nil {
			return err
		}

		c = b
	}

	return c.Complete(r)
}
