package sco

import (
	"context"

	"github.com/go-logr/logr"

	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/sco1237896/sco-operator/api/sco/v1alpha1"
	"github.com/sco1237896/sco-operator/pkg/controller"
	"github.com/sco1237896/sco-operator/pkg/controller/client"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewDeployAction(l logr.Logger) controller.Action[v1alpha1.Workspace] {
	return &deployAction{
		logger: l,
	}
}

type deployAction struct {
	logger logr.Logger
}

func (a *deployAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	b = b.Owns(&appsv1.Deployment{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
		)))

	return b, nil
}

func (a *deployAction) Cleanup(context.Context, *controller.ReconciliationRequest[v1alpha1.Workspace]) error {
	return nil
}

func (a *deployAction) Apply(ctx context.Context, rr *controller.ReconciliationRequest[v1alpha1.Workspace]) error {
	deploymentCondition := metav1.Condition{
		Type:               "Deployment",
		Status:             metav1.ConditionTrue,
		Reason:             "Deployed",
		Message:            "Deployed",
		ObservedGeneration: rr.Resource.Generation,
	}

	err := a.deploy(ctx, rr)
	if err != nil {
		deploymentCondition.Status = metav1.ConditionFalse
		deploymentCondition.Reason = "Failure"
		deploymentCondition.Message = err.Error()
	}

	meta.SetStatusCondition(&rr.Resource.Status.Conditions, deploymentCondition)

	return err
}

func (a *deployAction) deploy(ctx context.Context, rr *controller.ReconciliationRequest[v1alpha1.Workspace]) error {
	ip := camelv1.IntegrationPlatform{
		TypeMeta: metav1.TypeMeta{
			APIVersion: camelv1.SchemeGroupVersion.String(),
			Kind:       camelv1.IntegrationPlatformKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rr.Resource.Namespace,
			Name:      rr.Resource.Name,
			OwnerReferences: []metav1.OwnerReference{
				{
					UID:        rr.Resource.UID,
					APIVersion: rr.Resource.APIVersion,
					Kind:       rr.Resource.Kind,
					Name:       rr.Resource.Name,
				},
			},
			Labels: map[string]string{
				controller.KubernetesLabelAppName:      rr.Resource.Name,
				controller.KubernetesLabelAppPartOf:    "sco-operator",
				controller.KubernetesLabelAppManagedBy: "sco-operator",
				controller.KubernetesLabelCreatedBy:    "sco-operator",
			},
		},
	}

	or, err := controllerutil.CreateOrUpdate(ctx, rr.Client, &ip, func() error {
		return nil
	})
	if err != nil {
		return err
	}

	a.logger.Info("IntegrationPlatform", "Operation result", or, "ID", ip.UID)
	return nil
}
