package synapse

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	synapsev1alpha1 "github.com/vrutkovs/synapse-operator/pkg/apis/synapse/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileSynapse) reconcileDeployment(request reconcile.Request, instance *synapsev1alpha1.Synapse, reqLogger logr.Logger) (reconcile.Result, bool, error) {
	// Check if this Deployment already exists
	deployment := newDeploymentForCR(instance)

	// Set Synapse instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
		return reconcile.Result{}, false, err
	}

	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			return reconcile.Result{}, false, err
		}

		// Deployment created successfully - don't requeue
		return reconcile.Result{}, true, nil
	} else if err != nil {
		// Unknown error - requeue
		reqLogger.Info("Deployment reconcile error", "DeploymentDeployment.Namespace", found.Namespace, "Deployment.Name", found.Name, "Error", err)
		return reconcile.Result{Requeue: true}, false, nil
	} else if err == nil {
		expectedSpec := getExpectedDeploymentSpec(instance)
		// Check if deployment needs to be updated
		if deploymentNeedsUpdate(&found.Spec, &expectedSpec, reqLogger) {
			found.ObjectMeta = deployment.ObjectMeta
			controllerutil.SetControllerReference(instance, found, r.scheme)
			found.Spec = expectedSpec
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{Requeue: true}, false, err
			}
			reqLogger.Info("Deployment spec updated", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return reconcile.Result{}, false, nil
		}
	}

	// Deployment already exists - don't requeue
	reqLogger.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	return reconcile.Result{}, false, nil
}

func deploymentNeedsUpdate(actual, expected *appsv1.DeploymentSpec, reqLogger logr.Logger) bool {
	// Replicas
	if actual.Replicas != nil && expected.Replicas != nil && *actual.Replicas != *expected.Replicas {
		reqLogger.Info("Deployment replicas mismatch found", "actual", actual.Replicas, "expected", expected.Replicas)
		return true
	}

	// Strategy
	if !reflect.DeepEqual(actual.Strategy, expected.Strategy) {
		reqLogger.Info("Deployment strategy mismatch found", "actual", actual.Strategy, "expected", expected.Strategy)
		return true
	}

	// Template Labels
	if !reflect.DeepEqual(actual.Template.ObjectMeta.Labels, expected.Template.ObjectMeta.Labels) {
		reqLogger.Info("Deployment label mismatch found", "actual", actual.Template.ObjectMeta.Labels, "expected", expected.Template.ObjectMeta.Labels)
		return true
	}

	// Template Spec Volumes
	if !reflect.DeepEqual(actual.Template.Spec.Volumes, expected.Template.Spec.Volumes) {
		reqLogger.Info("Deployment volume mismatch found", "actual", actual.Template.Spec.Volumes, "expected", expected.Template.Spec.Volumes)
		return true
	}

	// Template Spec Containers length
	if len(actual.Template.Spec.Containers) != len(expected.Template.Spec.Containers) {
		reqLogger.Info("Deployment container number mismatch found", "actual", len(actual.Template.Spec.Containers), "expected", expected.Template.Spec.Containers)
		return true
	}

	// Template Spec Containers [0] Name
	if actual.Template.Spec.Containers[0].Name != expected.Template.Spec.Containers[0].Name {
		reqLogger.Info("Deployment name mismatch found", "actual", actual.Template.Spec.Containers[0].Name, "expected", expected.Template.Spec.Containers[0].Name)
	}

	// Template Spec Containers [0] ReadinessProbe
	if !reflect.DeepEqual(actual.Template.Spec.Containers[0].ReadinessProbe, expected.Template.Spec.Containers[0].ReadinessProbe) {
		reqLogger.Info("Deployment readiness probe mismatch found", "actual", actual.Template.Spec.Containers[0].ReadinessProbe, "expected", expected.Template.Spec.Containers[0].ReadinessProbe)
		return true
	}

	// Template Spec Containers [0] LivenessProbe
	if !reflect.DeepEqual(actual.Template.Spec.Containers[0].LivenessProbe, expected.Template.Spec.Containers[0].LivenessProbe) {
		reqLogger.Info("Deployment liveness probe mismatch found", "actual", actual.Template.Spec.Containers[0].LivenessProbe, "expected", expected.Template.Spec.Containers[0].LivenessProbe)
		return true
	}

	// Template Spec Containers [0] Image
	if actual.Template.Spec.Containers[0].Image != expected.Template.Spec.Containers[0].Image {
		reqLogger.Info("Deployment image mismatch found", "actual", actual.Template.Spec.Containers[0].Image, "expected", expected.Template.Spec.Containers[0].Image)
	}

	// Template Spec Containers [0] Ports
	if !reflect.DeepEqual(actual.Template.Spec.Containers[0].Ports, expected.Template.Spec.Containers[0].Ports) {
		reqLogger.Info("Deployment ports mismatch found", "actual", actual.Template.Spec.Containers[0].Ports, "expected", expected.Template.Spec.Containers[0].Ports)
		return true
	}

	// Template Spec Containers [0] VolumeMounts
	if !reflect.DeepEqual(actual.Template.Spec.Containers[0].VolumeMounts, expected.Template.Spec.Containers[0].VolumeMounts) {
		reqLogger.Info("Deployment volume mount mismatch found", "actual", actual.Template.Spec.Containers[0].VolumeMounts, "expected", expected.Template.Spec.Containers[0].VolumeMounts)
		return true
	}

	// Template Spec Containers [0] Args
	if !reflect.DeepEqual(actual.Template.Spec.Containers[0].Args, expected.Template.Spec.Containers[0].Args) {
		reqLogger.Info("Deployment args mismatch found", "actual", actual.Template.Spec.Containers[0].Args, "expected", expected.Template.Spec.Containers[0].Args)
		return true
	}

	// Template Spec Containers [0] Command
	if !reflect.DeepEqual(actual.Template.Spec.Containers[0].Command, expected.Template.Spec.Containers[0].Command) {
		reqLogger.Info("Deployment command mismatch found", "actual", actual.Template.Spec.Containers[0].Command, "expected", expected.Template.Spec.Containers[0].Command)
		return true
	}

	return false
}

