package test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	yaml3 "gopkg.in/yaml.v3"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type PlatformChartTemplateSuite struct {
	suite.Suite
	chartPath string
}

func TestPlatformChartTemplateSuite(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping platform template test not in short mode.")
	}
	suite.Run(t, new(PlatformChartTemplateSuite))
}

func (s *PlatformChartTemplateSuite) SetupTest() {
	path, err := filepath.Abs("../charts/platform")
	s.Require().NoError(err)
	s.chartPath = path
}

func (s *PlatformChartTemplateSuite) TestBasicDeploymentTemplateRender() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":                "latest",
			"sdk_config.client_secret": "test",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})

	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	s.Require().Equal(deployment.Name, releaseName+"-platform")
	s.Require().Len(deployment.Spec.Template.Spec.Containers, 1)
	s.Require().Equal(deployment.Spec.Template.Spec.Containers[0].Image, "registry.opentdf.io/platform:latest")
}

func (s *PlatformChartTemplateSuite) Test_SDK_Config_Set_Client_Secret_AND_Existing_Secret_Expect_Error() {
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

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, releaseName, []string{})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "You cannot set both client_secret and existingSecret in sdk_config.")
}

func (s *PlatformChartTemplateSuite) Test_Set_Mode_KAS_No_SDK_Config_Defined_Expect_Error() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"mode": "kas",
		},
	}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, releaseName, []string{})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "Mode does not contain 'core' or 'all'. You must configure the sdk_config")
}
func (s *PlatformChartTemplateSuite) Test_SDK_Config_Client_ID_Set_Without_Secret_Expect_Error() {
	releaseName := "client-id-no-secret"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":            "latest",
			"sdk_config.client_id": "test-client-id",
			// client_secret and existingSecret are intentionally omitted
		},
	}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, releaseName, []string{})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "If sdk_config.client_id is set, you must also set either sdk_config.client_secret or both sdk_config.existingSecret.name and sdk_config.existingSecret.key")
}

func (s *PlatformChartTemplateSuite) Test_Mode_Core_No_Kas_Volumes_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
			"mode":      "core",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			if volume.Secret.SecretName == "kas-private-keys" {
				volumeFound = true
			}
		}
	}
	s.Require().False(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "kas-private-keys" {
				volumeMountFound = true
			}
		}
	}
	s.Require().False(volumeMountFound)
}

func (s *PlatformChartTemplateSuite) Test_Mode_Core_And_Kas_Volumes_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
			"mode":      "core\\,kas",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			if volume.Secret.SecretName == "kas-private-keys" {
				volumeFound = true
			}
		}
	}
	s.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "kas-private-keys" {
				volumeMountFound = true
			}
		}
	}
	s.Require().True(volumeMountFound)
}

func (s *PlatformChartTemplateSuite) Test_Mode_All_Kas_Volumes_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
			"mode":      "all",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			if volume.Secret.SecretName == "kas-private-keys" {
				volumeFound = true
			}
		}
	}
	s.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "kas-private-keys" {
				volumeMountFound = true
			}
		}
	}
	s.Require().True(volumeMountFound)
}

func (s *PlatformChartTemplateSuite) Test_Mode_Kas_Expect_Volumes_Mounted() {
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

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			if volume.Secret.SecretName == "kas-private-keys" {
				volumeFound = true
			}
		}
	}
	s.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "kas-private-keys" {
				volumeMountFound = true
			}
		}
	}
	s.Require().True(volumeMountFound)
}

func (s *PlatformChartTemplateSuite) Test_Trusted_Cert_Volume_Exists_When_AdditionalTrustedCertsDefined() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
			"server.tls.additionalTrustedCerts[0].secret.name": "test",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "trusted-certs" {
			volumeFound = true
		}
	}
	s.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "trusted-certs" {
				volumeMountFound = true
			}
		}
	}
	s.Require().True(volumeMountFound)
}

