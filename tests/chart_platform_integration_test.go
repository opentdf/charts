package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/suite"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var traefikIngress = `---
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: platform
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(` + "`" + "keycloak.opentdf.local" + "`" + `)
      kind: Rule
      services:
        - name: platform-keycloak
          namespace: {{ .Namespace }}
          port: 80
          scheme: http
          passHostHeader: true
    - match: Host(` + "`" + "platform.opentdf.local" + "`" + `)
      kind: Rule
      services:
        - name: opentdf-platform
          namespace: {{ .Namespace }}
          port: 9000
          scheme: h2c
          passHostHeader: true
    - match: Host(` + "`" + "kas.opentdf.local" + "`" + `)
      kind: Rule
      services:
        - name: kas
          namespace: {{ .Namespace }}
          port: 9000
          scheme: h2c
          passHostHeader: true
  tls:
    secretName: platform-tls`

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
	suite.T().Log("Setting up test environment")
	var (
		namespaceName    = fmt.Sprintf("opentdf-%s", strings.ToLower(random.UniqueId()))
		releaseName      = "opentdf"
		kasReleaseName   = "kas"
		keycloakEndpoint = "https://keycloak.opentdf.local:9443"
		kcServiceName    = "platform-keycloak"
		platServiceName  = fmt.Sprintf("%s-platform", releaseName)
		kasServiceName   = "kas"
	)

	// Allow switching docker network
	platformDockerNetwork := os.Getenv("PLATFORM_DOCKER_NETWORK")
	if platformDockerNetwork == "" {
		platformDockerNetwork = "platform-k3d"
	}

	suite.T().Logf("Using docker network: %s", platformDockerNetwork)

	// Allow setting different keycloak config file path
	keycloakConfigFileLocation := os.Getenv("KEYCLOAK_CONFIG_FILE_LOCATION")
	if keycloakConfigFileLocation == "" {
		keycloakConfigFileLocation = "../platform/service/cmd/keycloak_data.yaml"
	}

	suite.T().Logf("Using keycloak config file location: %s", keycloakConfigFileLocation)

	// Setup the kubectl config and context. Here we choose to use the defaults, which is:
	// - HOME/.kube/config for the kubectl config file
	// - Current context of the kubectl config file
	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)

	k8s.CreateNamespace(suite.T(), kubectlOptions, namespaceName)

	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		ValuesFiles:    []string{"../charts/platform/values.yaml"},
		SetValues: map[string]string{
			"playground": "true",
			// Keycloak
			"keycloak.ingress.enabled": "false",
			// Keycloak Config Cli
			"keycloak.keycloakConfigCli.backoffLimit": "10",
			// Platform
			"logger.level":       "debug",
			"server.auth.issuer": fmt.Sprintf("%s/realms/opentdf", keycloakEndpoint),
			"server.tls.additionalTrustedCerts[0].secret.name":          "platform-tls",
			"server.tls.additionalTrustedCerts[0].secret.optional":      "false",
			"server.tls.additionalTrustedCerts[0].secret.items[0].key":  "tls.crt",
			"server.tls.additionalTrustedCerts[0].secret.items[0].path": "traefik.crt",
			"services.entityresolution.url":                             keycloakEndpoint,
			"services.entityresolution.clientid":                        "tdf-entity-resolution",
			"services.entityresolution.clientsecret":                    "secret",
			"services.entityresolution.realm":                           "opentdf",
			"services.kas.config.registered_kas_uri":                    "https://kas.opentdf.local:9443/realms/opentdf/protocol/openid-connect/token",
			"services.kas.config.preview.key_management":                "true",
			"services.kas.root_key_secret.name":                         "my-root-key-secret",
			"services.kas.root_key_secret.key":                          "my-root-key",
		},
	}

	// These are needed because of a java bug running on M4 Macs
	if os.Getenv("M4_DEVICE") != "" {
		options.SetValues["keycloak.extraEnvVars[1].name"] = "JAVA_OPTS"
		options.SetValues["keycloak.extraEnvVars[1].value"] = "-XX:UseSVE=0"
		options.SetValues["keycloak.keycloakConfigCli.command[0]"] = "java"
		options.SetValues["keycloak.keycloakConfigCli.command[1]"] = "-XX:UseSVE=0"
		options.SetValues["keycloak.keycloakConfigCli.command[2]"] = "-jar"
		options.SetValues["keycloak.keycloakConfigCli.command[3]"] = "./keycloak-config-cli.jar"
	}

	suite.T().Logf(
		"Deploying platform with options: %+v",
		options,
	)

	kasOptions := &helm.Options{
		KubectlOptions: kubectlOptions,
		ValuesFiles:    []string{"../charts/platform/values.yaml"},
		SetValues: map[string]string{
			"nameOverride":                                              "kas",
			"mode":                                                      "kas",
			"sdk_config.client_id":                                      "opentdf",
			"sdk_config.client_secret":                                  "secret",
			"sdk_config.connections.core.endpoint":                      fmt.Sprintf("http://opentdf-platform.%s.svc.cluster.local:9000", namespaceName),
			"sdk_config.connections.core.plaintext":                     "true",
			"server.auth.issuer":                                        fmt.Sprintf("%s/realms/opentdf", keycloakEndpoint),
			"server.tls.additionalTrustedCerts[0].secret.name":          "platform-tls",
			"server.tls.additionalTrustedCerts[0].secret.optional":      "false",
			"server.tls.additionalTrustedCerts[0].secret.items[0].key":  "tls.crt",
			"server.tls.additionalTrustedCerts[0].secret.items[0].path": "traefik.crt",
			"services.kas.config.registered_kas_uri":                    "https://kas.opentdf.local:9443/realms/opentdf/protocol/openid-connect/token",
			"services.kas.config.preview.key_management":                "true",
			"services.kas.root_key_secret.name":                         "my-root-key-secret",
			"services.kas.root_key_secret.key":                          "my-root-key",
		},
	}

	suite.T().Logf("Deploying KAS with options: %v", kasOptions)

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

	k8s.RunKubectl(suite.T(), kubectlOptions, "create", "secret", "generic", "my-root-key-secret",
		fmt.Sprintf("--from-literal=%s=%s", "my-root-key", "value"),
	)

	kasSecret := k8s.GetSecret(suite.T(), kubectlOptions, "kas-private-keys")

	suite.Require().Equal(kasSecret.Data["kas-ec-private.pem"], privECKey)

	// Apply tls secret
	k8s.RunKubectl(suite.T(), kubectlOptions, "create", "secret", "tls", "platform-tls", "--cert=../tls.crt", "--key=../tls.key")

	// Build and Apply Traefik IngressRoute
	var ingRendered bytes.Buffer
	ingTmpl, err := template.New("traefik").Parse(traefikIngress)
	suite.Require().NoError(err)
	err = ingTmpl.Execute(&ingRendered, map[string]string{"Namespace": namespaceName})
	suite.Require().NoError(err)
	err = os.WriteFile("traefik.yaml", ingRendered.Bytes(), 0o644)
	suite.Require().NoError(err)

	traefikIngressCfg, err := filepath.Abs("traefik.yaml")
	suite.Require().NoError(err)

	k8s.KubectlApply(suite.T(), kubectlOptions, traefikIngressCfg)

	// Get the IngressRoute and Check for all host names
	ingressRouteYaml, err := k8s.RunKubectlAndGetOutputE(suite.T(), kubectlOptions, "get", "ingressroute.traefik.io", "-n", namespaceName, "platform", "-ojson")
	suite.Require().NoError(err)
	var ingressRoute traefikv1alpha1.IngressRoute
	err = yaml.Unmarshal([]byte(ingressRouteYaml), &ingressRoute)
	suite.Require().NoError(err)

	hostnames := []string{"Host(`platform.opentdf.local`)", "Host(`kas.opentdf.local`)", "Host(`keycloak.opentdf.local`)"}
	for _, route := range ingressRoute.Spec.Routes {
		suite.T().Logf("Match: %s Port: %d Service: %s Scheme: %s", route.Match, int(route.Services[0].Port.IntVal), route.Services[0].Name, route.Services[0].Scheme)
		suite.Require().Contains(hostnames, route.Match)
	}

	// Install Core Platform
	helm.Install(suite.T(), options, suite.chartPath, releaseName)
	suite.T().Logf(
		"Core Platform installed successfully. Release name: %s, Namespace: %s",
		releaseName, namespaceName,
	)
	defer func() {
		if suite.T().Failed() {
			pods := k8s.ListPods(suite.T(), kubectlOptions, metav1.ListOptions{})
			fmt.Println("Number of Pods: ", len(pods))
			for _, pod := range pods {
				fmt.Println("Pod Name: ", pod.Name)
				fmt.Println("Pod Status: ", pod.Status.Phase)
				fmt.Println("Pod Reason: ", pod.Status.Reason)
				if strings.Contains(pod.Name, "opentdf-platform") || strings.Contains(pod.Name, "kas") {
					podJson, _ := json.MarshalIndent(pod, "", "  ")
					fmt.Println("Pod: ", string(podJson))
					platLogs := k8s.GetPodLogs(suite.T(), kubectlOptions, &pod, "platform")
					fmt.Println("Platform Logs: ", platLogs)
				}
			}
		}
		helm.Delete(suite.T(), options, releaseName, true)
		helm.Delete(suite.T(), kasOptions, kasReleaseName, true)
		k8s.DeleteNamespace(suite.T(), kubectlOptions, namespaceName)
	}()

	k8s.WaitUntilServiceAvailable(suite.T(), kubectlOptions, kcServiceName, 10, 10*time.Second)

	// Provision Keycloak
	kcSecret := k8s.GetSecret(suite.T(), kubectlOptions, "platform-keycloak")
	var kcAdminPass string
	// Allow override of KEYCLOAK_PASSWORD
	if kcAdminPass = os.Getenv("KEYCLOAK_PASSWORD"); kcAdminPass == "" {
		kcAdminPass = string(kcSecret.Data["admin-password"])
	}

	kcDataPath, err := filepath.Abs(keycloakConfigFileLocation)
	suite.Require().NoError(err)
	dockerRun := exec.Command("docker", "run", "--rm", fmt.Sprintf("--network=%s", platformDockerNetwork), "--add-host=keycloak.opentdf.local:10.255.127.1", "-v", fmt.Sprintf("%s:/keycloak_data.yaml", kcDataPath), "registry.opentdf.io/platform:latest", "provision", "keycloak", "-p", kcAdminPass, "-e", keycloakEndpoint, "-f", "/keycloak_data.yaml")
	dockerRunOutput, err := dockerRun.CombinedOutput()

	suite.Require().NoError(err, string(dockerRunOutput))
	if err == nil {
		suite.T().Log(string(dockerRunOutput))
	}

	suite.T().Logf("Keycloak provisioned successfully")

	k8s.RunKubectl(suite.T(), kubectlOptions, "rollout", "restart", "deployment", platServiceName)

	// Install KAS
	helm.Install(suite.T(), kasOptions, suite.chartPath, kasReleaseName)

	suite.T().Logf("KAS installed successfully")

	k8s.RunKubectl(suite.T(), kubectlOptions, "rollout", "restart", "deployment", kasServiceName)

	k8s.WaitUntilServiceAvailable(suite.T(), kubectlOptions, platServiceName, 10, 10*time.Second)
	k8s.WaitUntilServiceAvailable(suite.T(), kubectlOptions, kasServiceName, 10, 10*time.Second)

	suite.T().Logf(
		"Platform and KAS services are available")

	// Give everything time to settle
	time.Sleep(30 * time.Second)

	pods := k8s.ListPods(suite.T(), kubectlOptions, metav1.ListOptions{})

	for _, pod := range pods {
		suite.T().Logf("Found pod: %s", pod.Name)
	}

	// // suite.Require().Len(pods, 3)
	for _, pod := range pods {
		if strings.Contains(pod.Name, "opentdf-platform") || strings.Contains(pod.Name, "kas") {
			podLogs := k8s.GetPodLogs(suite.T(), kubectlOptions, &pod, "platform")
			suite.T().Logf("Pod %s Logs: %s", pod.Name, podLogs)
		}

		suite.T().Logf("Pod %s Status: %s Message: %s Reason: %s", pod.Name, pod.Status.Phase, pod.Status.Message, pod.Status.Reason)
		k8s.WaitUntilPodAvailable(suite.T(), kubectlOptions, pod.Name, 6, 10*time.Second)
		suite.Require().Equal(pod.Status.Phase, corev1.PodRunning, fmt.Sprintf("Pod %s is not running", pod.Name))
	}

	// Get Ingress Resources
	ingresses := k8s.ListIngresses(suite.T(), kubectlOptions, metav1.ListOptions{})
	suite.Require().Len(ingresses, 0)

	// Run bats tests
	suite.T().Setenv("KUBE_NAMESPACE", namespaceName)

	batsTestFile, err := filepath.Abs("bats/e2e.bats")
	suite.Require().NoError(err)

	cmd := exec.Command("bats", batsTestFile)
	output, err := cmd.CombinedOutput()
	suite.Require().NoError(err, string(output))
	if err == nil {
		suite.T().Log(string(output))
	}
}
