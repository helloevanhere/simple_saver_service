package awsHelpers

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func CreateAWSSession() (*session.Session, error) {
	// Get the AWS credentials from the environment variables
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")

	// Create a new AWS session with the credentials
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating AWS session: %v", err)
	}
	return sess, nil
}

func ListS3Buckets(sess *session.Session) ([]string, error) {
	svc := s3.New(sess)

	result, err := svc.ListBuckets(nil)
	if err != nil {
		return nil, err
	}

	bucketNames := make([]string, 0, len(result.Buckets))

	for _, bucket := range result.Buckets {
		bucketNames = append(bucketNames, *bucket.Name)
	}

	return bucketNames, nil
}

func ListBucketObjects(sess *session.Session, bucketName string) (*s3.ListObjectsV2Output, error) {
	svc := s3.New(sess)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	resp, err := svc.ListObjectsV2(input)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
