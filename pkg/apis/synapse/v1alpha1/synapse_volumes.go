package v1alpha1

import corev1 "k8s.io/api/core/v1"

func (cr *Synapse) getUserVolumes() []corev1.Volume {
	volumes := []corev1.Volume{}
	for _, volume := range cr.Spec.Config.Volumes {
		volumes = append(volumes, volume.Volume)
	}
	return volumes
}

func (cr *Synapse) getSecretAndConfigVolumes() []corev1.Volume {
	mode := int32(420)
	return []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cr.GetConfigMapName(),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "homeserver",
							Path: "homeserver.yaml",
						},
						{
							Key:  "logging",
							Path: cr.Spec.ServerName + ".log.config",
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
					SecretName: cr.GetSecretName(),
					Items: []corev1.KeyToPath{
						{
							Key:  "signingKey",
							Path: cr.Spec.ServerName + ".signing.key",
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
	}
}

// GetVolumes returns a list of volumes mounted in synapse container
func (cr *Synapse) GetVolumes() []corev1.Volume {
	return append(cr.getSecretAndConfigVolumes(), cr.getUserVolumes()...)
}

func (cr *Synapse) getSecretsVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "config",
			MountPath: "/synapse/config",
		},
		{
			Name:      "keys",
			MountPath: "/synapse/keys",
		},
	}
}

func (cr *Synapse) getUserVolumeMounts() []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}
	for _, volume := range cr.Spec.Config.Volumes {
		volumeMounts = append(volumeMounts, volume.Mount)
	}
	return volumeMounts
}

// GetVolumeMounts returns a list of volume mounts in synapse container
func (cr *Synapse) GetVolumeMounts() []corev1.VolumeMount {
	return append(cr.getSecretsVolumeMounts(), cr.getUserVolumeMounts()...)
}
