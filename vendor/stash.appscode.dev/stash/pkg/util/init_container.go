/*
Copyright The Stash Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"

	"stash.appscode.dev/stash/apis"
	v1alpha1_api "stash.appscode.dev/stash/apis/stash/v1alpha1"
	v1beta1_api "stash.appscode.dev/stash/apis/stash/v1beta1"
	"stash.appscode.dev/stash/pkg/docker"

	"github.com/appscode/go/types"
	core "k8s.io/api/core/v1"
	"kmodules.xyz/client-go/tools/cli"
	"kmodules.xyz/client-go/tools/clientcmd"
	ofst_util "kmodules.xyz/offshoot-api/util"
)

func NewInitContainer(r *v1alpha1_api.Restic, workload v1alpha1_api.LocalTypedReference, image docker.Docker) core.Container {
	container := NewSidecarContainer(r, workload, image)
	container.Args = []string{
		"backup",
		"--restic-name=" + r.Name,
		"--workload-kind=" + workload.Kind,
		"--workload-name=" + workload.Name,
		"--docker-registry=" + image.Registry,
		"--image-tag=" + image.Tag,
		"--pushgateway-url=" + PushgatewayURL(),
		fmt.Sprintf("--use-kubeapiserver-fqdn-for-aks=%v", clientcmd.UseKubeAPIServerFQDNForAKS()),
		fmt.Sprintf("--enable-analytics=%v", cli.EnableAnalytics),
	}
	container.Args = append(container.Args, cli.LoggerOptions.ToFlags()...)

	return container
}

func NewRestoreInitContainer(rs *v1beta1_api.RestoreSession, repository *v1alpha1_api.Repository, image docker.Docker) core.Container {
	initContainer := core.Container{
		Name:  apis.StashInitContainer,
		Image: image.ToContainerImage(),
		Args: append([]string{
			"restore",
			"--restoresession=" + rs.Name,
			"--secret-dir=" + apis.StashSecretMountDir,
			fmt.Sprintf("--enable-cache=%v", !rs.Spec.TempDir.DisableCaching),
			fmt.Sprintf("--max-connections=%v", repository.Spec.Backend.MaxConnections()),
			"--metrics-enabled=true",
			"--pushgateway-url=" + PushgatewayURL(),
			fmt.Sprintf("--use-kubeapiserver-fqdn-for-aks=%v", clientcmd.UseKubeAPIServerFQDNForAKS()),
			fmt.Sprintf("--enable-analytics=%v", cli.EnableAnalytics),
		}, cli.LoggerOptions.ToFlags()...),
		Env: []core.EnvVar{
			{
				Name: apis.KeyNodeName,
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "spec.nodeName",
					},
				},
			},
			{
				Name: apis.KeyPodName,
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
		},
		VolumeMounts: []core.VolumeMount{
			{
				Name:      apis.StashSecretVolume,
				MountPath: apis.StashSecretMountDir,
			},
		},
	}

	// mount tmp volume
	initContainer.VolumeMounts = UpsertTmpVolumeMount(initContainer.VolumeMounts)

	// mount the volumes specified in RestoreSession inside this init-container
	for _, srcVol := range rs.Spec.Target.VolumeMounts {
		initContainer.VolumeMounts = append(initContainer.VolumeMounts, core.VolumeMount{
			Name:      srcVol.Name,
			MountPath: srcVol.MountPath,
			SubPath:   srcVol.SubPath,
		})
	}

	// if Repository uses local volume as backend, we have to mount it inside the initContainer
	if repository.Spec.Backend.Local != nil {
		_, mnt := repository.Spec.Backend.Local.ToVolumeAndMount(apis.LocalVolumeName)
		initContainer.VolumeMounts = append(initContainer.VolumeMounts, mnt)
	}

	// pass container runtime settings from RestoreSession to init-container
	if rs.Spec.RuntimeSettings.Container != nil {
		initContainer = ofst_util.ApplyContainerRuntimeSettings(initContainer, *rs.Spec.RuntimeSettings.Container)
	}

	// In order to preserve file ownership, restore process need to be run as root user.
	// Stash image uses non-root user "stash"(1005). We have to use securityContext to run stash as root user.
	// If a user specify securityContext either in pod level or container level in RuntimeSetting,
	// don't overwrite that. In this case, user must take the responsibility of possible file ownership modification.
	securityContext := &core.SecurityContext{
		RunAsUser:  types.Int64P(0),
		RunAsGroup: types.Int64P(0),
	}
	if rs.Spec.RuntimeSettings.Container != nil {
		initContainer.SecurityContext = UpsertSecurityContext(securityContext, rs.Spec.RuntimeSettings.Container.SecurityContext)
	} else {
		initContainer.SecurityContext = securityContext
	}

	return initContainer
}
