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
	files := map[string]S3Object{
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
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	svc := s3.NewFromConfig(cfg)

	emptyBucket(svc, bucket)

	terraformOptions := &terraform.Options{
		TerraformDir: "../examples/simple",
		Upgrade:      true,
		LockTimeout:  "5m",
	}

	// Act
	out := terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertResult(t, out, 1, 0, 1)
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
	out = terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertResult(t, out, 1, 0, 1)
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
	out = terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertResult(t, out, 1, 0, 1)
	assertObjects(t, svc, bucket, files)
}
