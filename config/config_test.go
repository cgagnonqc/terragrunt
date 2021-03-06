package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/remote"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/stretchr/testify/assert"
)

var mockOptions = options.NewTerragruntOptionsForTest("test-time-mock")
var mockOldInclude = IncludeConfig{Path: OldTerragruntConfigPath}

func TestParseTerragruntConfigRemoteStateMinimalConfig(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  remote_state {
    backend = "s3"
  }
}
`

	terragruntConfig, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, terragruntConfig.Terraform)

	if assert.NotNil(t, terragruntConfig.RemoteState) {
		assert.Equal(t, "s3", terragruntConfig.RemoteState.Backend)
		assert.Empty(t, terragruntConfig.RemoteState.Config)
	}
}

func TestParseTerragruntConfigRemoteStateMissingBackend(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  remote_state {
  }
}
`

	_, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	assert.True(t, errors.IsError(err, remote.RemoteBackendMissing), "Unexpected error of type %s: %s", reflect.TypeOf(err), err)
}

func TestParseTerragruntConfigRemoteStateFullConfig(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  remote_state {
    backend = "s3"
    config {
      encrypt = true
      bucket = "my-bucket"
      key = "terraform.tfstate"
      region = "us-east-1"
    }
  }
}
`

	terragruntConfig, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, terragruntConfig.Terraform)

	if assert.NotNil(t, terragruntConfig.RemoteState) {
		assert.Equal(t, "s3", terragruntConfig.RemoteState.Backend)
		assert.NotEmpty(t, terragruntConfig.RemoteState.Config)
		assert.Equal(t, true, terragruntConfig.RemoteState.Config["encrypt"])
		assert.Equal(t, "my-bucket", terragruntConfig.RemoteState.Config["bucket"])
		assert.Equal(t, "terraform.tfstate", terragruntConfig.RemoteState.Config["key"])
		assert.Equal(t, "us-east-1", terragruntConfig.RemoteState.Config["region"])
	}
}

func TestParseTerragruntConfigDependenciesOnePath(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  dependencies {
    paths = ["../vpc"]
  }
}
`

	terragruntConfig, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, terragruntConfig.RemoteState)
	assert.Nil(t, terragruntConfig.Terraform)

	if assert.NotNil(t, terragruntConfig.Dependencies) {
		assert.Equal(t, []string{"../vpc"}, terragruntConfig.Dependencies.Paths)
	}
}

func TestParseTerragruntConfigDependenciesMultiplePaths(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  dependencies {
    paths = ["../vpc", "../mysql", "../backend-app"]
  }
}
`

	terragruntConfig, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, terragruntConfig.RemoteState)
	assert.Nil(t, terragruntConfig.Terraform)

	if assert.NotNil(t, terragruntConfig.Dependencies) {
		assert.Equal(t, []string{"../vpc", "../mysql", "../backend-app"}, terragruntConfig.Dependencies.Paths)
	}
}

func TestParseTerragruntConfigRemoteStateDynamoDbTerraformConfigAndDependenciesFullConfig(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  terraform {
    source = "foo"
  }

  remote_state {
    backend = "s3"
    config {
      encrypt = true
      bucket = "my-bucket"
      key = "terraform.tfstate"
      region = "us-east-1"
    }
  }

  dependencies {
    paths = ["../vpc", "../mysql", "../backend-app"]
  }
}
`

	terragruntConfig, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	if err != nil {
		t.Fatal(err)
	}

	if assert.NotNil(t, terragruntConfig.Terraform) {
		assert.Equal(t, "foo", terragruntConfig.Terraform.Source)
	}

	if assert.NotNil(t, terragruntConfig.RemoteState) {
		assert.Equal(t, "s3", terragruntConfig.RemoteState.Backend)
		assert.NotEmpty(t, terragruntConfig.RemoteState.Config)
		assert.Equal(t, true, terragruntConfig.RemoteState.Config["encrypt"])
		assert.Equal(t, "my-bucket", terragruntConfig.RemoteState.Config["bucket"])
		assert.Equal(t, "terraform.tfstate", terragruntConfig.RemoteState.Config["key"])
		assert.Equal(t, "us-east-1", terragruntConfig.RemoteState.Config["region"])
	}

	if assert.NotNil(t, terragruntConfig.Dependencies) {
		assert.Equal(t, []string{"../vpc", "../mysql", "../backend-app"}, terragruntConfig.Dependencies.Paths)
	}
}

