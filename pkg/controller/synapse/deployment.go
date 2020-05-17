package synapse

import (
	"context"
	"reflect"

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

func (r *ReconcileSynapse) reconcileDeployment(request reconcile.Request, instance *synapsev1alpha1.Synapse, reqLogger logr.Logger, secretName, configMapName, deploymentName string) (reconcile.Result, error) {
	// Check if this Deployment already exists
	deployment := newDeploymentForCR(instance, secretName, configMapName, deploymentName)

	// Set Synapse instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Deployment created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		// Unknown error - requeue
		reqLogger.Info("Deployment reconcile error", "DeploymentDeployment.Namespace", found.Namespace, "Deployment.Name", found.Name, "Error", err)
		return reconcile.Result{Requeue: true}, nil
	} else if err == nil {
		expectedSpec := getExpectedDeploymentSpec(instance, secretName, configMapName, deploymentName)
		// Check if deployment needs to be updated
		if deploymentNeedsUpdate(&found.Spec, &expectedSpec, reqLogger) {
			found.ObjectMeta = deployment.ObjectMeta
			controllerutil.SetControllerReference(instance, found, r.scheme)
			found.Spec = expectedSpec
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{Requeue: true}, err
			}
			reqLogger.Info("Deployment spec updated", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return reconcile.Result{}, nil
		}
	}

	// Deployment already exists - don't requeue
	reqLogger.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	return reconcile.Result{}, nil
}

func deploymentNeedsUpdate(actual, expected *appsv1.DeploymentSpec, reqLogger logr.Logger) bool {
	// Replicas
	if actual.Replicas != nil && expected.Replicas != nil && *actual.Replicas != *expected.Replicas {
		reqLogger.Info("Deployment replicas mismatch found", "actual", actual.Replicas, "expected", expected.Replicas)
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

func getVolumes(cr *synapsev1alpha1.Synapse, configMapName, secretName string) []corev1.Volume {
	mode := int32(420)
	return []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "homeserver",
							Path: "homeserver.yaml",
						},
						{
							Key:  "logging",
							Path: cr.Spec.ServerName + ".log.config",
						},
					},
					DefaultMode: &mode,
				},
			},
		},
		{
			Name: "keys",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
					Items: []corev1.KeyToPath{
						{
							Key:  "signingKey",
							Path: cr.Spec.ServerName + ".signing.key",
						},
						{
							Key:  "cert",
							Path: "tls.crt",
						},
						{
							Key:  "key",
							Path: "tls.key",
						},
					},
					DefaultMode: &mode,
				},
			},
		},
		{
			Name: "mediastore",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

func getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "config",
			MountPath: "/synapse/config",
		},
		{
			Name:      "keys",
			MountPath: "/synapse/keys",
		},
		{
			Name:      "mediastore",
			MountPath: "/media_store",
		},
	}
}

func getReadinessProbe() corev1.Probe {
	return corev1.Probe{
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

func getLivenessProbe() corev1.Probe {
	return corev1.Probe{
		InitialDelaySeconds: 120,
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

func getContainerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: 8008,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "https",
			ContainerPort: 8448,
			Protocol:      corev1.ProtocolTCP,
		},
	}
}

func getDeploymentLabels(cr *synapsev1alpha1.Synapse) map[string]string {
	return map[string]string{
		"app": cr.Name,
	}
}

func getExpectedDeploymentSpec(cr *synapsev1alpha1.Synapse, secretName, configMapName, deploymentName string) appsv1.DeploymentSpec {

	replicas := int32(1)
	readinessProbe := getReadinessProbe()
	livenessProbe := getLivenessProbe()

	return appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: getDeploymentLabels(cr),
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName + "-pod",
				Namespace: cr.Namespace,
				Labels:    getDeploymentLabels(cr),
			},
			Spec: corev1.PodSpec{
				Volumes: getVolumes(cr, configMapName, secretName),
				Containers: []corev1.Container{
					{
						Name:           "synapse",
						Image:          cr.Spec.Image,
						ReadinessProbe: &readinessProbe,
						LivenessProbe:  &livenessProbe,
						Ports:          getContainerPorts(),
						VolumeMounts:   getVolumeMounts(),
					},
				},
			},
		},
	}
}

// newDeploymentForCR returns a busybox pod with the same name/namespace as the cr
func newDeploymentForCR(cr *synapsev1alpha1.Synapse, secretName, configMapName, deploymentName string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: cr.Namespace,
			Labels:    getDeploymentLabels(cr),
		},
		Spec: getExpectedDeploymentSpec(cr, secretName, configMapName, deploymentName),
	}
}
