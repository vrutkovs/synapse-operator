package riot

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	riotv1alphav1 "github.com/vrutkovs/synapse-operator/pkg/apis/riot/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileRiot) reconcileService(request reconcile.Request, instance *riotv1alphav1.Riot, reqLogger logr.Logger) (reconcile.Result, error) {
	service := newServiceForCR(instance)

	// Set Riot instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Service created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		// Unknown error - requeue
		reqLogger.Info("Service reconcile error", "Service.Namespace", found.Namespace, "Service.Name", found.Name, "Error", err)
		return reconcile.Result{Requeue: true}, nil
	} else if err == nil {
		// Check if Service fields haven't change

		expectedSpec := getExpectedServiceSpec(instance)
		if serviceNeedsUpdate(&found.Spec, &expectedSpec, reqLogger) {
			found.ObjectMeta = service.ObjectMeta
			controllerutil.SetControllerReference(instance, found, r.scheme)
			found.Spec = expectedSpec
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{Requeue: true}, err
			}
			reqLogger.Info("Service spec updated", "Service.Namespace", found.Namespace, "Service.Name", found.Name)
			return reconcile.Result{}, nil
		}
	}

	// Service already exists - don't requeue
	reqLogger.Info("Skip reconcile: Service already exists", "Service.Namespace", found.Namespace, "Service.Name", found.Name)
	return reconcile.Result{}, nil
}

// getExpectedServiceData returns expected data stored in Service
func getExpectedServiceSpec(cr *riotv1alphav1.Riot) corev1.ServiceSpec {
	return corev1.ServiceSpec{
		Selector: getDeploymentLabels(cr),
		Type:     corev1.ServiceTypeClusterIP,
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: "http"},
				Port:       80,
			},
		},
	}
}

// newServiceForCR returns a busybox pod with the same name/namespace as the cr
func newServiceForCR(cr *riotv1alphav1.Riot) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetServiceName(),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: getExpectedServiceSpec(cr),
	}
}

func serviceNeedsUpdate(actual, expected *corev1.ServiceSpec, reqLogger logr.Logger) bool {
	// Selector
	if !reflect.DeepEqual(actual.Selector, expected.Selector) {
		reqLogger.Info("Service selector mismatch found", "actual", actual.Selector, "expected", expected.Selector)
		return true
	}

	// Ports
	if !reflect.DeepEqual(actual.Ports, expected.Ports) {
		reqLogger.Info("Service ports mismatch found", "actual", actual.Ports, "expected", expected.Ports)
		return true
	}

	return false
}
