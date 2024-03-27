package tests

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/avast/retry-go/v4"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/googleapi"
	"log"
	"os"
	"sort"
	"sync"
	"testing"
	"time"
)

func readJson(t *testing.T, path string) map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	bytes, err := os.ReadFile(path)
	require.NoError(t, err)
	err = json.Unmarshal(bytes, &ret)
	require.NoError(t, err)
	return ret
}

func emptyBucket(svc *s3.Client, bucket string) {
	ctx := context.TODO()
	result, err := svc.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Fatalf("Couldn't list objects in bucket, %v", err)
	}
	var objects []types.ObjectIdentifier
	for _, item := range result.Contents {
		objects = append(objects, types.ObjectIdentifier{Key: item.Key})
	}
	if len(objects) == 0 {
		return
	}
	_, err = svc.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: objects,
		},
	})
	if err != nil {
		log.Fatalf("Couldn't delete objects in bucket, %v", err)
	}
}

func assertOutputs(t *testing.T, terraformOptions *terraform.Options, expected map[string]interface{}) {
	actual := make(map[string]interface{})
	terraform.OutputStruct(t, terraformOptions, "s3_objects", &actual)
	assert.Equal(t, expected, actual)

}

func assertObjects(t *testing.T, svc *s3.Client, bucket string, files map[string]map[string]string) {
	ctx := context.TODO()

	// Assert objects exist
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
	isOK, _ := doRetry(func() (bool, error) {
		return assert.ObjectsAreEqual(expectedKeys, actualKeys), nil
	})
	if !isOK {
		assert.Equal(t, expectedKeys, actualKeys)
	}

	// Assert object metadata
	var wg sync.WaitGroup
	var mu sync.Mutex
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
			mu.Lock()
			defer mu.Unlock()
			metadata[*item.Key] = result
		}()
	}
	wg.Wait()

	for k, v := range metadata {
		matched := 0
		if v.ContentType != nil {
			// TODO: cross-platform MIME type check
			// assert.Equal(t, files[k]["Content-Type"], *v.ContentType)
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

func doRetry[T any](fn retry.RetryableFuncWithData[T]) (T, error) {
	return retry.DoWithData(fn,
		retry.OnRetry(func(n uint, err error) {
			log.Printf("(#%d/3) Retrying for eventual consistentency...\n", n+1)
		}),
		retry.RetryIf(func(err error) bool {
			var x *googleapi.Error
			if errors.As(err, &x) {
				return x.Code == 429
			}
			return false
		}),
		retry.Delay(1*time.Second),
		retry.Attempts(3),
	)
}
