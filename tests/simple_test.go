package tests

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimple(t *testing.T) {
	// Arrange
	bucket := "s3-deployment-simple-561678142736"
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
		Reconfigure:  true,
	}

	// Act
	terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertObjects(t, svc, bucket, files)
}