func (s *PlatformChartTemplateSuite) Test_Trusted_Cert_Volume_Does_Not_Exist_When_AdditionalTrustedCertsDefined_OR_Playground_Not_Defined() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "trusted-certs" {
			volumeFound = true
		}
	}
	s.Require().False(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "trusted-certs" {
				volumeMountFound = true
			}
		}
	}
	s.Require().False(volumeMountFound)
}

func (s *PlatformChartTemplateSuite) Test_Custom_Volume_Templates_Merged() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"volumeTemplate":      "platform.volumes.test.tpl",
			"volumeMountTemplate": "platform.volumeMounts.test.tpl",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "testVolume" {
			volumeFound = true
		}
	}
	s.Require().True(volumeFound)

	volumeMountFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == "testVolumeMount" {
				volumeMountFound = true
			}
		}
	}
	s.Require().True(volumeMountFound)
}

func (s *PlatformChartTemplateSuite) Test_Config_Map_Tracing() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	traceOptions := map[string]interface{}{
		"enabled": true,
		"provider": map[string]interface{}{
			"name": "otlp",
		},
	}
	jsonBytes, err := json.MarshalIndent(traceOptions, "", "  ")
	require.NoError(s.T(), err)
	jsonValuesMap := make(map[string]string)
	jsonValuesMap["trace"] = string(jsonBytes)

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetJsonValues:  jsonValuesMap,
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})
	var cm corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &cm)

	var config map[string]interface{}
	s.Require().NoError(yaml3.Unmarshal([]byte(cm.Data["opentdf.yaml"]), &config))

	serverConfig, ok := config["server"].(map[string]interface{})
	s.Require().True(ok)
	traceConfig, ok := serverConfig["trace"].(map[string]interface{})
	s.Require().True(ok)
	s.Require().True(traceConfig["enabled"].(bool))
}

func (s *PlatformChartTemplateSuite) Test_Custom_Config_Template_Services_Merged() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"configTemplate": "platform.configuration.test.tpl",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})
	var cm corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &cm)

	var config map[string]interface{}
	s.Require().NoError(yaml3.Unmarshal([]byte(cm.Data["opentdf.yaml"]), &config))

	s.Require().Equal(releaseName+"-platform", cm.Name)

	testServiceKeyFound := false
	for key := range config {
		if key == "services" {
			for sKey := range config[key].(map[string]interface{}) {
				if sKey == "testService" {
					testServiceKeyFound = true
				}
			}
		}
	}
	s.Require().True(testServiceKeyFound)
}

func (s *PlatformChartTemplateSuite) Test_TLS_Enabled_Expect_HTTPS_AppProtocol() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"server.tls.enabled": "true",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/service.yaml"})
	var svc corev1.Service
	helm.UnmarshalK8SYaml(s.T(), output, &svc)

	for _, port := range svc.Spec.Ports {
		s.Require().Equal("https", *port.AppProtocol)
	}

	output = helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, port := range container.Ports {
			s.Require().Equal("https", port.Name)
		}
		s.Require().Equal("https", container.ReadinessProbe.HTTPGet.Port.String())
		s.Require().Equal("https", container.LivenessProbe.HTTPGet.Port.String())
	}
}

func (s *PlatformChartTemplateSuite) Test_TLS_Disabled_Generic_K8S_Expect_K8S_H2C_AppProtocol() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/service.yaml"})
	var svc corev1.Service
	helm.UnmarshalK8SYaml(s.T(), output, &svc)

	for _, port := range svc.Spec.Ports {
		s.Require().Equal("kubernetes.io/h2c", *port.AppProtocol)
	}

	output = helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, port := range container.Ports {
			s.Require().Equal("http2", port.Name)
		}
		s.Require().Equal("http2", container.ReadinessProbe.HTTPGet.Port.String())
		s.Require().Equal("http2", container.LivenessProbe.HTTPGet.Port.String())
	}
}

func (s *PlatformChartTemplateSuite) Test_TLS_Disabled_Openshift_Expect_H2C_AppProtocol() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/service.yaml"}, "--api-versions", "security.openshift.io/v1/SecurityContextConstraints")
	var svc corev1.Service
	helm.UnmarshalK8SYaml(s.T(), output, &svc)

	for _, port := range svc.Spec.Ports {
		s.Require().Equal("h2c", *port.AppProtocol)
	}
}

