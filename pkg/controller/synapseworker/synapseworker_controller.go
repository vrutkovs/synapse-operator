package synapseworker

import (
	"context"

	synapsev1alpha1 "github.com/vrutkovs/synapse-operator/pkg/apis/synapse/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_synapseworker")

// Add creates a new SynapseWorker Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSynapseWorker{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("synapseworker-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SynapseWorker
	err = c.Watch(&source.Kind{Type: &synapsev1alpha1.SynapseWorker{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &synapsev1alpha1.SynapseWorker{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &synapsev1alpha1.SynapseWorker{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &synapsev1alpha1.SynapseWorker{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSynapseWorker implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSynapseWorker{}

// ReconcileSynapseWorker reconciles a SynapseWorker object
type ReconcileSynapseWorker struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a SynapseWorker object and makes changes based on the state read
// and what is in the SynapseWorker.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSynapseWorker) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SynapseWorker")

	// Fetch the SynapseWorker instance
	instance := &synapsev1alpha1.SynapseWorker{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Find referenced Synapse object
	s, err := instance.FindReferencedSynapse(r.client)
	if err != nil {
		reqLogger.Info("Deployment reconcile error", "Referenced Synapse object not found", err)
		return reconcile.Result{}, err
	}

	result, cmUpdated, err := r.reconcileConfigMap(request, instance, reqLogger, s)
	if err != nil {
		return result, err
	}

	result, created, err := r.reconcileDeployment(request, instance, reqLogger, s)
	if err != nil {
		return result, err
	}

	// If either of configMap or secret has been updated force rollout
	if cmUpdated && !created {
		if result, err := r.forceDeploymentRollout(request, instance, reqLogger, s); err != nil {
			return result, err
		}
	}

	result, err = r.reconcileService(request, instance, reqLogger, s)
	if err != nil {
		return result, err
	}

	return reconcile.Result{}, nil
}