func TestParseTerragruntConfigRemoteStateDynamoDbTerraformConfigAndDependenciesFullConfigOldConfigFormat(t *testing.T) {
	t.Parallel()

	config := `
terraform {
  source = "foo"
}

remote_state {
  backend = "s3"
  config {
    encrypt = true
    bucket = "my-bucket"
    key = "terraform.tfstate"
    region = "us-east-1"
  }
}

dependencies {
  paths = ["../vpc", "../mysql", "../backend-app"]
}
`

	terragruntConfig, err := parseConfigString(config, mockOptions, mockOldInclude)
	if err != nil {
		t.Fatal(err)
	}

	if assert.NotNil(t, terragruntConfig.Terraform) {
		assert.Equal(t, "foo", terragruntConfig.Terraform.Source)
	}

	if assert.NotNil(t, terragruntConfig.RemoteState) {
		assert.Equal(t, "s3", terragruntConfig.RemoteState.Backend)
		assert.NotEmpty(t, terragruntConfig.RemoteState.Config)
		assert.Equal(t, true, terragruntConfig.RemoteState.Config["encrypt"])
		assert.Equal(t, "my-bucket", terragruntConfig.RemoteState.Config["bucket"])
		assert.Equal(t, "terraform.tfstate", terragruntConfig.RemoteState.Config["key"])
		assert.Equal(t, "us-east-1", terragruntConfig.RemoteState.Config["region"])
	}

	if assert.NotNil(t, terragruntConfig.Dependencies) {
		assert.Equal(t, []string{"../vpc", "../mysql", "../backend-app"}, terragruntConfig.Dependencies.Paths)
	}
}

func TestParseTerragruntConfigInclude(t *testing.T) {
	t.Parallel()

	config :=
		fmt.Sprintf(`
terragrunt = {
  include {
    path = "../../../%s"
  }
}
`, DefaultTerragruntConfigPath)

	opts := options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/sub-sub-child/" + DefaultTerragruntConfigPath)
	terragruntConfig, err := parseConfigString(config, opts, IncludeConfig{Path: opts.TerragruntConfigPath})
	if assert.Nil(t, err, "Unexpected error: %v", errors.PrintErrorWithStackTrace(err)) {
		assert.Nil(t, terragruntConfig.Terraform)

		if assert.NotNil(t, terragruntConfig.RemoteState) {
			assert.Equal(t, "s3", terragruntConfig.RemoteState.Backend)
			assert.NotEmpty(t, terragruntConfig.RemoteState.Config)
			assert.Equal(t, true, terragruntConfig.RemoteState.Config["encrypt"])
			assert.Equal(t, "my-bucket", terragruntConfig.RemoteState.Config["bucket"])
			assert.Equal(t, "child/sub-child/sub-sub-child/terraform.tfstate", terragruntConfig.RemoteState.Config["key"])
			assert.Equal(t, "us-east-1", terragruntConfig.RemoteState.Config["region"])
		}
	}

}

func TestParseTerragruntConfigIncludeWithFindInParentFolders(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  include {
    path = "${find_in_parent_folders()}"
  }
}
`

	opts := options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/sub-sub-child/" + DefaultTerragruntConfigPath)
	terragruntConfig, err := parseConfigString(config, opts, IncludeConfig{Path: opts.TerragruntConfigPath})
	if assert.Nil(t, err, "Unexpected error: %v", errors.PrintErrorWithStackTrace(err)) {
		assert.Nil(t, terragruntConfig.Terraform)

		if assert.NotNil(t, terragruntConfig.RemoteState) {
			assert.Equal(t, "s3", terragruntConfig.RemoteState.Backend)
			assert.NotEmpty(t, terragruntConfig.RemoteState.Config)
			assert.Equal(t, true, terragruntConfig.RemoteState.Config["encrypt"])
			assert.Equal(t, "my-bucket", terragruntConfig.RemoteState.Config["bucket"])
			assert.Equal(t, "child/sub-child/sub-sub-child/terraform.tfstate", terragruntConfig.RemoteState.Config["key"])
			assert.Equal(t, "us-east-1", terragruntConfig.RemoteState.Config["region"])
		}
	}

}

func TestParseTerragruntConfigIncludeOverrideRemote(t *testing.T) {
	t.Parallel()

	config :=
		fmt.Sprintf(`
