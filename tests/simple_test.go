package tests

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"log"
	"testing"
)

func TestSimple(t *testing.T) {
	// Arrange
	bucket := "s3-deployment-simple-561678142736"
	region := "ap-northeast-1"
	files := map[string]map[string]string{
		"a.json": {
			"Content-Type": "application/json",
		},
		"b.json": {
			"Content-Type": "application/json",
		},
		"config-09e8d29e.js": {
			"Content-Type": "application/javascript",
		},
		"index.html": {
			"Content-Type": "text/html",
		},
		"octocat.png": {
			"Content-Type": "image/png",
		},
		"script.js": {
			"Content-Type": "application/javascript",
		},
		"style.css": {
			"Content-Type": "text/css",
		},
	}

	// S3 client
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	svc := s3.NewFromConfig(cfg)

	emptyBucket(svc, bucket)

	terraformOptions := &terraform.Options{
		TerraformDir: "../examples/simple",
		Upgrade:      true,
	}

	// Act
	terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertOutputs(t, terraformOptions, map[string]interface{}{})
	assertObjects(t, svc, bucket, files)

	// Add an object
	_, err = svc.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String("delete me"),
		CopySource: aws.String(bucket + "/a.json"),
	})
	if err != nil {
		log.Fatalf("cannot add an object, %v", err)
	}

	// Act
	terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertOutputs(t, terraformOptions, map[string]interface{}{})
	assertObjects(t, svc, bucket, files)

	// Delete an object
	_, err = svc.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("a.json"),
	})
	if err != nil {
		log.Fatalf("cannot delete an object, %v", err)
	}

	// Act
	terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertOutputs(t, terraformOptions, map[string]interface{}{})
	assertObjects(t, svc, bucket, files)
}