func (s *PlatformChartTemplateSuite) Test_DB_Required_Expect_EnvVars_Set() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	envVarFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "OPENTDF_DB_PASSWORD" {
				envVarFound = true
			}
		}
	}
	s.Require().True(envVarFound)
}

func (s *PlatformChartTemplateSuite) Test_DB_Not_Required_Expect_EnvVars_Not_Set() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"db.required": "false",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	envVarFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "OPENTDF_DB_PASSWORD" {
				envVarFound = true
			}
		}
	}
	s.Require().False(envVarFound)
}

func (s *PlatformChartTemplateSuite) Test_PBD_Not_Enabled() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":                   "latest",
			"sdk_config.client_secret":    "test",
			"podDisruptionBudget.enabled": "false",
		},
	}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, releaseName, []string{"templates/poddisruptionbudget.yaml"})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "could not find template templates/poddisruptionbudget.yaml in chart")
}

func (s *PlatformChartTemplateSuite) Test_PBD_Enabled() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":                        "latest",
			"sdk_config.client_secret":         "test",
			"podDisruptionBudget.enabled":      "true",
			"podDisruptionBudget.minAvailable": "1",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/poddisruptionbudget.yaml"})

	var pdb policyv1.PodDisruptionBudget
	helm.UnmarshalK8SYaml(s.T(), output, &pdb)

	s.Require().Equal(pdb.Spec.Selector.MatchLabels["app.kubernetes.io/name"], "platform")
	s.Require().Equal(pdb.Spec.Selector.MatchLabels["app.kubernetes.io/instance"], releaseName)
	oneIntStr := intstr.FromInt(1)
	s.Require().Equal(pdb.Spec.MinAvailable, &oneIntStr)
	var nilIntOrString *intstr.IntOrString = nil
	s.Require().Equal(pdb.Spec.MaxUnavailable, nilIntOrString)
}

func (s *PlatformChartTemplateSuite) Test_SDK_Config_Is_Not_Set_When_Client_ID_Is_Empty() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	s.Require().Contains(config.Data, "opentdf.yaml")
	configData := config.Data["opentdf.yaml"]
	s.Require().NotEmpty(configData, "opentdf.yaml data should not be empty")

	var parsedConfig map[string]interface{}
	err := yaml3.Unmarshal([]byte(configData), &parsedConfig)
	s.Require().NoError(err, "Failed to unmarshal opentdf.yaml content")

	// Assert that the 'sdk_config' key is NOT present in the parsed config
	_, sdkConfigExists := parsedConfig["sdk_config"]
	s.Require().False(sdkConfigExists, "sdk_config key should not exist when client_id is not set")

}

func (s *PlatformChartTemplateSuite) Test_SDK_Config_Is_Set_When_Client_ID_Is_Not_Empty() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"sdk_config.client_id":     "test-client-id",     // Set client_id to a non-empty value
			"sdk_config.client_secret": "test-client-secret", // Set client_secret to a non-empty value,
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	s.Require().Contains(config.Data, "opentdf.yaml")
	configData := config.Data["opentdf.yaml"]
	s.Require().NotEmpty(configData, "opentdf.yaml data should not be empty")

	var parsedConfig map[string]interface{}
	err := yaml3.Unmarshal([]byte(configData), &parsedConfig)
	s.Require().NoError(err, "Failed to unmarshal opentdf.yaml content")

	// Assert that the 'sdk_config' key is present in the parsed config
	sdkConfig, sdkConfigExists := parsedConfig["sdk_config"]
	s.Require().True(sdkConfigExists, "sdk_config key should exist when client_id is set")
	s.Require().IsType(map[string]interface{}{}, sdkConfig, "sdk_config should be a map")
}

