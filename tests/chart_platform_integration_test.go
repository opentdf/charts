package test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/suite"
)

type PlatformChartIntegrationSuite struct {
	suite.Suite
	chartPath string
}

func TestPlatformChartIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping platform integration test in short mode.")
	}
	suite.Run(t, new(PlatformChartIntegrationSuite))
}

func (suite *PlatformChartIntegrationSuite) SetupTest() {
	path, err := filepath.Abs("../charts/platform")
	suite.Require().NoError(err)
	suite.chartPath = path
}

func (suite *PlatformChartIntegrationSuite) TestBasicDeployment() {
	namespaceName := fmt.Sprintf("opentdf-%s", strings.ToLower(random.UniqueId()))
	releaseName := "basic"

	// Setup the kubectl config and context. Here we choose to use the defaults, which is:
	// - HOME/.kube/config for the kubectl config file
	// - Current context of the kubectl config file
	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)

	k8s.CreateNamespace(suite.T(), kubectlOptions, namespaceName)

	defer k8s.DeleteNamespace(suite.T(), kubectlOptions, namespaceName)

	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		SetValues: map[string]string{
			"sdk_config.clientsecret": "test",
			"playground":              "true",
		},
	}

	defer helm.Delete(suite.T(), options, releaseName, true)

	helm.Install(suite.T(), options, suite.chartPath, releaseName)

	serviceName := fmt.Sprintf("%s-platform", releaseName)

	k8s.WaitUntilServiceAvailable(suite.T(), kubectlOptions, serviceName, 10, 1*time.Second)

}