terragrunt = {
  include {
    path = "../../../%s"
  }

  # Configure Terragrunt to automatically store tfstate files in an S3 bucket
  remote_state {
    backend = "s3"
    config {
      encrypt = false
      bucket = "override"
      key = "override"
      region = "override"
    }
  }
}
`, DefaultTerragruntConfigPath)

	opts := options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/sub-sub-child/" + DefaultTerragruntConfigPath)
	terragruntConfig, err := parseConfigString(config, opts, IncludeConfig{Path: opts.TerragruntConfigPath})
	if assert.Nil(t, err, "Unexpected error: %v", errors.PrintErrorWithStackTrace(err)) {
		assert.Nil(t, terragruntConfig.Terraform)

		if assert.NotNil(t, terragruntConfig.RemoteState) {
			assert.Equal(t, "s3", terragruntConfig.RemoteState.Backend)
			assert.NotEmpty(t, terragruntConfig.RemoteState.Config)
			assert.Equal(t, false, terragruntConfig.RemoteState.Config["encrypt"])
			assert.Equal(t, "override", terragruntConfig.RemoteState.Config["bucket"])
			assert.Equal(t, "override", terragruntConfig.RemoteState.Config["key"])
			assert.Equal(t, "override", terragruntConfig.RemoteState.Config["region"])
		}
	}

}

func TestParseTerragruntConfigIncludeOverrideAll(t *testing.T) {
	t.Parallel()

	config :=
		fmt.Sprintf(`
