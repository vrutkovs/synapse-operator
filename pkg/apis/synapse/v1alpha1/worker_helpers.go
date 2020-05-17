package v1alpha1

import (
	"context"

	"gopkg.in/yaml.v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetConfigMapName returns SynapseWorker configmap name
func (w *SynapseWorker) GetConfigMapName() string {
	return w.ObjectMeta.Name + "-config"
}

// GetDeploymentName returns SynapseWorker deployment name
func (w *SynapseWorker) GetDeploymentName() string {
	return w.ObjectMeta.Name
}

// GetDeploymentPodName returns SynapseWorker deployment name
func (w *SynapseWorker) GetDeploymentPodName() string {
	return w.ObjectMeta.Name + "-pod"
}

// GetServiceName returns SynapseWorker deployment name
func (w *SynapseWorker) GetServiceName() string {
	return w.ObjectMeta.Name + "-server"
}

// SynapseWorkerConfig represents a worker config
type SynapseWorkerConfig struct {
	App             string                  `yaml:"worker_app"`
	ReplicationHost string                  `yaml:"worker_replication_host"`
	ReplicationPort int                     `yaml:"worker_replication_port"`
	Listeners       []SynapseWorkerListener `yaml:"worker_listeners"`
}

// SynapseWorkerListener represents listener config
type SynapseWorkerListener struct {
	Protocol  string                  `yaml:"type"`
	Port      int                     `yaml:"port"`
	Resources []SynapseWorkerResource `yaml:"resources"`
}

// GenerateConfig returns string config of the worker based on SynapseWorker config
func (w *SynapseWorker) GenerateConfig(s *Synapse) ([]byte, error) {
	workerConfig := SynapseWorkerConfig{
		App:             w.Spec.Worker,
		ReplicationHost: s.GetServiceName(),
		ReplicationPort: s.Spec.Ports.Replication,
		Listeners: []SynapseWorkerListener{{
			Protocol:  w.Spec.Protocol,
			Port:      w.Spec.Port,
			Resources: w.Spec.Resources,
		}},
	}

	return yaml.Marshal(workerConfig)
}

// FindReferencedSynapse returns a pointer to Synapse instance referenced in SynapseWorker object
func (w *SynapseWorker) FindReferencedSynapse(c client.Client) (*Synapse, error) {
	synapse := &Synapse{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: w.Spec.Synapse, Namespace: w.Namespace}, synapse)
	if err != nil {
		return nil, err
	}

	return synapse, nil
}