func (s *PlatformChartTemplateSuite) Test_Core_SDK_Connection_Is_Set_When_Endpoint_Is_Not_Empty() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"sdk_config.client_id":                 "test-client-id",     // Set client_id to a non-empty value
			"sdk_config.client_secret":             "test-client-secret", // Set client_secret to a non-empty value
			"sdk_config.connections.core.endpoint": "test-endpoint",      // Set endpoint to a non-empty value
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	s.Require().Contains(config.Data, "opentdf.yaml")
	configData := config.Data["opentdf.yaml"]
	s.Require().NotEmpty(configData, "opentdf.yaml data should not be empty")

	var parsedConfig map[string]interface{}
	err := yaml3.Unmarshal([]byte(configData), &parsedConfig)
	s.Require().NoError(err, "Failed to unmarshal opentdf.yaml content")

	// Assert that the 'sdk_config' key is present in the parsed config
	sdkConfig, sdkConfigExists := parsedConfig["sdk_config"]
	s.Require().True(sdkConfigExists, "sdk_config key should exist when client_id is set")
	s.Require().IsType(map[string]interface{}{}, sdkConfig, "sdk_config should be a map")

	// Assert that the 'sdk_config.core.endpoint' key is present and has the correct value
	endpoint, endpointExists := sdkConfig.(map[string]interface{})["core"].(map[string]interface{})["endpoint"]
	s.Require().True(endpointExists, "sdk_config.core.endpoint key should exist when endpoint is set")
	s.Require().Equal("test-endpoint", endpoint, "sdk_config.core.endpoint should have the correct value")
	s.Require().IsType("", endpoint, "sdk_config.core.endpoint should be a string")
}

func (s *PlatformChartTemplateSuite) Test_Two_SDK_Config_Connections_Are_Set_When_Endpoints_Are_Not_Empty() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"sdk_config.connections.core.endpoint":  "test-endpoint",      // Set endpoint to a non-empty value
			"sdk_config.connections.core2.endpoint": "test-endpoint2",     // Set another endpoint to a non-empty value
			"sdk_config.client_id":                  "test-client-id",     // Set client_id to a non-empty value
			"sdk_config.client_secret":              "test-client-secret", // Set client_secret to a non-empty value
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	s.Require().Contains(config.Data, "opentdf.yaml")
	configData := config.Data["opentdf.yaml"]
	s.Require().NotEmpty(configData, "opentdf.yaml data should not be empty")

	var parsedConfig map[string]interface{}
	err := yaml3.Unmarshal([]byte(configData), &parsedConfig)
	s.Require().NoError(err, "Failed to unmarshal opentdf.yaml content")

	// Assert that the 'sdk_config' key is present in the parsed config
	sdkConfig, sdkConfigExists := parsedConfig["sdk_config"]
	s.Require().True(sdkConfigExists, "sdk_config key should exist when client_id is set")
	s.Require().IsType(map[string]interface{}{}, sdkConfig, "sdk_config should be a map")

	// Assert that the 'sdk_config.core.endpoint' key is present and has the correct value
	endpoint, endpointExists := sdkConfig.(map[string]interface{})["core"].(map[string]interface{})["endpoint"]
	s.Require().True(endpointExists, "sdk_config.core.endpoint key should exist when endpoint is set")
	s.Require().Equal("test-endpoint", endpoint, "sdk_config.core.endpoint should have the correct value")
	s.Require().IsType("", endpoint, "sdk_config.core.endpoint should be a string")
	// Assert that the 'sdk_config.core2.endpoint' key is present and has the correct value
	endpoint2, endpointExists := sdkConfig.(map[string]interface{})["core2"].(map[string]interface{})["endpoint"]
	s.Require().True(endpointExists, "sdk_config.core2.endpoint key should exist when endpoint is set")
	s.Require().Equal("test-endpoint2", endpoint2, "sdk_config.core2.endpoint should have the correct value")
	s.Require().IsType("", endpoint2, "sdk_config.core2.endpoint should be a string")
}

