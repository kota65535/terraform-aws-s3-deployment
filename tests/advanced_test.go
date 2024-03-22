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

func TestAdvanced(t *testing.T) {
	// Arrange
	bucket := "s3-deployment-561678142736"
	region := "ap-northeast-1"
	files := map[string]map[string]string{
		"a.json": {
			"Content-Type":     "application/json",
			"Content-Language": "en-US",
		},
		"b.json": {
			"Content-Type":        "binary/octet-stream",
			"Cache-Control":       "public, max-age=31536000, immutable",
			"Content-Disposition": "inline",
			"Content-Encoding":    "compress",
			"Content-Language":    "ja-JP",
		},
		"config-09e8d29e.js": {
			"Content-Type":  "text/javascript",
			"Cache-Control": "public, max-age=0, must-revalidate",
		},
		"index.html": {
			"Content-Type":  "text/html",
			"Cache-Control": "public, max-age=0, must-revalidate",
		},
		"octocat.png": {
			"Content-Type": "image/png",
		},
		"script.js": {
			"Content-Type":  "text/javascript",
			"Cache-Control": "public, max-age=0, must-revalidate",
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
		TerraformDir: "../examples/advanced",
		Upgrade:      true,
		LockTimeout:  "5m",
	}

	// Act
	terraform.InitAndApply(t, terraformOptions)

	// Assert
	expected := readJson(t, "../tests/advanced_expected.json")
	assertOutputs(t, terraformOptions, expected)
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
	assertOutputs(t, terraformOptions, expected)
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
	assertOutputs(t, terraformOptions, expected)
	assertObjects(t, svc, bucket, files)
}