terragrunt = {
  include {
    path = "../../../%s"
  }

  terraform {
    source = "foo"
  }

  # Configure Terragrunt to automatically store tfstate files in an S3 bucket
  remote_state {
    backend = "s3"
    config {
      encrypt = false
      bucket = "override"
      key = "override"
      region = "override"
    }
  }

  dependencies {
    paths = ["override"]
  }
}
`, DefaultTerragruntConfigPath)

	opts := options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/sub-sub-child/" + DefaultTerragruntConfigPath)
	terragruntConfig, err := parseConfigString(config, opts, IncludeConfig{Path: opts.TerragruntConfigPath})
	if assert.Nil(t, err, "Unexpected error: %v", errors.PrintErrorWithStackTrace(err)) {
		if assert.NotNil(t, terragruntConfig.Terraform) {
			assert.Equal(t, "foo", terragruntConfig.Terraform.Source)
		}

		if assert.NotNil(t, terragruntConfig.RemoteState) {
			assert.Equal(t, "s3", terragruntConfig.RemoteState.Backend)
			assert.NotEmpty(t, terragruntConfig.RemoteState.Config)
			assert.Equal(t, false, terragruntConfig.RemoteState.Config["encrypt"])
			assert.Equal(t, "override", terragruntConfig.RemoteState.Config["bucket"])
			assert.Equal(t, "override", terragruntConfig.RemoteState.Config["key"])
			assert.Equal(t, "override", terragruntConfig.RemoteState.Config["region"])
		}

		assert.Equal(t, []string{"override"}, terragruntConfig.Dependencies.Paths)
	}
}

func TestParseTerragruntConfigTwoLevels(t *testing.T) {
	t.Parallel()

	configPath := "../test/fixture-parent-folders/multiple-terragrunt-in-parents/child/sub-child/" + DefaultTerragruntConfigPath

	config, err := util.ReadFileAsString(configPath)
	if err != nil {
		t.Fatal(err)
	}

	terragruntConfig, err := parseConfigString(config, options.NewTerragruntOptionsForTest(configPath), IncludeConfig{Path: configPath})
	assert.Nil(t, err)
	assert.NotNil(t, terragruntConfig)
}

func TestParseTerragruntConfigThreeLevels(t *testing.T) {
	t.Parallel()

	configPath := "../test/fixture-parent-folders/multiple-terragrunt-in-parents/child/sub-child/sub-sub-child/" + DefaultTerragruntConfigPath

	config, err := util.ReadFileAsString(configPath)
	if err != nil {
		t.Fatal(err)
	}

	terragruntConfig, err := parseConfigString(config, options.NewTerragruntOptionsForTest(configPath), IncludeConfig{Path: configPath})
	assert.Nil(t, err)
	assert.NotNil(t, terragruntConfig)
}

func TestParseWithBootStrapFile(t *testing.T) {
	// Cannot be run in parallel since it defines an environment variable
	fixture := "../test/fixture-bootstrap/"
	absolute, _ := filepath.Abs(fixture)
	os.Setenv(options.EnvBootConfigs, fmt.Sprintf("%[1]s/a.tfvars:%[1]s/b.tfvars", absolute))
	defer func() { os.Setenv(options.EnvBootConfigs, "") }()

	configPath := fixture + DefaultTerragruntConfigPath

	config, err := util.ReadFileAsString(configPath)
	if err != nil {
		t.Fatal(err)
	}

	terragruntConfig, err := parseConfigString(config, options.NewTerragruntOptionsForTest(configPath), IncludeConfig{Path: configPath})
	assert.Nil(t, err)
	assert.NotNil(t, terragruntConfig)
	assert.Equal(t, len(terragruntConfig.PreHooks), 2, "Should have 2 pre_hook(s)")
}

func TestParseWithNoFile(t *testing.T) {
	t.Parallel()
	config, err := ReadTerragruntConfig(options.NewTerragruntOptionsForTest("../test/fixture-noconfig/no-file/" + DefaultTerragruntConfigPath))
	assert.Nil(t, err)
	assert.NotNil(t, config)
}

func TestParseWithNoConfig(t *testing.T) {
	t.Parallel()
	config, err := ReadTerragruntConfig(options.NewTerragruntOptionsForTest("../test/fixture-noconfig/no-terragrunt/" + DefaultTerragruntConfigPath))
	assert.Nil(t, err)
	assert.NotNil(t, config)
}

func TestParseWithBadPath(t *testing.T) {
	t.Parallel()
	config, err := ReadTerragruntConfig(options.NewTerragruntOptionsForTest("../test/fixture-noconfig/bad-path/" + DefaultTerragruntConfigPath))
	assert.NotNil(t, err)
	assert.Nil(t, config)
}

func TestParseValid(t *testing.T) {
	t.Parallel()
	config, err := ReadTerragruntConfig(options.NewTerragruntOptionsForTest("../test/fixture-noconfig/valid/" + DefaultTerragruntConfigPath))
	assert.Nil(t, err)
	assert.NotNil(t, config)
}

func TestParseInvalid(t *testing.T) {
	t.Parallel()
	config, err := ReadTerragruntConfig(options.NewTerragruntOptionsForTest("../test/fixture-noconfig/invalid/" + DefaultTerragruntConfigPath + ".invalid"))
	assert.NotNil(t, err)
	assert.Nil(t, config)
}

func TestParseTerragruntConfigEmptyConfig(t *testing.T) {
	t.Parallel()

	config := ``

	_, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	assert.True(t, errors.IsError(err, CouldNotResolveTerragruntConfigInFile(DefaultTerragruntConfigPath)))
}

func TestParseTerragruntConfigEmptyConfigOldConfig(t *testing.T) {
	t.Parallel()

	config := ``

	terragruntConfig, err := parseConfigString(config, mockOptions, mockOldInclude)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, terragruntConfig.RemoteState)
}

type argConfig struct {
	name      string
	extraArgs []string
}

func getExtraArgsConfig(options *options.TerragruntOptions, argConfigs ...argConfig) TerragruntConfig {
	args := []TerraformExtraArguments{}
	for _, argConfig := range argConfigs {
		base := TerragruntExtensionBase{Name: argConfig.name}
		base.init(&TerragruntConfigFile{})
		base.config().options = options
		args = append(args, TerraformExtraArguments{TerragruntExtensionBase: base, Arguments: argConfig.extraArgs})
	}
	return TerragruntConfig{Terraform: &TerraformConfig{ExtraArgs: args}}
}

func TestMergeConfigIntoIncludedConfig(t *testing.T) {
	t.Parallel()

	options := options.NewTerragruntOptionsForTest("TestMergeConfigIntoIncludedConfig")

	testCases := []struct {
		config         TerragruntConfig
		includedConfig TerragruntConfig
		expected       TerragruntConfig
	}{
		{
			TerragruntConfig{},
			TerragruntConfig{},
			TerragruntConfig{},
		},
		{
			TerragruntConfig{},
			TerragruntConfig{Terraform: &TerraformConfig{Source: "foo"}},
			TerragruntConfig{Terraform: &TerraformConfig{Source: "foo"}},
		},
		{
			TerragruntConfig{},
			TerragruntConfig{RemoteState: &remote.RemoteState{Backend: "bar"}, Terraform: &TerraformConfig{Source: "foo"}},
			TerragruntConfig{RemoteState: &remote.RemoteState{Backend: "bar"}, Terraform: &TerraformConfig{Source: "foo"}},
		},
		{
			TerragruntConfig{RemoteState: &remote.RemoteState{Backend: "foo"}, Terraform: &TerraformConfig{Source: "foo"}},
			TerragruntConfig{RemoteState: &remote.RemoteState{Backend: "bar"}, Terraform: &TerraformConfig{Source: "bar"}},
			TerragruntConfig{RemoteState: &remote.RemoteState{Backend: "foo"}, Terraform: &TerraformConfig{Source: "foo"}},
		},
		{
			TerragruntConfig{Terraform: &TerraformConfig{Source: "foo"}},
			TerragruntConfig{RemoteState: &remote.RemoteState{Backend: "bar"}, Terraform: &TerraformConfig{Source: "bar"}},
			TerragruntConfig{RemoteState: &remote.RemoteState{Backend: "bar"}, Terraform: &TerraformConfig{Source: "foo"}},
		},
		{
			getExtraArgsConfig(options, argConfig{name: "childArgs"}),
			getExtraArgsConfig(options),
			getExtraArgsConfig(options, argConfig{name: "childArgs"}),
		},
		{
			getExtraArgsConfig(options, argConfig{name: "childArgs"}),
			getExtraArgsConfig(options, argConfig{name: "parentArgs"}),
			getExtraArgsConfig(options, argConfig{name: "parentArgs"}, argConfig{name: "childArgs"}),
		},
		{
			getExtraArgsConfig(options, argConfig{name: "overrideArgs", extraArgs: []string{"-child"}}),
			getExtraArgsConfig(options, argConfig{name: "overrideArgs", extraArgs: []string{"-parent"}}),
			getExtraArgsConfig(options, argConfig{name: "overrideArgs", extraArgs: []string{"-child"}}),
		},
	}

	for _, testCase := range testCases {
		testCase.config.mergeIncludedConfig(testCase.includedConfig, mockOptions)
		assert.Equal(t, testCase.config, testCase.expected, "For config %v and includeConfig %v", testCase.config, testCase.includedConfig)
	}
}

func TestParseTerragruntConfigTerraformNoSource(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  terraform {
  }
}
`

	terragruntConfig, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, terragruntConfig.RemoteState)
	assert.Nil(t, terragruntConfig.Dependencies)

	if assert.NotNil(t, terragruntConfig.Terraform) {
		assert.Empty(t, terragruntConfig.Terraform.Source)
	}
}

