package tests

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"log"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func TestAdvanced(t *testing.T) {
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

	// Arrange
	terraformOptions := &terraform.Options{
		TerraformDir: "../examples/advanced",
	}

	// Act
	terraform.InitAndApply(t, terraformOptions)

	// Assert
	expected := readJson(t, "../tests/advanced_expected.json")
	actual := make(map[string]interface{})
	terraform.OutputStruct(t, terraformOptions, "s3_objects", &actual)
	assert.True(t, reflect.DeepEqual(expected, actual))

	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	svc := s3.NewFromConfig(cfg)
	result, err := svc.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Fatalf("Couldn't list objects in bucket, %v", err)
	}

	var actualKeys []string
	for _, item := range result.Contents {
		actualKeys = append(actualKeys, *item.Key)
	}
	sort.Strings(actualKeys)
	var expectedKeys []string
	for k, _ := range files {
		expectedKeys = append(expectedKeys, k)
	}
	sort.Strings(expectedKeys)
	assert.Equal(t, expectedKeys, actualKeys)

	var wg sync.WaitGroup
	metadata := make(map[string]*s3.HeadObjectOutput, 0)
	for _, item := range result.Contents {
		item := item
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := svc.HeadObject(ctx, &s3.HeadObjectInput{
				Bucket: aws.String(bucket),
				Key:    item.Key,
			})
			if err != nil {
				log.Fatalf("Couldn't head object in bucket, %v", err)
			}
			metadata[*item.Key] = result
		}()
	}
	wg.Wait()

	for k, v := range metadata {
		matched := 0
		if v.ContentType != nil {
			assert.Equal(t, files[k]["Content-Type"], *v.ContentType)
			matched++
		}
		if v.CacheControl != nil {
			assert.Equal(t, files[k]["Cache-Control"], *v.CacheControl)
			matched++
		}
		if v.ContentDisposition != nil {
			assert.Equal(t, files[k]["Content-Disposition"], *v.ContentDisposition)
			matched++
		}
		if v.ContentEncoding != nil {
			assert.Equal(t, files[k]["Content-Encoding"], *v.ContentEncoding)
			matched++
		}
		if v.ContentLanguage != nil {
			assert.Equal(t, files[k]["Content-Language"], *v.ContentLanguage)
			matched++
		}
		assert.Equal(t, len(files[k]), matched)
	}

}