func (r *ReconcileSynapse) forceDeploymentRollout(request reconcile.Request, instance *synapsev1alpha1.Synapse, reqLogger logr.Logger) (reconcile.Result, error) {
	deployment := newDeploymentForCR(instance)
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
	if err != nil {
		// No deployment exists, odd
		return reconcile.Result{Requeue: true}, err
	}

	// Update annotation in the pod template to force deployment rollout
	reqLogger.Info("Config/Secret changed: rolling out new pods", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	found.Spec.Template.Annotations = map[string]string{
		"synapse-operator/force-rollout": fmt.Sprintf("config changed at %q", time.Now().String()),
	}
	err = r.client.Update(context.TODO(), found)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}
	return reconcile.Result{}, nil

}

func getDefaultProbe() *corev1.Probe {
	return &corev1.Probe{
		InitialDelaySeconds: 10,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/_matrix/client/versions",
				Port:   intstr.FromString("http"),
				Scheme: "HTTP",
			},
		},
	}
}

func getReadinessProbe() corev1.Probe {
	return *getDefaultProbe()
}

func getLivenessProbe() corev1.Probe {
	probe := *getDefaultProbe()
	probe.InitialDelaySeconds = 120
	return probe
}

func getContainerPorts(cr *synapsev1alpha1.Synapse) []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: int32(cr.Spec.Ports.HTTP),
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "https",
			ContainerPort: int32(cr.Spec.Ports.HTTPS),
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "replication",
			ContainerPort: int32(cr.Spec.Ports.Replication),
			Protocol:      corev1.ProtocolTCP,
		},
	}
}

func getDeploymentLabels(cr *synapsev1alpha1.Synapse) map[string]string {
	return map[string]string{
		"app": cr.Name,
	}
}

func getExpectedDeploymentSpec(cr *synapsev1alpha1.Synapse) appsv1.DeploymentSpec {

	replicas := int32(1)
	readinessProbe := getReadinessProbe()
	livenessProbe := getLivenessProbe()

	return appsv1.DeploymentSpec{
		Replicas: &replicas,
		Strategy: appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		},
		Selector: &metav1.LabelSelector{
			MatchLabels: getDeploymentLabels(cr),
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cr.GetDeploymentPodName(),
				Namespace: cr.Namespace,
				Labels:    getDeploymentLabels(cr),
			},
			Spec: corev1.PodSpec{
				Volumes: cr.GetVolumes(),
				Containers: []corev1.Container{
					{
						Name:           "synapse",
						Image:          cr.Spec.Image,
						ReadinessProbe: &readinessProbe,
						LivenessProbe:  &livenessProbe,
						Ports:          getContainerPorts(cr),
						VolumeMounts:   cr.GetVolumeMounts(),
					},
				},
			},
		},
	}
}

// newDeploymentForCR returns a busybox pod with the same name/namespace as the cr
func newDeploymentForCR(cr *synapsev1alpha1.Synapse) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetDeploymentName(),
			Namespace: cr.Namespace,
			Labels:    getDeploymentLabels(cr),
		},
		Spec: getExpectedDeploymentSpec(cr),
	}
}