func TestParseTerragruntConfigTerraformWithSource(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  terraform {
    source = "foo"
  }
}
`

	terragruntConfig, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, terragruntConfig.RemoteState)
	assert.Nil(t, terragruntConfig.Dependencies)

	if assert.NotNil(t, terragruntConfig.Terraform) {
		assert.Equal(t, "foo", terragruntConfig.Terraform.Source)
	}
}

func TestParseTerragruntConfigTerraformWithExtraArguments(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  terraform {
    extra_arguments "secrets" {
      arguments = [
        "-var-file=terraform.tfvars",
        "-var-file=terraform-secret.tfvars"
      ]
      commands = ["${get_terraform_commands_that_need_vars()}"]
    }
  }
}
`

	terragruntConfig, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, terragruntConfig.RemoteState)
	assert.Nil(t, terragruntConfig.Dependencies)

	if assert.NotNil(t, terragruntConfig.Terraform) {
		assert.Equal(t, "secrets", terragruntConfig.Terraform.ExtraArgs[0].Name)
		assert.Equal(t,
			[]string{
				"-var-file=terraform.tfvars",
				"-var-file=terraform-secret.tfvars",
			},
			terragruntConfig.Terraform.ExtraArgs[0].Arguments)
		assert.Equal(t,
			TerraformCommandWithVarFile,
			terragruntConfig.Terraform.ExtraArgs[0].Commands)
	}
}

