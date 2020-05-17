package v1alpha1

// GetConfigMapName returns managed configmap name
func (s *Synapse) GetConfigMapName() string {
	return s.ObjectMeta.Name + "-config"
}

// GetSecretName returns managed secret name
func (s *Synapse) GetSecretName() string {
	return s.ObjectMeta.Name + "-secret"
}

// GetDeploymentName returns managed deployment name
func (s *Synapse) GetDeploymentName() string {
	return s.ObjectMeta.Name
}

// GetDeploymentPodName returns generated pod name in the deployment
func (s *Synapse) GetDeploymentPodName() string {
	return s.ObjectMeta.Name + "-pod"
}
