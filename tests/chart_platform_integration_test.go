package test

import (
	"encoding/json"
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
	releaseName := "opentdf"

	// Setup the kubectl config and context. Here we choose to use the defaults, which is:
	// - HOME/.kube/config for the kubectl config file
	// - Current context of the kubectl config file
	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)

	k8s.CreateNamespace(suite.T(), kubectlOptions, namespaceName)

	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		SetValues: map[string]string{
			"playground":               "true",
			"keycloak.ingress.enabled": "false",
			"server.auth.issuer":       "https://keycloak.opentdf.local/realms/opentdf",
			"server.tls.additionalTrustedCerts[0].secret.name":          "platform-tls",
			"server.tls.additionalTrustedCerts[0].secret.optional":      "false",
			"server.tls.additionalTrustedCerts[0].secret.items[0].key":  "tls.crt",
			"server.tls.additionalTrustedCerts[0].secret.items[0].path": "traefik.crt",
		},
	}

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

	// Apply tls secret
	k8s.RunKubectl(suite.T(), kubectlOptions, "create", "secret", "tls", "platform-tls", "--cert=../tls.crt", "--key=../tls.key")

	// Install the chart
	helm.Install(suite.T(), options, suite.chartPath, releaseName)

	defer func() {
		secret := k8s.GetSecret(suite.T(), kubectlOptions, "platform-tls")
		secretJson, _ := json.MarshalIndent(secret, "", "  ")
		fmt.Println("TLS Secret: ", string(secretJson))
		pods := k8s.ListPods(suite.T(), kubectlOptions, metav1.ListOptions{})
		for _, pod := range pods {
			if strings.Contains(pod.Name, "opentdf-platform") {
				fmt.Println("Pod Name: ", pod.Name)
				fmt.Println("Pod Status: ", pod.Status.Phase)
				fmt.Println("Pod Reason: ", pod.Status.Reason)
				podJson, _ := json.MarshalIndent(pod, "", "  ")
				fmt.Println("Pod: ", string(podJson))
				platLogs := k8s.GetPodLogs(suite.T(), kubectlOptions, &pod, "platform")
				fmt.Println("Platform Logs: ", platLogs)
			}
		}
		helm.Delete(suite.T(), options, releaseName, true)
		k8s.DeleteNamespace(suite.T(), kubectlOptions, namespaceName)
	}()

	kcServiceName := "platform-keycloak"

	k8s.WaitUntilServiceAvailable(suite.T(), kubectlOptions, kcServiceName, 10, 1*time.Second)

	platServiceName := fmt.Sprintf("%s-platform", releaseName)

	k8s.RunKubectl(suite.T(), kubectlOptions, "rollout", "restart", "deployment", platServiceName)

	k8s.WaitUntilServiceAvailable(suite.T(), kubectlOptions, platServiceName, 10, 1*time.Second)

	pods := k8s.ListPods(suite.T(), kubectlOptions, metav1.ListOptions{})
	// suite.Require().Len(pods, 3)
	for _, pod := range pods {
		k8s.WaitUntilPodAvailable(suite.T(), kubectlOptions, pod.Name, 30, 1*time.Second)
		suite.Require().Equal(pod.Status.Phase, corev1.PodRunning, fmt.Sprintf("Pod %s is not running", pod.Name))
	}

	// Get Ingress Resources
	ingresses := k8s.ListIngresses(suite.T(), kubectlOptions, metav1.ListOptions{})
	suite.Require().Len(ingresses, 0)

	traefikIngressCfg, err := filepath.Abs("traefik.yaml")
	suite.Require().NoError(err)

	k8s.KubectlApply(suite.T(), kubectlOptions, traefikIngressCfg)

	// Provision Keycloak
	dockerRun := exec.Command("docker", "run", "--rm", "--network=host", "-v", "./platform/service/cmd/keycloak_data.yaml:/keycloak_data.yaml", "registry.opentdf.io/platform:nightly", "provision", "keycloak", "-e", "https://keycloak.opentdf.local", "-f", "/keycloak_data.yaml")
	dockerRunOutput, err := dockerRun.CombinedOutput()
	suite.Require().NoError(err, string(dockerRunOutput))

	// Run bats tests
	batsTestFile, err := filepath.Abs("bats/tutorial.bats")
	suite.Require().NoError(err)

	cmd := exec.Command("bats", batsTestFile)
	output, err := cmd.CombinedOutput()
	suite.Require().NoError(err, string(output))
}