func TestParseTerragruntConfigTerraformWithMultipleExtraArguments(t *testing.T) {
	t.Parallel()

	config := `
terragrunt = {
  terraform {
    extra_arguments "json_output" {
      arguments = [
        "-json"
      ]
      commands = [
        "output"
      ]
    }

    extra_arguments "fmt_diff" {
      arguments = [
        "-diff=true"
      ]
      commands = [
        "fmt"
      ]
    }

    extra_arguments "required_tfvars" {
      required_var_files = [
        "file1.tfvars",
				"file2.tfvars"
      ]
      commands = ["${get_terraform_commands_that_need_vars()}"]
    }

    extra_arguments "optional_tfvars" {
      optional_var_files = [
        "opt1.tfvars",
				"opt2.tfvars"
      ]
      commands = ["${get_terraform_commands_that_need_vars()}"]
    }
  }
}
`

	terragruntConfig, err := parseConfigString(config, mockOptions, mockDefaultInclude)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, terragruntConfig.RemoteState)
	assert.Nil(t, terragruntConfig.Dependencies)

	if assert.NotNil(t, terragruntConfig.Terraform) {
		assert.Equal(t, "json_output", terragruntConfig.Terraform.ExtraArgs[0].Name)
		assert.Equal(t, []string{"-json"}, terragruntConfig.Terraform.ExtraArgs[0].Arguments)
		assert.Equal(t, []string{"output"}, terragruntConfig.Terraform.ExtraArgs[0].Commands)
		assert.Equal(t, "fmt_diff", terragruntConfig.Terraform.ExtraArgs[1].Name)
		assert.Equal(t, []string{"-diff=true"}, terragruntConfig.Terraform.ExtraArgs[1].Arguments)
		assert.Equal(t, []string{"fmt"}, terragruntConfig.Terraform.ExtraArgs[1].Commands)
		assert.Equal(t, "required_tfvars", terragruntConfig.Terraform.ExtraArgs[2].Name)
		assert.Equal(t, []string{"file1.tfvars", "file2.tfvars"}, terragruntConfig.Terraform.ExtraArgs[2].RequiredVarFiles)
		assert.Equal(t, TerraformCommandWithVarFile, terragruntConfig.Terraform.ExtraArgs[2].Commands)
		assert.Equal(t, "optional_tfvars", terragruntConfig.Terraform.ExtraArgs[3].Name)
		assert.Equal(t, []string{"opt1.tfvars", "opt2.tfvars"}, terragruntConfig.Terraform.ExtraArgs[3].OptionalVarFiles)
		assert.Equal(t, TerraformCommandWithVarFile, terragruntConfig.Terraform.ExtraArgs[3].Commands)
	}
}

func TestFindConfigFilesInPathNone(t *testing.T) {
	t.Parallel()

	expected := []string{}
	actual, err := FindConfigFilesInPath(newOptionsWorkingDir("../test/fixture-config-files/none"))

	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, expected, actual)
}

func TestFindConfigFilesInPathOneNewConfig(t *testing.T) {
	t.Parallel()

	expected := []string{"../test/fixture-config-files/one-new-config/subdir/terraform.tfvars"}
	actual, err := FindConfigFilesInPath(newOptionsWorkingDir("../test/fixture-config-files/one-new-config"))

	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, expected, actual)
}

func TestFindConfigFilesInPathOneOldConfig(t *testing.T) {
	t.Parallel()

	expected := []string{"../test/fixture-config-files/one-old-config/subdir/.terragrunt"}
	actual, err := FindConfigFilesInPath(newOptionsWorkingDir("../test/fixture-config-files/one-old-config"))

	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, expected, actual)
}

func TestFindConfigFilesInPathMultipleConfigs(t *testing.T) {
	t.Parallel()

	expected := []string{
		"../test/fixture-config-files/multiple-configs/terraform.tfvars",
		"../test/fixture-config-files/multiple-configs/subdir-2/subdir/.terragrunt",
		"../test/fixture-config-files/multiple-configs/subdir-3/terraform.tfvars",
	}
	actual, err := FindConfigFilesInPath(newOptionsWorkingDir("../test/fixture-config-files/multiple-configs"))

	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, expected, actual)
}

func newOptionsWorkingDir(workingDir string) *options.TerragruntOptions {
	opts := options.NewTerragruntOptionsForTest("")
	opts.WorkingDir = workingDir
	return opts
}

func newOptionsVariables(variables map[string]interface{}) *options.TerragruntOptions {
	opts := options.NewTerragruntOptionsForTest("")
	newMap := make(map[string]options.Variable, len(variables))
	for key, value := range variables {
		newMap[key] = options.Variable{Value: value}
	}
	opts.Variables = newMap
	return opts
}
