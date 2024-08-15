package test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type PlatformChartTemplateSuite struct {
	suite.Suite
	chartPath string
}

func TestPlatformChartTemplateSuite(t *testing.T) {
	// if !testing.Short() {
	// 	t.Skip("skipping platform template test not in short mode.")
	// }
	suite.Run(t, new(PlatformChartTemplateSuite))
}

func (suite *PlatformChartTemplateSuite) SetupTest() {
	path, err := filepath.Abs("../charts/platform")
	suite.Require().NoError(err)
	suite.chartPath = path
}

func (suite *PlatformChartTemplateSuite) TestBasicDeploymentTemplateRender() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":                "latest",
			"sdk_config.client_secret": "test",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})

	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	suite.Require().Equal(deployment.Name, releaseName+"-platform")
	suite.Require().Len(deployment.Spec.Template.Spec.Containers, 1)
	suite.Require().Equal(deployment.Spec.Template.Spec.Containers[0].Image, "registry.opentdf.io/platform:latest")
}

func (suite *PlatformChartTemplateSuite) Test_SDK_Config_Set_Client_Secret_AND_Existing_Secret_Expect_Error() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":                      "latest",
			"sdk_config.client_secret":       "test",
			"sdk_config.existingSecret.name": "test",
			"sdk_config.existingSecret.key":  "test",
		},
	}

	_, err := helm.RenderTemplateE(suite.T(), options, suite.chartPath, releaseName, []string{})
	suite.Require().Error(err)
	suite.Require().ErrorContains(err, "You cannot set both client_secret and existingSecret in sdk_config.")
}

func (suite *PlatformChartTemplateSuite) Test_Set_Mode_KAS_No_SDK_Config_Defined_Expect_Error() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"mode": "kas",
		},
	}

	_, err := helm.RenderTemplateE(suite.T(), options, suite.chartPath, releaseName, []string{})
	suite.Require().Error(err)
	suite.Require().ErrorContains(err, "Mode does not contain 'core' or 'all'. You must configure the sdk_config")
}

func (suite *PlatformChartTemplateSuite) Test_Playground_Enabled_AND_Keycloak_Ing_Enabled_Trusted_Cert_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":  "latest",
			"playground": "true",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	found := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Projected != nil {
			for _, source := range volume.Projected.Sources {
				suite.T().Log("Secret Name: ", source.Secret.Name)
				if source.Secret != nil && source.Secret.Name == "keycloak.local-tls" {
					suite.Require().Equal("ca.crt", source.Secret.Items[0].Key)
					suite.Require().Equal("kc-ca.crt", source.Secret.Items[0].Path)
					found = true
				}
			}
		}
	}
	suite.Require().True(found)
}

func (suite *PlatformChartTemplateSuite) Test_Playground_Enabled_AND_Keycloak_Ing_Disabled_Trusted_Cert_Not_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":                "latest",
			"playground":               "true",
			"keycloak.ingress.enabled": "false",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	found := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Projected != nil {
			for _, source := range volume.Projected.Sources {
				if source.Secret != nil && source.Secret.Name == "keycloak.local-tls" {
					found = true
				}
			}
		}
	}
	suite.Require().False(found)
}

func (suite *PlatformChartTemplateSuite) Test_Playground_Enabled_AND_Keycloak_Ing_Enabled_AND_TLS_Disabled_Trusted_Cert_Not_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":                "latest",
			"playground":               "true",
			"keycloak.ingress.enabled": "true",
			"keycloak.ingress.tls":     "false",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	found := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Projected != nil {
			for _, source := range volume.Projected.Sources {
				if source.Secret != nil && source.Secret.Name == "keycloak.local-tls" {
					found = true
				}
			}
		}
	}
	suite.Require().False(found)
}

func (suite *PlatformChartTemplateSuite) Test_Mode_Core_No_Kas_Volumes_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
			"mode":      "core",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			if volume.Secret.SecretName == "kas-private-keys" {
				volumeFound = true
			}
		}
	}
	suite.Require().False(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "kas-private-keys" {
				volumeMountFound = true
			}
		}
	}
	suite.Require().False(volumeMountFound)
}

