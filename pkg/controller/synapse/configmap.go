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

func (r *ReconcileSynapse) reconcileConfigMap(request reconcile.Request, instance *synapsev1alpha1.Synapse, reqLogger logr.Logger) (reconcile.Result, bool, error) {
	configMap := newConfigMapForCR(instance)

	// Set Synapse instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, configMap, r.scheme); err != nil {
		return reconcile.Result{}, false, err
	}

	// Check if this ConfigMap already exists
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
		err = r.client.Create(context.TODO(), configMap)
		if err != nil {
			return reconcile.Result{}, false, err
		}

		// ConfigMap created successfully - don't requeue
		return reconcile.Result{}, true, nil
	} else if err != nil {
		// Unknown error - requeue
		reqLogger.Info("ConfigMap reconcile error", "ConfigMap.Namespace", found.Namespace, "ConfigMap.Name", found.Name, "Error", err)
		return reconcile.Result{Requeue: true}, false, nil
	} else if err == nil {
		// Check if configmap fields haven't change
		expectedData := getExpectedConfigmapData(instance)
		if !reflect.DeepEqual(found.Data, expectedData) {
			found.ObjectMeta = configMap.ObjectMeta
			controllerutil.SetControllerReference(instance, found, r.scheme)
			found.Data = expectedData
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{Requeue: true}, false, err
			}
			reqLogger.Info("ConfigMap contents updated", "ConfigMap.Namespace", found.Namespace, "ConfigMap.Name", found.Name)
			return reconcile.Result{}, true, nil
		}
	}

	// ConfigMap already exists - don't requeue
	reqLogger.Info("Skip reconcile: ConfigMap already exists", "ConfigMap.Namespace", found.Namespace, "ConfigMap.Name", found.Name)
	return reconcile.Result{}, false, nil
}

// getExpectedConfigmapData returns expected data stored in configmap
func getExpectedConfigmapData(cr *synapsev1alpha1.Synapse) map[string]string {
	return map[string]string{
		"homeserver": cr.Spec.Config.Homeserver,
		"logging":    cr.Spec.Config.Logging,
	}
}

// newConfigMapForCR returns a busybox pod with the same name/namespace as the cr
func newConfigMapForCR(cr *synapsev1alpha1.Synapse) *corev1.ConfigMap {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetConfigMapName(),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: getExpectedConfigmapData(cr),
	}
}
