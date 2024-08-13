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
	chartPath string
}

func TestPlatformChartTemplateSuite(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping platform template test not in short mode.")
	}
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
			"image.tag":               "latest",
			"sdk_config.clientsecret": "test",
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
			"sdk_config.clientsecret":        "test",
			"sdk_config.existingSecret.name": "test",
			"sdk_config.existingSecret.key":  "test",
		},
	}

	_, err := helm.RenderTemplateE(suite.T(), options, suite.chartPath, releaseName, []string{})
	suite.Require().Error(err)
	suite.Require().ErrorContains(err, "You cannot set both clientsecret and existingSecret in sdk_config.")
}
