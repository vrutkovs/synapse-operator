package synapse

import (
	"flag"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	g "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	synapsev1alpha1 "github.com/vrutkovs/synapse-operator/pkg/apis/synapse/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

var _ = ginkgo.Describe("[synapse]", func() {
	var (
		cl   client.Client
		t    *testing.T
		name string
		ns   string
	)
	ginkgo.BeforeEach(func() {
		t = Testing
		name = "example-synapse"
		ns = "synapse"
	})

	ginkgo.It("should create secret", func() {
		spec := synapsev1alpha1.SynapseSpec{
			Secrets: synapsev1alpha1.SynapseSecrets{
				Cert:       "foo",
				Key:        "bar",
				SigningKey: "baz",
			},
		}
		instance := initFakeSynapse(t, name, ns, &spec)
		cl = initFakeClient(t, instance, name, ns)
		secret := getSecret(t, instance, cl, ns)
		g.Expect(secret.Name).To(g.Equal(instance.GetSecretName()))
		g.Expect(secret.Labels).To(g.Equal(map[string]string{"app": name}))
		g.Expect(secret.Data).To(g.Equal(map[string][]byte{
			"cert":       []byte("foo"),
			"key":        []byte("bar"),
			"signingKey": []byte("baz"),
		}))
	})

	ginkgo.It("should create configmap", func() {
		spec := synapsev1alpha1.SynapseSpec{
			Config: synapsev1alpha1.SynapseConfig{
				Homeserver: "foo",
				Logging:    "bar",
			},
		}
		instance := initFakeSynapse(t, name, ns, &spec)
		cl = initFakeClient(t, instance, name, ns)
		cm := getConfigMap(t, instance, cl, ns)
		g.Expect(cm.Name).To(g.Equal(instance.GetConfigMapName()))
		g.Expect(cm.Labels).To(g.Equal(map[string]string{"app": name}))
		g.Expect(cm.Data).To(g.Equal(map[string]string{
			"homeserver": "foo",
			"logging":    "bar",
		}))
	})

	ginkgo.It("should create service", func() {
		spec := synapsev1alpha1.SynapseSpec{
			Ports: synapsev1alpha1.SynapsePorts{
				HTTP:        80,
				HTTPS:       443,
				Replication: 9092,
			},
		}
		instance := initFakeSynapse(t, name, ns, &spec)
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
			{
				Name:       "https",
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: "https"},
				Port:       int32(443),
			},
			{
				Name:       "replication",
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: "replication"},
				Port:       int32(9092),
			},
		}))
	})

	ginkgo.It("should create deployment", func() {
		image := "docker.io/foo/bar"
		spec := synapsev1alpha1.SynapseSpec{
			ServerName: "foo.bar",
			Image:      image,
			Ports: synapsev1alpha1.SynapsePorts{
				HTTP:        80,
				HTTPS:       443,
				Replication: 9092,
			},
		}
		instance := initFakeSynapse(t, name, ns, &spec)
		cl = initFakeClient(t, instance, name, ns)
		dep := getDeployment(t, instance, cl, ns)
		g.Expect(dep.Name).To(g.Equal(instance.GetDeploymentName()))
		g.Expect(dep.Labels).To(g.Equal(map[string]string{"app": name}))
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
								Key:  "homeserver",
								Path: "homeserver.yaml",
							},
							{
								Key:  "logging",
								Path: "foo.bar.log.config",
							},
						},
						DefaultMode: &mode,
					},
				},
			},
			{
				Name: "keys",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: instance.GetSecretName(),
						Items: []corev1.KeyToPath{
							{
								Key:  "signingKey",
								Path: "foo.bar.signing.key",
							},
							{
								Key:  "cert",
								Path: "tls.crt",
							},
							{
								Key:  "key",
								Path: "tls.key",
							},
						},
						DefaultMode: &mode,
					},
				},
			},
		}))

		container := pod.Spec.Containers[0]
		g.Expect(container.Name).To(g.Equal("synapse"))
		g.Expect(container.Image).To(g.Equal(image))
		g.Expect(container.Ports).To(g.Equal([]corev1.ContainerPort{
			{
				Name:          "http",
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: int32(80),
			},
			{
				Name:          "https",
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: int32(443),
			},
			{
				Name:          "replication",
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: int32(9092),
			},
		}))
		g.Expect(container.LivenessProbe).NotTo(g.BeNil())
		g.Expect(container.ReadinessProbe).NotTo(g.BeNil())
		g.Expect(container.VolumeMounts).To(g.Equal([]corev1.VolumeMount{
			{
				Name:      "config",
				MountPath: "/synapse/config",
			},
			{
				Name:      "keys",
				MountPath: "/synapse/keys",
			},
		}))
	})

})
