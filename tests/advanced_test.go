package tests

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAdvanced(t *testing.T) {
	// Arrange
	bucket := "s3-deployment-561678142736"
	region := "ap-northeast-1"
	files := map[string]*S3Object{
		"a.json": {
			Metadata: map[string]string{
				"Content-Type":     "application/json",
				"Content-Language": "en-US",
			},
			Content: "{\n  \"a\": \"1\",\n  \"b\": 2,\n  \"c\": {\n    \"d\": \"3\"\n  }\n}\n",
		},
		"b.json": {
			Metadata: map[string]string{
				"Content-Type":        "binary/octet-stream",
				"Cache-Control":       "public, max-age=31536000, immutable",
				"Content-Disposition": "inline",
				"Content-Encoding":    "compress",
				"Content-Language":    "ja-JP",
			},
			Content: "{\"a\":\"1\",\"h\":\"2\",\"i\":{\"j\":3,\"k\":\"4\"}}\n",
		},
		"config-09e8d29e.js": {
			Metadata: map[string]string{
				"Content-Type":  "text/javascript",
				"Cache-Control": "public, max-age=0, must-revalidate",
			},
			Content: "const c = JSON.parse('{\"abc\":[1,2,3],\"unicorns\":\"awesome\"}'); export default c;\n",
		},
		"index.html": {
			Metadata: map[string]string{
				"Content-Type":  "text/html",
				"Cache-Control": "public, max-age=0, must-revalidate",
			},
		},
		"octocat.png": {
			Metadata: map[string]string{
				"Content-Type": "image/png",
			},
		},
		"script.js": {
			Metadata: map[string]string{
				"Content-Type":  "text/javascript",
				"Cache-Control": "public, max-age=0, must-revalidate",
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

	// === Initialization ==

	terraformOptions := &terraform.Options{
		TerraformDir: "../examples/advanced",
		Upgrade:      true,
		LockTimeout:  "5m",
		Vars: map[string]interface{}{
			"archive_path": "test.zip",
		},
	}

	// Act
	out := terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertObjects(t, svc, bucket, files)

	// === Test added object in the bucket will be deleted ===

	// Add an object to the bucket
	_, err = svc.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String("delete me"),
		CopySource: aws.String(bucket + "/a.json"),
	})
	require.NoError(t, err, "cannot add an object")

	// Act
	out = terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertResult(t, out, 2, 0, 2)
	assertObjects(t, svc, bucket, files)

	// === Test deleted object in the bucket will be added ===

	// Delete an object from the bucket
	_, err = svc.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("a.json"),
	})
	require.NoError(t, err, "cannot delete an object")

	// Act
	out = terraform.InitAndApply(t, terraformOptions)

	// Assert
	assertResult(t, out, 2, 0, 2)
	assertObjects(t, svc, bucket, files)

	// === Test deploying another archive ===

	terraformOptions.Vars = map[string]interface{}{
		"archive_path": "test2.zip",
	}

	// Act
	out = terraform.InitAndApply(t, terraformOptions)

	files["a.json"].Content = "{\n  \"a\": \"9\",\n  \"b\": 2,\n  \"c\": {\n    \"d\": \"3\"\n  }\n}\n"

	// Assert
	assertResult(t, out, 2, 0, 2)
	assertObjects(t, svc, bucket, files)
}
