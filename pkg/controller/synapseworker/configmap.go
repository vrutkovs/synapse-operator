package synapseworker

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	synapsev1alphav1 "github.com/vrutkovs/synapse-operator/pkg/apis/synapse/v1alpha1"
	synapseworkerv1alphav1 "github.com/vrutkovs/synapse-operator/pkg/apis/synapse/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileSynapseWorker) reconcileConfigMap(request reconcile.Request, instance *synapseworkerv1alphav1.SynapseWorker, reqLogger logr.Logger, s *synapsev1alphav1.Synapse) (reconcile.Result, bool, error) {

	// Find referenced synapse
	configMap, err := r.newConfigMapForCR(instance, s)
	if err != nil {
		return reconcile.Result{}, false, err
	}

	// Set SynapseWorker instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, configMap, r.scheme); err != nil {
		return reconcile.Result{}, false, err
	}

	// Check if this ConfigMap already exists
	found := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, found)
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
		expectedData, err := r.getExpectedConfigmapData(instance, s)
		if err != nil {
			reqLogger.Info("Error generating worker configmap", "ConfigMap.Namespace", found.Namespace, "ConfigMap.Name", found.Name, err)
			return reconcile.Result{}, false, err
		}
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

// newConfigMapForCR returns a busybox pod with the same name/namespace as the cr
func (r *ReconcileSynapseWorker) newConfigMapForCR(cr *synapseworkerv1alphav1.SynapseWorker, s *synapsev1alphav1.Synapse) (*corev1.ConfigMap, error) {
	labels := map[string]string{
		"app": cr.Name,
	}
	data, err := r.getExpectedConfigmapData(cr, s)
	if err != nil {
		return nil, err
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetConfigMapName(),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: data,
	}, err
}

// getExpectedConfigmapData returns expected data stored in configmap
func (r *ReconcileSynapseWorker) getExpectedConfigmapData(cr *synapseworkerv1alphav1.SynapseWorker, s *synapsev1alphav1.Synapse) (map[string]string, error) {
	config, err := cr.GenerateConfig(s)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"worker.yaml": string(config),
	}, nil
}
