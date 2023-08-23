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

	"github.com/sco1237896/sco-operator/pkg/controller"

	"github.com/go-logr/logr"
	client "github.com/sco1237896/sco-operator/pkg/controller/client"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	wsApi "github.com/sco1237896/sco-operator/api/sco/v1alpha1"
)

func NewKWorkspaceReconciler(manager ctrl.Manager) (*WorkspaceReconciler, error) {
	c, err := client.NewClient(manager.GetConfig(), manager.GetScheme(), manager.GetClient())
	if err != nil {
		return nil, err
	}

	rec := WorkspaceReconciler{}
	rec.l = ctrl.Log.WithName("controller")
	rec.Client = c
	rec.Scheme = manager.GetScheme()
	rec.ClusterType = controller.ClusterTypeVanilla

	isOpenshift, err := c.IsOpenShift()
	if err != nil {
		return nil, err
	}
	if isOpenshift {
		rec.ClusterType = controller.ClusterTypeOpenShift
	}

	rec.actions = make([]controller.Action[wsApi.Workspace], 0)

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

func (r *WorkspaceReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
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
