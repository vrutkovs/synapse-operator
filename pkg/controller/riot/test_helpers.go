package riot

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	riotv1alpha1 "github.com/vrutkovs/synapse-operator/pkg/apis/riot/v1alpha1"

	g "github.com/onsi/gomega"
)

func initFakeRiot(t *testing.T, name, ns string, spec *riotv1alpha1.RiotSpec) *riotv1alpha1.Riot {
	return &riotv1alpha1.Riot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: *spec,
	}
}

func initFakeClient(t *testing.T, riot *riotv1alpha1.Riot, name, ns string) client.Client {
	objs := []runtime.Object{riot}
	s := scheme.Scheme
	s.AddKnownTypes(riotv1alpha1.SchemeGroupVersion, objs...)

	// Reconcile
	cl := fake.NewFakeClientWithScheme(s, objs...)
	r := &ReconcileRiot{client: cl, scheme: s}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: ns,
		},
	}
	res, err := r.Reconcile(req)
	g.Expect(err).NotTo(g.HaveOccurred(), "failed to reconcile")
	g.Expect(res).To(g.Equal(reconcile.Result{}), "reconcile did not return an empty Result")
	return cl
}

func getConfigMap(t *testing.T, riot *riotv1alpha1.Riot, cl client.Client, ns string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: riot.GetConfigMapName(), Namespace: ns}, cm)
	g.Expect(err).NotTo(g.HaveOccurred(), "failed to get configmap")
	return cm
}

func getService(t *testing.T, riot *riotv1alpha1.Riot, cl client.Client, ns string) *corev1.Service {
	svc := &corev1.Service{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: riot.GetServiceName(), Namespace: ns}, svc)
	g.Expect(err).NotTo(g.HaveOccurred(), "failed to get service")
	return svc
}

func getDeployment(t *testing.T, riot *riotv1alpha1.Riot, cl client.Client, ns string) *appsv1.Deployment {
	dep := &appsv1.Deployment{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: riot.GetDeploymentName(), Namespace: ns}, dep)
	g.Expect(err).NotTo(g.HaveOccurred(), "failed to get deployment")
	return dep
}