func (s *PlatformChartTemplateSuite) Test_KeyManagement_Enabled_Without_RootKeySecret_Expect_Error() {
	releaseName := "key-management-no-secret"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"services.kas.config.preview.key_management": "true",
			"services.kas.root_key_secret.name":          "",
			"services.kas.root_key_secret.key":           "",
		},
	}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, releaseName, []string{})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "When services.kas.config.preview.key_management is true, you must set both services.kas.root_key_secret.name and services.kas.root_key_secret.key")
}

func (s *PlatformChartTemplateSuite) Test_KeyManagement_Enabled_With_RootKeySecret_Expect_EnvVar_Set() {
	releaseName := "key-management-with-secret"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"services.kas.config.preview.key_management": "true",
			"services.kas.root_key_secret.name":          "my-root-key-secret",
			"services.kas.root_key_secret.key":           "my-root-key",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	envVarFound := false
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "OPENTDF_KAS_ROOT_KEY" {
				s.Require().Equal("my-root-key-secret", envVar.ValueFrom.SecretKeyRef.Name)
				s.Require().Equal("my-root-key", envVar.ValueFrom.SecretKeyRef.Key)
				envVarFound = true
			}
		}
	}
	s.Require().True(envVarFound)
}

func (s *PlatformChartTemplateSuite) Test_Kas_PrivateKeySecret_Coalesce_NewValueTakesPrecedence() {
	releaseName := "kas-secret-coalesce-new"
	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"mode":                             "kas",
			"sdk_config.client_id":             "test",
			"sdk_config.client_secret":         "test",
			"services.kas.private_keys_secret": "new-secret",
			"services.kas.privateKeysSecret":   "old-secret",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "kas-private-keys" {
			s.Require().Equal("new-secret", volume.Secret.SecretName)
			volumeFound = true
		}
	}
	s.Require().True(volumeFound, "Volume 'kas-private-keys' not found")
}

func (s *PlatformChartTemplateSuite) Test_Kas_PrivateKeySecret_Coalesce_FallbackToOldValue() {
	releaseName := "kas-secret-coalesce-old"
	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"mode":                             "kas",
			"sdk_config.client_id":             "test",
			"sdk_config.client_secret":         "test",
			"services.kas.private_keys_secret": "", // Default empty value
			"services.kas.privateKeysSecret":   "old-secret",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	volumeFound := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "kas-private-keys" {
			s.Require().Equal("old-secret", volume.Secret.SecretName)
			volumeFound = true
		}
	}
	s.Require().True(volumeFound, "Volume 'kas-private-keys' not found")
}

