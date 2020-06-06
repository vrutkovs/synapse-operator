package riot

import (
	"context"
	"flag"
	"reflect"
	"testing"

	riotv1alpha1 "github.com/vrutkovs/synapse-operator/pkg/apis/riot/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Handle operator-sdk flags so that unit tests could be run locally
var (
	namespacedMan      = flag.String("namespacedMan", "", "")
	globalMan          = flag.String("globalMan", "", "")
	root               = flag.String("root", "", "")
	skipCleanupOnError = flag.Bool("skipCleanupOnError", false, "")
)

func TestRiotController(t *testing.T) {
	var (
		name           = "riotInstance"
		namespace      = "synapse-operator"
		labelKey       = "label-key"
		labelValue     = "label-value"
		image          = "locahost/riot/image:tag"
		replicas       = 1
		expectedLabels = map[string]string{"app": name}
	)

	riot := &riotv1alpha1.Riot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				labelKey: labelValue,
			},
		},
		Spec: riotv1alpha1.RiotSpec{
			Replicas:   replicas,
			Image:      image,
			ServerName: "matrix.example.com",
			Config:     "{}",
		},
	}

	objs := []runtime.Object{riot}

	s := scheme.Scheme
	s.AddKnownTypes(riotv1alpha1.SchemeGroupVersion, riot)

	// Reconcile
	cl := fake.NewFakeClientWithScheme(s, objs...)
	r := &ReconcileRiot{client: cl, scheme: s}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res != (reconcile.Result{}) {
		t.Error("reconcile did not return an empty Result")
	}

	// Check configmap
	cm := &corev1.ConfigMap{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: riot.GetConfigMapName(), Namespace: namespace}, cm)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}

	// Configmap labels
	actualCMLabels := cm.Labels
	if !reflect.DeepEqual(actualCMLabels, expectedLabels) {
		t.Fatalf("configmap labels don't match:\n%v\n%v", actualCMLabels, expectedLabels)
	}

	// Configmap data
	actualCMData := cm.Data
	expectedCMData := riot.GetExpectedConfigmapData()
	if !reflect.DeepEqual(actualCMData, expectedCMData) {
		t.Fatalf("configmap data don't match:\n%v\n%v", actualCMData, expectedCMData)
	}

	// Check deployment
	deployment := &appsv1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: riot.GetDeploymentName(), Namespace: namespace}, deployment)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}

	// Deployment labels
	actualDeploymentLabels := deployment.Labels
	if !reflect.DeepEqual(actualDeploymentLabels, expectedLabels) {
		t.Fatalf("deployment labels don't match:\n%v\n%v", actualCMLabels, actualDeploymentLabels)
	}

	// Deployment replicas
	actualDeploymentReplicas := deployment.Spec.Replicas
	if *actualDeploymentReplicas != int32(replicas) {
		t.Fatalf("deployment replicas don't match:\n%v\n%v", *actualDeploymentReplicas, replicas)
	}

	// Deployment label selector
	actualDeploymentLabelSelector := deployment.Spec.Selector.MatchLabels
	if !reflect.DeepEqual(actualDeploymentLabelSelector, expectedLabels) {
		t.Fatalf("deployment selector don't match:\n%v\n%v", actualDeploymentLabelSelector, expectedLabels)
	}

	podTemplate := deployment.Spec.Template
	// Pod name
	actualPodName := podTemplate.Name
	expectedPodName := riot.GetDeploymentPodName()
	if actualPodName != expectedPodName {
		t.Fatalf("deployment pod name don't match:\n%v\n%v", actualPodName, expectedPodName)
	}

	// Pod labels
	actualPodLabels := podTemplate.Labels
	if !reflect.DeepEqual(actualPodLabels, expectedLabels) {
		t.Fatalf("deployment pod labels don't match:\n%v\n%v", actualPodLabels, expectedLabels)
	}

	// Pod volumes
	actualPodVolumes := podTemplate.Spec.Volumes
	if len(actualPodLabels) != 1 {
		t.Fatalf("wrong number of pod volumes: %v", len(actualPodVolumes))
	}
	if actualPodVolumes[0].Name != "config" {
		t.Fatalf("wrong pod volume name: %v", actualPodVolumes[0].Name)
	}
	if actualPodVolumes[0].VolumeSource.ConfigMap == nil {
		t.Fatalf("wrong pod volume source type: %v", actualPodVolumes[0].VolumeSource)
	}
	firstVolume := actualPodVolumes[0].VolumeSource.ConfigMap
	if firstVolume.LocalObjectReference.Name != cm.Name {
		t.Fatalf("wrong pod volume reference: %v", firstVolume.LocalObjectReference)
	}
	if *firstVolume.DefaultMode != int32(420) {
		t.Fatalf("wrong default mode: %v", *firstVolume.DefaultMode)
	}
	if len(firstVolume.Items) != 1 {
		t.Fatalf("wrong number of pod volume items: %v", len(firstVolume.Items))
	}
	if firstVolume.Items[0].Key != "config.json" {
		t.Fatalf("wrong key referenced: %v", firstVolume.Items[0].Key)
	}
	if firstVolume.Items[0].Path != "config.json" {
		t.Fatalf("wrong path referenced: %v", firstVolume.Items[0].Path)
	}

	// Deployment containers
	if len(deployment.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("number of container in deployment don't match:\n%d", len(deployment.Spec.Template.Spec.Containers))
	}
	container := deployment.Spec.Template.Spec.Containers[0]
	if container.Name != "riot" {
		t.Fatalf("wrong container name: %v", container.Name)
	}
	if container.Image != image {
		t.Fatalf("wrong container image: %v", container.Image)
	}
	if container.ReadinessProbe == nil {
		t.Fatal("wrong container readiness probe")
	}
	if container.LivenessProbe == nil {
		t.Fatal("wrong container liveness probe")
	}
	if len(container.Ports) != 1 {
		t.Fatalf("wrong number of ports: %v", container.Ports)
	}
	if container.Ports[0].Name != "http" {
		t.Fatalf("wrong port name: %v", container.Ports[0].Name)
	}
	if container.Ports[0].ContainerPort != 80 {
		t.Fatalf("wrong container port: %v", container.Ports[0].ContainerPort)
	}
	if container.Ports[0].Protocol != corev1.ProtocolTCP {
		t.Fatalf("wrong port protocol: %v", container.Ports[0].Protocol)
	}
	if len(container.VolumeMounts) != 1 {
		t.Fatalf("wrong number of volume mounts: %v", len(container.VolumeMounts))
	}
	volumeMount := container.VolumeMounts[0]
	if volumeMount.Name != "config" {
		t.Fatalf("wrong volume mounts name: %v", volumeMount.Name)
	}
	if volumeMount.MountPath != "/etc/riot-web/" {
		t.Fatalf("wrong volume mounts path: %v", volumeMount.MountPath)
	}
	if volumeMount.SubPath != "config.json" {
		t.Fatalf("wrong volume mounts subpath: %v", volumeMount.SubPath)
	}

	// Check service
	svc := &corev1.Service{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: riot.GetServiceName(), Namespace: namespace}, svc)
	if err != nil {
		t.Fatalf("get service: (%v)", err)
	}

	if !reflect.DeepEqual(svc.Labels, expectedLabels) {
		t.Fatalf("service labels don't match:\n%v\n%v", svc.Labels, expectedLabels)
	}
	if !reflect.DeepEqual(svc.Spec.Selector, expectedLabels) {
		t.Fatalf("service selector don't match:\n%v\n%v", svc.Spec.Selector, expectedLabels)
	}
	if svc.Spec.Type != corev1.ServiceTypeClusterIP {
		t.Fatalf("wrong service type: %v", svc.Spec.Type)
	}
	if len(svc.Spec.Ports) != 1 {
		t.Fatalf("wrong number of service ports: %v", len(svc.Spec.Ports))
	}
	port := svc.Spec.Ports[0]
	if port.Name != "http" {
		t.Fatalf("wrong service port name: %v", port.Name)
	}
	if port.TargetPort.StrVal != "http" {
		t.Fatalf("wrong service target port name: %v", port.TargetPort)
	}
	if port.Port != int32(80) {
		t.Fatalf("wrong service port: %v", port.Port)
	}
	if port.Protocol != corev1.ProtocolTCP {
		t.Fatalf("wrong service port protocol: %v", port.Protocol)
	}
}
