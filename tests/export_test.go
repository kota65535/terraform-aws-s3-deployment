package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"os"
	"sort"
	"strings"
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

func assertResult(t *testing.T, out string, added int, changed int, destroyed int) {
	assert.True(t, strings.Contains(out, fmt.Sprintf("Apply complete! Resources: %d added, %d changed, %d destroyed.", added, changed, destroyed)))
}

type S3Object struct {
	Metadata map[string]string
	Content  string
}

func assertObjects(t *testing.T, svc *s3.Client, bucket string, files map[string]*S3Object) {
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
	for k := range files {
		expectedKeys = append(expectedKeys, k)
	}
	sort.Strings(expectedKeys)
	isOK, _ := doRetry(func() (bool, error) {
		if !assert.ObjectsAreEqual(expectedKeys, actualKeys) {
			return false, fmt.Errorf("assertion failed")
		}
		return true, nil
	})
	if !isOK {
		assert.Equal(t, expectedKeys, actualKeys)
	}

	// Get objects
	var wg sync.WaitGroup
	var mu sync.Mutex
	objects := make(map[string]*s3.GetObjectOutput)
	for _, item := range result.Contents {
		item := item
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := svc.GetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    item.Key,
			})
			if err != nil {
				log.Fatalf("Couldn't get object in bucket, %v", err)
			}
			mu.Lock()
			defer mu.Unlock()
			objects[*item.Key] = result
		}()
	}
	wg.Wait()

	// Assert object metadata
	for k, v := range objects {
		matched := 0
		if v.ContentType != nil {
			// TODO: cross-platform MIME type check
			// assert.Equal(t, files[k]["Content-Type"], *v.ContentType)
			matched++
		}
		if v.CacheControl != nil {
			assert.Equal(t, files[k].Metadata["Cache-Control"], *v.CacheControl)
			matched++
		}
		if v.ContentDisposition != nil {
			assert.Equal(t, files[k].Metadata["Content-Disposition"], *v.ContentDisposition)
			matched++
		}
		if v.ContentEncoding != nil {
			assert.Equal(t, files[k].Metadata["Content-Encoding"], *v.ContentEncoding)
			matched++
		}
		if v.ContentLanguage != nil {
			assert.Equal(t, files[k].Metadata["Content-Language"], *v.ContentLanguage)
			matched++
		}
		assert.Equal(t, len(files[k].Metadata), matched)
	}

	// Assert object contents
	for k, v := range objects {
		defer v.Body.Close()
		if files[k].Content == "" {
			continue
		}
		body, err := io.ReadAll(v.Body)
		if err != nil {
			log.Fatalf("Couldn't read object body, %v", err)
		}
		assert.Equal(t, files[k].Content, string(body))
	}
}

func doRetry[T any](fn retry.RetryableFuncWithData[T]) (T, error) {
	return retry.DoWithData(fn,
		retry.OnRetry(func(n uint, err error) {
			log.Printf("(#%d/3) Retrying for eventual consistentency...\n", n+1)
		}),
		retry.Delay(3*time.Second),
		retry.Attempts(3),
	)
}