func (s *PlatformChartTemplateSuite) Test_GRPC_Option_Override_maxRecvMsgSize() {
	releaseName := "basic"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetStrValues: map[string]string{
			"configFileKey":                  "my-config",
			"server.grpc.maxCallRecvMsgSize": fmt.Sprintf("%d", 10*1024*1024), // 10 MB
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["my-config.yaml"]
	s.Require().True(ok, "config map has my-config.yaml")

	s.Require().Contains(data, "maxCallRecvMsgSize", "maxCallRecvMsgSize should be set in the config file")
	s.Require().Contains(data, fmt.Sprintf("%d", 10*1024*1024), "maxCallRecvMsgSize should be set to 10 MB in the config file")
	s.Require().NotContains(data, "maxCallSendMsgSize", "maxCallSendMsgSize should not be set in the config file, as it is not overridden")
}

func (s *PlatformChartTemplateSuite) Test_GRPC_Option_Override_maxSendMsgSize() {
	releaseName := "basic"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetStrValues: map[string]string{
			"configFileKey":                  "my-config",
			"server.grpc.maxCallSendMsgSize": fmt.Sprintf("%d", 10*1024*1024), // 10 MB
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["my-config.yaml"]
	s.Require().True(ok, "config map has my-config.yaml")

	s.Require().Contains(data, "maxCallSendMsgSize", "maxCallSendMsgSize should be set in the config file")
	s.Require().Contains(data, fmt.Sprintf("%d", 10*1024*1024), "maxSendMsgSize should be set to 10 MB in the config file")
	s.Require().NotContains(data, "maxCallRecvMsgSize", "maxCallRecvMsgSize should not be set in the config file, as it is not overridden")
}

func (s *PlatformChartTemplateSuite) Test_HTTP_Server_Option_Override() {
	releaseName := "basic"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetStrValues: map[string]string{
			"configFileKey":            "my-config",
			"server.http.readTimeout":  "21s",
			"server.http.writeTimeout": "23s",
		},
		SetValues: map[string]string{
			"server.http.maxHeaderBytes": "2048",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["my-config.yaml"]
	s.Require().True(ok, "config map has my-config.yaml")

	s.Require().Contains(data, "readTimeout: 21s", "readTimeout should be set in the config file")
	s.Require().Contains(data, "writeTimeout: 23s", "writeTimeout should be set in the config file")
	s.Require().Contains(data, "maxHeaderBytes: 2048", "maxHeaderBytes should be set in the config file")
	s.Require().Contains(data, "idleTimeout: null", "idleTimeout should be null in the config file")
	s.Require().Contains(data, "readHeaderTimeout: null", "readHeaderTimeout should be null in the config file")
}

func (s *PlatformChartTemplateSuite) Test_HTTP_Server_Option_PublicHostname() {
	releaseName := "basic"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetStrValues: map[string]string{
			"server.public_hostname": "test.invalid",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["opentdf.yaml"]
	s.Require().True(ok, "config map has opentdf.yaml")

	s.Contains(data, "public_hostname: \"test.invalid\"", "public_hostname should be set in the config file")
}
func (s *PlatformChartTemplateSuite) Test_HTTP_Server_Option_NoPublicHostname() {
	releaseName := "basic"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["opentdf.yaml"]
	s.Require().True(ok, "config map has opentdf.yaml")

	s.NotContains(data, "public_hostname: test.invalid", "public_hostname should not be set in the config file")
}

func (s *PlatformChartTemplateSuite) Test_Registered_KAS_URI_Present() {
	releaseName := "basic"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"services.kas.config.registered_kas_uri": "https://kas.example.com:8080",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["opentdf.yaml"]
	s.Require().True(ok, "config map has opentdf.yaml")

	s.Contains(data, "registered_kas_uri: https://kas.example.com:8080", "registered_kas_uri should be set in the config file")
}

func (s *PlatformChartTemplateSuite) Test_Registered_KAS_URI_Not_Present() {
	releaseName := "basic"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["opentdf.yaml"]
	s.Require().True(ok, "config map has opentdf.yaml")

	s.NotContains(data, "registered_kas_uri: https://kas.example.com:8080", "registered_kas_uri should not be set in the config file")
}

func (s *PlatformChartTemplateSuite) Test_InitContainers_Not_Present_When_Empty() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	s.Require().Nil(deployment.Spec.Template.Spec.InitContainers, "InitContainers should not be present when not configured")
}

func (s *PlatformChartTemplateSuite) Test_InitContainers_Present_When_Configured() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":                    "latest",
			"initContainers[0].name":       "database-migration",
			"initContainers[0].image":      "registry.opentdf.io/platform:latest",
			"initContainers[0].command[0]": "/bin/sh",
			"initContainers[0].args[0]":    "-c",
			"initContainers[0].args[1]":    "platform migrate up",
			"initContainers[1].name":       "setup-data",
			"initContainers[1].image":      "busybox:latest",
			"initContainers[1].command[0]": "/bin/sh",
			"initContainers[1].args[0]":    "-c",
			"initContainers[1].args[1]":    "echo 'Setting up initial data...'",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	s.Require().NotNil(deployment.Spec.Template.Spec.InitContainers, "InitContainers should be present when configured")
	s.Require().Len(deployment.Spec.Template.Spec.InitContainers, 2, "Should have 2 init containers")

	// Verify first init container
	firstContainer := deployment.Spec.Template.Spec.InitContainers[0]
	s.Require().Equal("database-migration", firstContainer.Name)
	s.Require().Equal("registry.opentdf.io/platform:latest", firstContainer.Image)
	s.Require().Equal([]string{"/bin/sh"}, firstContainer.Command)
	s.Require().Equal([]string{"-c", "platform migrate up"}, firstContainer.Args)

	// Verify second init container
	secondContainer := deployment.Spec.Template.Spec.InitContainers[1]
	s.Require().Equal("setup-data", secondContainer.Name)
	s.Require().Equal("busybox:latest", secondContainer.Image)
	s.Require().Equal([]string{"/bin/sh"}, secondContainer.Command)
	s.Require().Equal([]string{"-c", "echo 'Setting up initial data...'"}, secondContainer.Args)
}

func (s *PlatformChartTemplateSuite) Test_CORS_AdditionalHeaders_Rendered() {
	releaseName := "cors-additional-headers"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"server.cors.additionalheaders[0]": "X-Custom-Header",
			"server.cors.additionalheaders[1]": "X-Another-Header",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["opentdf.yaml"]
	s.Require().True(ok, "config map has opentdf.yaml")

	s.Contains(data, "additionalheaders:", "additionalheaders should be present in config")
	s.Contains(data, "X-Custom-Header", "X-Custom-Header should be in additionalheaders")
	s.Contains(data, "X-Another-Header", "X-Another-Header should be in additionalheaders")
}

func (s *PlatformChartTemplateSuite) Test_CORS_AdditionalMethods_Rendered() {
	releaseName := "cors-additional-methods"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"server.cors.additionalmethods[0]": "PATCH",
			"server.cors.additionalmethods[1]": "CUSTOM",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["opentdf.yaml"]
	s.Require().True(ok, "config map has opentdf.yaml")

	s.Contains(data, "additionalmethods:", "additionalmethods should be present in config")
	s.Contains(data, "PATCH", "PATCH should be in additionalmethods")
	s.Contains(data, "CUSTOM", "CUSTOM should be in additionalmethods")
}

func (s *PlatformChartTemplateSuite) Test_CORS_AdditionalExposedHeaders_Rendered() {
	releaseName := "cors-additional-exposed"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"server.cors.additionalexposedheaders[0]": "X-Request-ID",
			"server.cors.additionalexposedheaders[1]": "X-Trace-ID",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["opentdf.yaml"]
	s.Require().True(ok, "config map has opentdf.yaml")

	s.Contains(data, "additionalexposedheaders:", "additionalexposedheaders should be present in config")
	s.Contains(data, "X-Request-ID", "X-Request-ID should be in additionalexposedheaders")
	s.Contains(data, "X-Trace-ID", "X-Trace-ID should be in additionalexposedheaders")
}

func (s *PlatformChartTemplateSuite) Test_CORS_AdditionalFields_Not_Rendered_When_Empty() {
	releaseName := "cors-empty-additional"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["opentdf.yaml"]
	s.Require().True(ok, "config map has opentdf.yaml")

	s.NotContains(data, "additionalheaders:", "additionalheaders should not be present when empty")
	s.NotContains(data, "additionalmethods:", "additionalmethods should not be present when empty")
	s.NotContains(data, "additionalexposedheaders:", "additionalexposedheaders should not be present when empty")
}

func (s *PlatformChartTemplateSuite) Test_CORS_AllowedAndAdditionalHeaders_Both_Rendered() {
	releaseName := "cors-both-headers"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"server.cors.allowedheaders[0]":    "Authorization",
			"server.cors.allowedheaders[1]":    "Content-Type",
			"server.cors.additionalheaders[0]": "X-Custom-Header",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["opentdf.yaml"]
	s.Require().True(ok, "config map has opentdf.yaml")

	s.Contains(data, "allowedheaders:", "allowedheaders should be present in config")
	s.Contains(data, "Authorization", "Authorization should be in allowedheaders")
	s.Contains(data, "Content-Type", "Content-Type should be in allowedheaders")
	s.Contains(data, "additionalheaders:", "additionalheaders should be present in config")
	s.Contains(data, "X-Custom-Header", "X-Custom-Header should be in additionalheaders")
}
