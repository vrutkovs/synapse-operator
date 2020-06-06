package v1alpha1

// GetConfigMapName returns managed configmap name
func (s *Riot) GetConfigMapName() string {
	return s.ObjectMeta.Name + "-config"
}

// GetDeploymentName returns managed deployment name
func (s *Riot) GetDeploymentName() string {
	return s.ObjectMeta.Name
}

// GetDeploymentPodName returns generated pod name in the deployment
func (s *Riot) GetDeploymentPodName() string {
	return s.ObjectMeta.Name + "-pod"
}

// GetServiceName returns generated pod name in the deployment
func (s *Riot) GetServiceName() string {
	return s.ObjectMeta.Name + "-service"
}

// GetExpectedConfigmapData returns expected data stored in configmap
func (s *Riot) GetExpectedConfigmapData() map[string]string {
	return map[string]string{
		"config.json": s.Spec.Config,
	}
}
