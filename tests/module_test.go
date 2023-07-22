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

	expected := make(map[string]interface{}, 0)
	bytes, err := os.ReadFile("../tests/expected.json")
	require.NoError(t, err)
	err = json.Unmarshal(bytes, &expected)
	require.NoError(t, err)

	actual := terraform.OutputMapOfObjects(t, terraformOptions, "s3_objects")
	assert.True(t, reflect.DeepEqual(expected, actual))
}
