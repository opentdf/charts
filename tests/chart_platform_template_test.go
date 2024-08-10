package test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/suite"
	appv1 "k8s.io/api/apps/v1"
)

type PlatformChartTemplateSuite struct {
	suite.Suite
}

func TestPlatformChartTemplateSuite(t *testing.T) {
	suite.Run(t, new(PlatformChartTemplateSuite))
}

func (suite *PlatformChartTemplateSuite) TestBasicDeploymentTemplateRender() {

	helmChartPath, err := filepath.Abs("../charts/platform")
	suite.Require().NoError(err)

	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":               "latest",
			"sdk_config.clientsecret": "test",
		},
	}

	output, err := helm.RenderTemplateE(suite.T(), options, helmChartPath, releaseName, []string{"templates/deployment.yaml"})
	suite.Require().NoError(err)

	var deployment appv1.Deployment
	helm.UnmarshalK8SYaml(suite.T(), output, &deployment)

	suite.Require().Equal(deployment.Name, releaseName+"-platform")
	suite.Require().Len(deployment.Spec.Template.Spec.Containers, 1)
	suite.Require().Equal(deployment.Spec.Template.Spec.Containers[0].Image, "registry.opentdf.io/platform:latest")
}

func (suite *PlatformChartTemplateSuite) Test_Empty_SDK_Config_Client_Secret_AND_Existing_Secret_Expect_Error() {

	helmChartPath, err := filepath.Abs("../charts/platform")
	suite.Require().NoError(err)

	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag": "latest",
		},
	}

	_, err = helm.RenderTemplateE(suite.T(), options, helmChartPath, releaseName, []string{})
	suite.Require().Error(err)
	suite.Require().ErrorContains(err, "You must set either clientsecret and existingSecret in sdk_config.")

}

func (suite *PlatformChartTemplateSuite) Test_SDK_Config_Set_Client_Secret_AND_Existing_Secret_Expect_Error() {

	helmChartPath, err := filepath.Abs("../charts/platform")
	suite.Require().NoError(err)

	releaseName := "basic"

	namespaceName := "opentdf-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		SetValues: map[string]string{
			"image.tag":                      "latest",
			"sdk_config.clientsecret":        "test",
			"sdk_config.existingSecret.name": "test",
			"sdk_config.existingSecret.key":  "test",
		},
	}

	_, err = helm.RenderTemplateE(suite.T(), options, helmChartPath, releaseName, []string{})
	suite.Require().Error(err)
	suite.Require().ErrorContains(err, "You cannot set both clientsecret and existingSecret in sdk_config.")
}
