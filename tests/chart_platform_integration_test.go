package test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			"sdk_config.clientsecret":  "test",
			"playground":               "true",
			"keycloak.ingress.enabled": "false",
		},
	}

	defer helm.Delete(suite.T(), options, releaseName, true)

	// Generate KAS Keys
	privECKey, pubECKey, err := generateKasECDHKeyPair()
	suite.Require().NoError(err)
	privRSAKey, pubRSAKey, err := generateKasRSAKeyPair()
	suite.Require().NoError(err)

	k8s.RunKubectl(suite.T(), kubectlOptions, "create", "secret", "generic", "kas-private-keys",
		fmt.Sprintf("--from-literal=kas-ec-private.pem=%s", string(privECKey)),
		fmt.Sprintf("--from-literal=kas-ec-cert.pem=%s", string(pubECKey)),
		fmt.Sprintf("--from-literal=kas-private.pem=%s", string(privRSAKey)),
		fmt.Sprintf("--from-literal=kas-cert.pem=%s", string(pubRSAKey)),
	)

	kasSecret := k8s.GetSecret(suite.T(), kubectlOptions, "kas-private-keys")

	suite.Require().Equal(kasSecret.Data["kas-ec-private.pem"], privECKey)

	helm.Install(suite.T(), options, suite.chartPath, releaseName)

	serviceName := fmt.Sprintf("%s-platform", releaseName)

	k8s.WaitUntilServiceAvailable(suite.T(), kubectlOptions, serviceName, 10, 1*time.Second)

	pods := k8s.ListPods(suite.T(), kubectlOptions, metav1.ListOptions{})
	suite.Require().Len(pods, 3)
	for _, pod := range pods {
		suite.Require().Equal(pod.Status.Phase, corev1.PodRunning, fmt.Sprintf("Pod %s is not running", pod.Name))
	}

	// Get Ingress Resources
	ingresses := k8s.ListIngresses(suite.T(), kubectlOptions, metav1.ListOptions{})
	suite.Require().Len(ingresses, 0)

	// Apply tls secret
	k8s.RunKubectl(suite.T(), kubectlOptions, "create", "secret", "tls", "platform-tls", "--cert=../tls.crt", "--key=../tls.key")

	traefikIngressCfg, err := filepath.Abs("traefik.yaml")
	suite.Require().NoError(err)

	k8s.KubectlApply(suite.T(), kubectlOptions, traefikIngressCfg)

	// Run bats tests
	batsTestFile, err := filepath.Abs("bats/tutorial.bats")
	suite.Require().NoError(err)

	cmd := exec.Command("bats", batsTestFile)
	output, err := cmd.CombinedOutput()
	suite.Require().NoError(err, string(output))
}
