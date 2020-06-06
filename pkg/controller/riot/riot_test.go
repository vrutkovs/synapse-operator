package riot

import (
	"flag"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	g "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	riotv1alpha1 "github.com/vrutkovs/synapse-operator/pkg/apis/riot/v1alpha1"
)

var (
	// Handle operator-sdk flags so that unit tests could be run locally
	namespacedMan      = flag.String("namespacedMan", "", "")
	globalMan          = flag.String("globalMan", "", "")
	root               = flag.String("root", "", "")
	skipCleanupOnError = flag.Bool("skipCleanupOnError", false, "")

	// Other vars
	Testing *testing.T

	// Timeouts
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestGinkgo(t *testing.T) {
	g.RegisterFailHandler(ginkgo.Fail)
	Testing = t
	ginkgo.RunSpecs(t, "unit tests")
}

var _ = ginkgo.Describe("[riot]", func() {
	var (
		cl   client.Client
		t    *testing.T
		name string
		ns   string
	)
	ginkgo.BeforeEach(func() {
		t = Testing
		name = "example-riot"
		ns = "synapse"
	})

	ginkgo.It("should create configmap", func() {
		spec := riotv1alpha1.RiotSpec{
			Config: "{foo: bar}",
		}
		instance := initFakeRiot(t, name, ns, &spec)
		cl = initFakeClient(t, instance, name, ns)
		cm := getConfigMap(t, instance, cl, ns)
		g.Expect(cm.Name).To(g.Equal(instance.GetConfigMapName()))
		g.Expect(cm.Labels).To(g.Equal(map[string]string{"app": name}))
		g.Expect(cm.Data).To(g.Equal(map[string]string{
			"config.json": "{foo: bar}",
		}))
	})

	ginkgo.It("should create service", func() {
		spec := riotv1alpha1.RiotSpec{}
		instance := initFakeRiot(t, name, ns, &spec)
		cl = initFakeClient(t, instance, name, ns)
		svc := getService(t, instance, cl, ns)
		g.Expect(svc.Name).To(g.Equal(instance.GetServiceName()))
		g.Expect(svc.Labels).To(g.Equal(map[string]string{"app": name}))
		g.Expect(svc.Spec.Selector).To(g.Equal(map[string]string{"app": name}))
		g.Expect(svc.Spec.Type).To(g.Equal(corev1.ServiceTypeClusterIP))
		g.Expect(svc.Spec.Ports).To(g.Equal([]corev1.ServicePort{
			{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: "http"},
				Port:       int32(80),
			},
		}))
	})

	ginkgo.It("should create deployment", func() {
		image := "docker.io/foo/bar"
		replicas := int32(1)
		spec := riotv1alpha1.RiotSpec{
			Replicas: int(replicas),
			Image:    image,
		}
		instance := initFakeRiot(t, name, ns, &spec)
		cl = initFakeClient(t, instance, name, ns)
		cm := getConfigMap(t, instance, cl, ns)
		g.Expect(cm.Name).To(g.Equal(instance.GetConfigMapName()))
		g.Expect(cm.Labels).To(g.Equal(map[string]string{"app": name}))
		dep := getDeployment(t, instance, cl, ns)
		g.Expect(dep.Name).To(g.Equal(instance.GetDeploymentName()))
		g.Expect(dep.Labels).To(g.Equal(map[string]string{"app": name}))
		g.Expect(dep.Spec.Replicas).To(g.Equal(&replicas))
		g.Expect(dep.Spec.Selector.MatchLabels).To(g.Equal(map[string]string{"app": name}))

		pod := dep.Spec.Template
		g.Expect(pod.Name).To(g.Equal(instance.GetDeploymentPodName()))
		g.Expect(len(pod.Spec.Containers)).To(g.Equal(1))

		mode := int32(420)
		g.Expect(pod.Spec.Volumes).To(g.Equal([]corev1.Volume{
			{
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: instance.GetConfigMapName(),
						},
						Items: []corev1.KeyToPath{
							{
								Key:  "config.json",
								Path: "config.json",
							},
						},
						DefaultMode: &mode,
					},
				},
			},
		}))

		container := pod.Spec.Containers[0]
		g.Expect(container.Name).To(g.Equal("riot"))
		g.Expect(container.Image).To(g.Equal(image))
		g.Expect(container.Ports).To(g.Equal([]corev1.ContainerPort{
			{
				Name:          "http",
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: int32(80),
			},
		}))
		g.Expect(container.LivenessProbe).NotTo(g.BeNil())
		g.Expect(container.ReadinessProbe).NotTo(g.BeNil())
		g.Expect(container.VolumeMounts).To(g.Equal([]corev1.VolumeMount{
			{
				Name:      "config",
				MountPath: "/etc/riot-web/",
				SubPath:   "config.json",
			},
		}))
	})
})
