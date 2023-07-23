package tests

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"reflect"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestModule(t *testing.T) {
	// Arrange
	terraformOptions := &terraform.Options{
		TerraformDir: "../examples",
	}

	// Act
	terraform.InitAndApply(t, terraformOptions)

	// Assert
	expected := readJson(t, "../tests/expected.json")
	actual := make(map[string]interface{}, 0)
	terraform.OutputStruct(t, terraformOptions, "s3_objects", &actual)
	assert.True(t, reflect.DeepEqual(expected, actual))
}

func readJson(t *testing.T, path string) map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	bytes, err := os.ReadFile(path)
	require.NoError(t, err)
	err = json.Unmarshal(bytes, &ret)
	require.NoError(t, err)
	return ret
}
