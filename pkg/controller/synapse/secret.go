package synapse

import (
	"context"

	"github.com/go-logr/logr"
	synapsev1alpha1 "github.com/vrutkovs/synapse-operator/pkg/apis/synapse/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileSynapse) reconcileSecret(request reconcile.Request, instance *synapsev1alpha1.Synapse, reqLogger logr.Logger, secretName string) (reconcile.Result, error) {
	// Check if this Secret already exists
	Secret := newSecretForCR(instance, secretName)

	// Set Synapse instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, Secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: Secret.Name, Namespace: Secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", Secret.Namespace, "Secret.Name", Secret.Name)
		err = r.client.Create(context.TODO(), Secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Secret created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Secret already exists - don't requeue
	reqLogger.Info("Skip reconcile: Secret already exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{}, nil
}

// newSecretForCR returns a busybox pod with the same name/namespace as the cr
func newSecretForCR(cr *synapsev1alpha1.Synapse, secretName string) *corev1.Secret {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string][]byte{
			"cert":       []byte(cr.Spec.Secrets.Cert),
			"key":        []byte(cr.Spec.Secrets.Key),
			"signingKey": []byte(cr.Spec.Secrets.SigningKey),
		},
	}
}
