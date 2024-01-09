package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Impl struct {
	client *s3.Client
	bucket string
}

func InitializeS3Service(bucketName string) Fetcher {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-1"))
	if err != nil {
		panic(fmt.Errorf("Unable to initialize S3: %s", err))
	}
	return &S3Impl{s3.NewFromConfig(cfg), bucketName}
}

func (service *S3Impl) ListFiles() []string {
	log.Printf("Fetching files in bucket %s\n", service.bucket)
	listRequest := &s3.ListObjectsV2Input{
		Bucket: aws.String(service.bucket),
	}
	response := service.getListResponse(listRequest)
	keys := make([]string, 0, len(response.Contents))
	for _, obj := range response.Contents {
		keys = append(keys, aws.ToString(obj.Key))
	}
	// It is possible that we weren't able to fetch all the files
	// in the first request so we paginate if the result was
	// truncated.
	for *response.IsTruncated {
		listRequest.ContinuationToken = response.NextContinuationToken
		response = service.getListResponse(listRequest)
		for _, obj := range response.Contents {
			keys = append(keys, aws.ToString(obj.Key))
		}
	}
	log.Printf("Fetched %d objects", len(keys))
	return keys
}

func (service *S3Impl) getListResponse(
	req *s3.ListObjectsV2Input) *s3.ListObjectsV2Output {
	response, err := service.client.ListObjectsV2(context.TODO(), req)
	if err != nil {
		panic(fmt.Errorf("Unable to list objects: %s", err))
	}
	return response
}

func (service *S3Impl) Fetch(objectName string, bRange byteRange) []byte {
	var rangeField string
	if bRange.end == 0 {
		rangeField = fmt.Sprintf("bytes=%d-", bRange.start)
	} else {
		rangeField = fmt.Sprintf("bytes=%d-%d", bRange.start, bRange.end)
	}
	req := &s3.GetObjectInput{
		Bucket: aws.String(service.bucket),
		Key:    aws.String(objectName),
		Range:  aws.String(rangeField),
	}
	res, err := service.client.GetObject(context.TODO(), req)
	if err != nil {
		panic(fmt.Errorf("GetObject request failed: %s", err))
	}
	defer res.Body.Close()
	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(fmt.Errorf("Unable to read response body: %s", err))
	}
	return resBytes
}
