package tests

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"log"
	"strings"
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
			ETag: "\"d5524a5b020a0553b930cc3f2f8e4cce\"",
		},
		"b.json": {
			Metadata: map[string]string{
				"Content-Type": "application/json",
			},
			ETag: "\"8f9c289b2cb8faa7199cde10bd185b99\"",
		},
		"config-09e8d29e.js": {
			Metadata: map[string]string{
				"Content-Type": "application/javascript",
			},
			ETag: "\"96d43a1e51087a6a225bb56885a63553\"",
		},
		"index.html": {
			Metadata: map[string]string{
				"Content-Type": "text/html",
			},
			ETag: "\"faeee5e2efb928e33b0eb232a1ce1f85\"",
		},
		"octocat.png": {
			Metadata: map[string]string{
				"Content-Type": "image/png",
			},
			ETag: "\"f1d23c21191e970573e34ceb555c332b\"",
		},
		"script.js": {
			Metadata: map[string]string{
				"Content-Type": "application/javascript",
			},
			ETag: "\"847706ee8f66d4a9b30302446279ea5e\"",
		},
		"style.css": {
			Metadata: map[string]string{
				"Content-Type": "text/css",
			},
			ETag: "\"ed143c36d3a6cb0fec57de6d828bb52a\"",
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

	// Update an object
	_, err = svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("b.json"),
		Body:   strings.NewReader("{\n  \"a\": \"9\"\n}\n"),
	})
	if err != nil {
		log.Fatalf("cannot update an object, %v", err)
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
