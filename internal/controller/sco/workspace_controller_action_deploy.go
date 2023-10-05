package sco

import (
	"context"

	"github.com/go-logr/logr"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/sco1237896/sco-operator/api/sco/v1alpha1"
	"github.com/sco1237896/sco-operator/pkg/controller"
	"github.com/sco1237896/sco-operator/pkg/controller/client"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/builder"

	camelv1ac "github.com/apache/camel-k/v2/pkg/client/camel/applyconfiguration/camel/v1"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
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
	resource := camelv1ac.IntegrationPlatform(rr.Resource.Name, rr.Resource.Namespace).
		WithOwnerReferences(metav1ac.OwnerReference().
			WithAPIVersion(rr.Resource.GetObjectKind().GroupVersionKind().GroupVersion().String()).
			WithKind(rr.Resource.GetObjectKind().GroupVersionKind().Kind).
			WithName(rr.Resource.GetName()).
			WithUID(rr.Resource.GetUID()).
			WithBlockOwnerDeletion(true).
			WithController(true)).
		WithLabels(map[string]string{
			controller.KubernetesLabelAppName:      rr.Resource.Name,
			controller.KubernetesLabelAppPartOf:    ApplicationName,
			controller.KubernetesLabelAppManagedBy: OperatorName,
		})

	result, err := rr.Client.Camel.CamelV1().IntegrationPlatforms(rr.Resource.Namespace).Apply(
		ctx,
		resource,
		metav1.ApplyOptions{
			FieldManager: OperatorName,
			Force:        true,
		},
	)

	if err != nil {
		return err
	}

	a.logger.Info("IntegrationPlatform applied", "ID", result.UID)

	return nil
}
