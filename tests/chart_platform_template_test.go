package test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
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

func (s *PlatformChartTemplateSuite) Test_Playground_Enabled_AND_Keycloak_Ing_Enabled_Trusted_Cert_Mounted() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":  "latest",
			"playground": "true",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	// Find projected volume trusted-certs and check if keycloak cert is mounted
	found := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Projected != nil {
			for _, source := range volume.Projected.Sources {
				s.T().Log("Secret Name: ", source.Secret.Name)
				if source.Secret != nil && source.Secret.Name == "keycloak.local-tls" {
					s.Require().Equal("ca.crt", source.Secret.Items[0].Key)
					s.Require().Equal("kc-ca.crt", source.Secret.Items[0].Path)
					found = true
				}
			}
		}
	}
	s.Require().True(found)
}

func (s *PlatformChartTemplateSuite) Test_Playground_Enabled_AND_Keycloak_Ing_Disabled_Trusted_Cert_Not_Mounted() {
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

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

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
	s.Require().False(found)
}

func (s *PlatformChartTemplateSuite) Test_Playground_Enabled_AND_Keycloak_Ing_Enabled_AND_TLS_Disabled_Trusted_Cert_Not_Mounted() {
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

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

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
	s.Require().False(found)
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

func (s *PlatformChartTemplateSuite) Test_Trusted_Cert_Volume_Exists_When_Playground_True() {
	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":  "latest",
			"playground": "true",
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
			"services.kas.config.preview_features.key_management": "true",
			"services.kas.root_key_secret.name":                   "",
			"services.kas.root_key_secret.key":                    "",
		},
	}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, releaseName, []string{})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "When services.kas.config.preview_features.key_management is true, you must set both services.kas.root_key_secret.name and services.kas.root_key_secret.key")
}

func (s *PlatformChartTemplateSuite) Test_KeyManagement_Enabled_With_RootKeySecret_Expect_EnvVar_Set() {
	releaseName := "key-management-with-secret"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"services.kas.config.preview_features.key_management": "true",
			"services.kas.root_key_secret.name":                   "my-root-key-secret",
			"services.kas.root_key_secret.key":                    "my-root-key",
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

func (s *PlatformChartIntegrationSuite) Test_GRPC_Option_Override() {
	releaseName := "basic"

	namespaceName := "platform-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetStrValues: map[string]string{
			"configFileKey":              "my-config",
			"server.grpc.maxRecvMsgSize": fmt.Sprintf("%d", 10*1024*1024), // 10 MB
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"charts/platform/templates/config.yaml"})

	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &config)

	data, ok := config.Data["my-config.yaml"]
	s.Require().True(ok, "config map has my-config.yaml")

	s.Require().Contains(data, "maxRecvMsgSize", "maxRecvMsgSize should be set in the config file")
	s.Require().Contains(data, fmt.Sprintf("%d", 10*1024*1024), "maxRecvMsgSize should be set to 10 MB in the config file")
	s.Require().NotContains(data, "maxSendMsgSize", "maxSendMsgSize should not be set in the config file, as it is not overridden")
}
