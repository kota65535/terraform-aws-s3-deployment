package tests

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestSimple(t *testing.T) {
	platform, _ := os.LookupEnv("TF_PLATFORM")
	if platform == "" {
		platform = "unknown"
	}
	version, _ := os.LookupEnv("TF_VERSION")
	if version == "" {
		version = "unknown"
	}

	bucket := fmt.Sprintf("s3-deployment-561678142736-simple-%s-%s", platform, version)
	backendKey := fmt.Sprintf("terraform-s3-deployment-561678142736-simple-%s-%s", platform, version)

	// Arrange
	region := "ap-northeast-1"
	files := map[string]*S3Object{
		"a.json": {
			Metadata: map[string]string{
				"Content-Type": "application/json",
			},
		},
		"b.json": {
			Metadata: map[string]string{
				"Content-Type": "application/json",
			},
		},
		"config-09e8d29e.js": {
			Metadata: map[string]string{
				"Content-Type": "application/javascript",
			},
		},
		"index.html": {
			Metadata: map[string]string{
				"Content-Type": "text/html",
			},
		},
		"octocat.png": {
			Metadata: map[string]string{
				"Content-Type": "image/png",
			},
		},
		"script.js": {
			Metadata: map[string]string{
				"Content-Type": "application/javascript",
			},
		},
		"style.css": {
			Metadata: map[string]string{
				"Content-Type": "text/css",
			},
		},
	}

	// S3 client
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	require.NoError(t, err, "unable to load SDK config")

	svc := s3.NewFromConfig(cfg)

	terraformOptions := &terraform.Options{
		TerraformDir: "../examples/simple",
		Upgrade:      true,
		LockTimeout:  "5m",
		Vars: map[string]interface{}{
			"bucket": bucket,
		},
		BackendConfig: map[string]interface{}{
			"key": backendKey,
		},
		Reconfigure: true,
	}

	// Act
	terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertObjects(t, svc, bucket, files)
}
