package synapse

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	synapsev1alpha1 "github.com/vrutkovs/synapse-operator/pkg/apis/synapse/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileSynapse) reconcileSecret(request reconcile.Request, instance *synapsev1alpha1.Synapse, reqLogger logr.Logger) (reconcile.Result, error) {
	// Check if this Secret already exists
	secret := newSecretForCR(instance)

	// Set Synapse instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Secret created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Info("Secret reconcile error", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name, "Error", err)
		return reconcile.Result{Requeue: true}, nil
	} else if err == nil {
		// Check if secret fields haven't change
		expectedData := getExpectedSecretData(instance)
		if !reflect.DeepEqual(found.Data, expectedData) {
			found.ObjectMeta = secret.ObjectMeta
			controllerutil.SetControllerReference(instance, found, r.scheme)
			found.Data = expectedData
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{Requeue: true}, err
			}
			reqLogger.Info("Secret contents updated", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
			return reconcile.Result{}, nil
		}
	}

	// Secret already exists - don't requeue
	reqLogger.Info("Skip reconcile: Secret already exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{}, nil
}

func getExpectedSecretData(cr *synapsev1alpha1.Synapse) map[string][]byte {
	return map[string][]byte{
		"cert":       []byte(cr.Spec.Secrets.Cert),
		"key":        []byte(cr.Spec.Secrets.Key),
		"signingKey": []byte(cr.Spec.Secrets.SigningKey),
	}
}

// newSecretForCR returns a busybox pod with the same name/namespace as the cr
func newSecretForCR(cr *synapsev1alpha1.Synapse) *corev1.Secret {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetSecretName(),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: getExpectedSecretData(cr),
	}
}
