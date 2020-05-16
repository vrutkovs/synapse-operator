package synapse

import (
	"context"

	"github.com/go-logr/logr"
	synapsev1alpha1 "github.com/vrutkovs/synapse-operator/pkg/apis/synapse/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
		return reconcile.Result{}, err
	}

	// Deployment already exists - don't requeue
	reqLogger.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	return reconcile.Result{}, nil
}

// newDeploymentForCR returns a busybox pod with the same name/namespace as the cr
func newDeploymentForCR(cr *synapsev1alpha1.Synapse, secretName, configMapName, deploymentName string) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}
	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName + "-pod",
					Namespace: cr.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
								},
							},
						},
						{
							Name: "secrets",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: secretName,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "synapse",
							Image: cr.Spec.Image,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/synapse/config/homeserver.yaml",
									SubPath:   "homeserver",
								},
								{
									Name:      "config",
									MountPath: "/synapse/config/log.yaml",
									SubPath:   "logging",
								},
								{
									Name:      "secrets",
									MountPath: "/synapse/config/" + cr.Spec.ServerName + ".signing.key",
									SubPath:   "configSigningKey",
								},
								{
									Name:      "secrets",
									MountPath: "/synapse/config/" + cr.Spec.ServerName + ".tls.crt",
									SubPath:   "tlsCrt",
								},
								{
									Name:      "secrets",
									MountPath: "/synapse/config/" + cr.Spec.ServerName + ".tls.key",
									SubPath:   "key",
								},
								{
									Name:      "secrets",
									MountPath: "/synapse/config/" + cr.Spec.ServerName + ".tls.dh",
									SubPath:   "dhParams",
								},
								{
									Name:      "secrets",
									MountPath: "/synapse/tls/tls.crt",
									SubPath:   "cert",
								},
								{
									Name:      "secrets",
									MountPath: "/synapse/tls/tls.key",
									SubPath:   "key",
								},
								{
									Name:      "secrets",
									MountPath: "/synapse/keys/dhparams.key",
									SubPath:   "dhParams",
								},
								{
									Name:      "secrets",
									MountPath: "/synapse/keys/signing.key",
									SubPath:   "tlsSigningKey",
								},
							},
						},
					},
				},
			},
		},
	}
}
