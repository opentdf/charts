package test

import (
	"strings"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	yaml3 "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

// These tests lock in the config-rendering fixes from
// "fix(platform): suppress empty service blocks and fix auth.enabled bool rendering":
//   - server.auth.enabled honors an explicit `false` (Helm's `default true` used to coerce it back to true)
//   - server.cryptoProvider is omitted when standard.keys is empty
//   - services.entityresolution is omitted unless a url is configured
//   - services.authorization is omitted when left as the default empty map

// renderOpentdfConfig renders templates/config.yaml and returns the parsed opentdf.yaml document.
func (s *PlatformChartTemplateSuite) renderOpentdfConfig(options *helm.Options, releaseName string) map[string]interface{} {
	output := helm.RenderTemplate(s.T(), options, s.chartPath, releaseName, []string{"templates/config.yaml"})

	var cm corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &cm)

	data, ok := cm.Data["opentdf.yaml"]
	s.Require().True(ok, "config map should contain opentdf.yaml")
	s.Require().NotEmpty(data, "opentdf.yaml data should not be empty")

	var config map[string]interface{}
	s.Require().NoError(yaml3.Unmarshal([]byte(data), &config), "failed to unmarshal opentdf.yaml")
	return config
}

func (s *PlatformChartTemplateSuite) serverBlock(config map[string]interface{}) map[string]interface{} {
	server, ok := config["server"].(map[string]interface{})
	s.Require().True(ok, "server block should be a map")
	return server
}

func (s *PlatformChartTemplateSuite) servicesBlock(config map[string]interface{}) map[string]interface{} {
	services, ok := config["services"].(map[string]interface{})
	s.Require().True(ok, "services block should be a map")
	return services
}

func (s *PlatformChartTemplateSuite) configOptions(setValues, setJSONValues map[string]string) *helm.Options {
	return &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", "opentdf-"+strings.ToLower(random.UniqueId())),
		SetValues:      setValues,
		SetJsonValues:  setJSONValues,
	}
}

func (s *PlatformChartTemplateSuite) Test_Server_Auth_Enabled_Defaults_To_True() {
	config := s.renderOpentdfConfig(s.configOptions(nil, nil), "auth-default")

	auth, ok := s.serverBlock(config)["auth"].(map[string]interface{})
	s.Require().True(ok, "server.auth should be a map")
	s.Require().Equal(true, auth["enabled"], "server.auth.enabled should default to true when unset")
}

func (s *PlatformChartTemplateSuite) Test_Server_Auth_Enabled_False_Is_Honored() {
	options := s.configOptions(map[string]string{"server.auth.enabled": "false"}, nil)
	config := s.renderOpentdfConfig(options, "auth-false")

	auth, ok := s.serverBlock(config)["auth"].(map[string]interface{})
	s.Require().True(ok, "server.auth should be a map")
	// Regression: `enabled | default true` coerced an explicit false back to true.
	s.Require().Equal(false, auth["enabled"], "server.auth.enabled: false must render as false")
}

func (s *PlatformChartTemplateSuite) Test_Server_Auth_Enabled_True_Is_Honored() {
	options := s.configOptions(map[string]string{"server.auth.enabled": "true"}, nil)
	config := s.renderOpentdfConfig(options, "auth-true")

	auth, ok := s.serverBlock(config)["auth"].(map[string]interface{})
	s.Require().True(ok, "server.auth should be a map")
	s.Require().Equal(true, auth["enabled"], "server.auth.enabled: true should render as true")
}

func (s *PlatformChartTemplateSuite) Test_CryptoProvider_Rendered_By_Default() {
	config := s.renderOpentdfConfig(s.configOptions(nil, nil), "crypto-default")

	_, ok := s.serverBlock(config)["cryptoProvider"]
	s.Require().True(ok, "server.cryptoProvider should render when standard.keys is populated")
}

func (s *PlatformChartTemplateSuite) Test_CryptoProvider_Absent_When_Keys_Empty() {
	options := s.configOptions(
		map[string]string{"mode": "core"},
		map[string]string{"server.cryptoProvider.standard.keys": "[]"},
	)
	config := s.renderOpentdfConfig(options, "crypto-empty")

	_, ok := s.serverBlock(config)["cryptoProvider"]
	s.Require().False(ok, "server.cryptoProvider should be omitted when standard.keys is empty")
}

// Test_Mode_Core_Renders_Valid_Config_With_Services_Omitted guards against the
// empty-services regression: when entityresolution, authorization, and kas are all
// suppressed (mode=core with defaults), the merged services block must not collapse
// into a dangling "{}" that yields invalid YAML. renderOpentdfConfig fails the test
// if opentdf.yaml does not parse.
func (s *PlatformChartTemplateSuite) Test_Mode_Core_Renders_Valid_Config_With_Services_Omitted() {
	config := s.renderOpentdfConfig(s.configOptions(map[string]string{"mode": "core"}, nil), "mode-core")

	_, ok := config["services"]
	s.Require().False(ok, "services should be omitted entirely (not rendered as '{}') when no services are configured")

	// The rest of the document must still be well-formed.
	_, ok = s.serverBlock(config)["auth"].(map[string]interface{})
	s.Require().True(ok, "server.auth should still render for mode=core")
}

func (s *PlatformChartTemplateSuite) Test_EntityResolution_Absent_When_Url_Empty() {
	config := s.renderOpentdfConfig(s.configOptions(nil, nil), "er-default")

	_, ok := s.servicesBlock(config)["entityresolution"]
	s.Require().False(ok, "services.entityresolution should be omitted when url is empty")
}

func (s *PlatformChartTemplateSuite) Test_EntityResolution_Rendered_When_Url_Set() {
	options := s.configOptions(map[string]string{"services.entityresolution.url": "https://idp.example.com"}, nil)
	config := s.renderOpentdfConfig(options, "er-url")

	er, ok := s.servicesBlock(config)["entityresolution"].(map[string]interface{})
	s.Require().True(ok, "services.entityresolution should render when url is set")
	s.Require().Equal("https://idp.example.com", er["url"], "entityresolution.url should match the configured value")
}

func (s *PlatformChartTemplateSuite) Test_Authorization_Absent_When_Default_Empty() {
	config := s.renderOpentdfConfig(s.configOptions(nil, nil), "authz-default")

	_, ok := s.servicesBlock(config)["authorization"]
	s.Require().False(ok, "services.authorization should be omitted when left as the default empty map")
}

func (s *PlatformChartTemplateSuite) Test_Authorization_Rendered_When_Set() {
	options := s.configOptions(map[string]string{"services.authorization.rego.path": "/etc/platform/entitlements.rego"}, nil)
	config := s.renderOpentdfConfig(options, "authz-set")

	authz, ok := s.servicesBlock(config)["authorization"].(map[string]interface{})
	s.Require().True(ok, "services.authorization should render when configured")
	rego, ok := authz["rego"].(map[string]interface{})
	s.Require().True(ok, "authorization.rego should be a map")
	s.Require().Equal("/etc/platform/entitlements.rego", rego["path"], "authorization.rego.path should match the configured value")
}