func (suite *PlatformChartTemplateSuite) Test_Mode_Core_And_Kas_Volumes_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
			"mode":      "core\\,kas",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			if volume.Secret.SecretName == "kas-private-keys" {
				volumeFound = true
			}
		}
	}
	suite.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "kas-private-keys" {
				volumeMountFound = true
			}
		}
	}
	suite.Require().True(volumeMountFound)
}

func (suite *PlatformChartTemplateSuite) Test_Mode_All_Kas_Volumes_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
			"mode":      "all",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			if volume.Secret.SecretName == "kas-private-keys" {
				volumeFound = true
			}
		}
	}
	suite.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "kas-private-keys" {
				volumeMountFound = true
			}
		}
	}
	suite.Require().True(volumeMountFound)
}

func (suite *PlatformChartTemplateSuite) Test_Mode_Kas_Expect_Volumes_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":                "latest",
			"mode":                     "kas",
			"sdk_config.endpoint":      "http://localhost:8080",
			"sdk_config.client_id":     "test",
			"sdk_config.client_secret": "test",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			if volume.Secret.SecretName == "kas-private-keys" {
				volumeFound = true
			}
		}
	}
	suite.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "kas-private-keys" {
				volumeMountFound = true
			}
		}
	}
	suite.Require().True(volumeMountFound)
}

func (suite *PlatformChartTemplateSuite) Test_Trusted_Cert_Volume_Exists_When_Playground_True() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":  "latest",
			"playground": "true",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "trusted-certs" {
			volumeFound = true
		}
	}
	suite.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "trusted-certs" {
				volumeMountFound = true
			}
		}
	}
	suite.Require().True(volumeMountFound)
}

func (suite *PlatformChartTemplateSuite) Test_Trusted_Cert_Volume_Exists_When_AdditionalTrustedCertsDefined() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
			"server.tls.additionalTrustedCerts[0].secret.name": "test",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "trusted-certs" {
			volumeFound = true
		}
	}
	suite.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "trusted-certs" {
				volumeMountFound = true
			}
		}
	}
	suite.Require().True(volumeMountFound)
}

func (suite *PlatformChartTemplateSuite) Test_Trusted_Cert_Volume_Does_Not_Exist_When_AdditionalTrustedCertsDefined_OR_Playground_Not_Defined() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "trusted-certs" {
			volumeFound = true
		}
	}
	suite.Require().False(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "trusted-certs" {
				volumeMountFound = true
			}
		}
	}
	suite.Require().False(volumeMountFound)
}

func (suite *PlatformChartTemplateSuite) Test_Custom_Volume_Templates_Merged() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"volumeTemplate":      "platform.volumes.test.tpl",
			"volumeMountTemplate": "platform.volumeMounts.test.tpl",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "testVolume" {
			volumeFound = true
		}
	}
	suite.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "testVolumeMount" {
				volumeMountFound = true
			}
		}
	}
	suite.Require().True(volumeMountFound)
}

func (suite *PlatformChartTemplateSuite) Test_Custom_Config_Template_Services_Merged() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"configTemplate": "platform.configuration.test.tpl",
		},
	}

	output := helm.RenderTemplate(suite.T(), options, suite.chartPath, releaseName, []string{"templates/config.yaml"})
	var cm corev1.ConfigMap
	helm.UnmarshalK8SYaml(suite.T(), output, &cm)

	var config map[string]interface{}
	suite.Suite.Require().NoError(yaml.Unmarshal([]byte(cm.Data["opentdf.yaml"]), &config))

	suite.Require().Equal(releaseName+"-platform", cm.Name)

	testServiceKeyFound := false
	for key, _ := range config {
		if key == "services" {
			for sKey, _ := range config[key].(map[string]interface{}) {
				if sKey == "testService" {
					testServiceKeyFound = true
				}
			}
		}
	}
	suite.Require().True(testServiceKeyFound)
}
