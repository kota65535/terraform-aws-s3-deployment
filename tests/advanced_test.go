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
	files := map[string]S3Object{
		"a.json": {
			Metadata: map[string]string{
				"Content-Type":     "application/json",
				"Content-Language": "en-US",
			},
			ETag: "\"d5524a5b020a0553b930cc3f2f8e4cce\"",
		},
		"b.json": {
			Metadata: map[string]string{
				"Content-Type":        "binary/octet-stream",
				"Cache-Control":       "public, max-age=31536000, immutable",
				"Content-Disposition": "inline",
				"Content-Encoding":    "compress",
				"Content-Language":    "ja-JP",
			},
			ETag: "\"64cd52391a8a2843d7cb347e872720b0\"",
		},
		"config-09e8d29e.js": {
			Metadata: map[string]string{
				"Content-Type":  "text/javascript",
				"Cache-Control": "public, max-age=0, must-revalidate",
			},
			ETag: "\"5f56ab0a8e07afb6ef5885a8486927ab\"",
		},
		"index.html": {
			Metadata: map[string]string{
				"Content-Type":  "text/html",
				"Cache-Control": "public, max-age=0, must-revalidate",
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
				"Content-Type":  "text/javascript",
				"Cache-Control": "public, max-age=0, must-revalidate",
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
