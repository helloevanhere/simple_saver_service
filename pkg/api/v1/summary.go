package v1

import (
	// "net/http"
	"os"
	// "github.com/labstack/echo/v4"
	"sort"
	"fmt"
	"time"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

type s3Summary struct {
	TotalCount			int64		`json:"bucket_count_total"`
	TotalSize			int64		`json:"bucket_size_total"`		//in bytes
	TotalObjectCount	int64		`json:"object_count_total"`
	AvgObjectCount		int64		`json:"object_count_avg"`
	AvgSize				int64		`json:"bucket_size_avg"`
	BucketSummaries []bucketSummary `json:"bucket_summaries"`
}

type bucketSummary struct {
	Name				string 		`json:"bucket_name"`
	ObjectCount 		int64 		`json:"object_count"`
	Size 				int64		`json:"bucket_size"` 			//in bytes
	LargestObjectSize 	int64		`json:"largest_object_size"` 	//in bytes
	SmallestObjectSize	int64		`json:"smallest_object_size"`	//in bytes
	ModifiedLastAt		time.Time 	`json:"modified_last_at"`		//nil equivalent if empty	
}

func createAWSSession() (*session.Session, error) {
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

func listS3Buckets(sess *session.Session) ([]string, error) {
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

func listBucketObjects(sess *session.Session, bucketName string) (*s3.ListObjectsV2Output, error) {
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

func createBucketSummaries(sess *session.Session, buckets []string) ([]bucketSummary, error) {
	b := bucketSummary{}
	bucketSummaries := []bucketSummary{}

	// Loop through the buckets and get metadata
	for _, bucketName := range buckets {
		// Get bucket objects
		objResp, err := listBucketObjects(sess, bucketName)
		if err != nil {
			return bucketSummaries, fmt.Errorf("error getting objects for bucket %s: %v", bucketName, err)
		}

		// Skip empty buckets
		if len(objResp.Contents) == 0 {
			b = bucketSummary{
				Name:	bucketName,
				ObjectCount: 0,
				Size: 0,
				LargestObjectSize: 0,
				SmallestObjectSize: 0,
				ModifiedLastAt: time.Time{},
			}
			bucketSummaries = append(bucketSummaries, b)
			continue
		}

		// Initialize variables for metadata
		var totalSize int64
		var largestSize int64
		var smallestSize int64
		var lastModTime time.Time
		numObjects := len(objResp.Contents)

		// Loop through objects and calculate metadata
		for _, obj := range objResp.Contents {
			totalSize += *obj.Size

			if *obj.Size > largestSize {
				largestSize = *obj.Size
			}

			if smallestSize == 0 || *obj.Size < smallestSize {
				smallestSize = *obj.Size
			}

			if obj.LastModified.After(lastModTime) {
				lastModTime = *obj.LastModified
			}
		}

		b = bucketSummary{
			Name:	bucketName,
			ObjectCount: int64(numObjects),
			Size: totalSize,
			LargestObjectSize: largestSize,
			SmallestObjectSize: smallestSize,
			ModifiedLastAt: lastModTime,
		} 

		bucketSummaries = append(bucketSummaries, b)
	}

	return bucketSummaries, nil
}

func createS3Summary(sess *session.Session, buckets []string) (s3Summary, error) {
	summary := s3Summary{}
	var err error

	summary.BucketSummaries, err = createBucketSummaries(sess, buckets)
	if err != nil {
		return summary, fmt.Errorf("error getting bucket summaries: %v", err)
	}

	summary.TotalCount = int64(len(summary.BucketSummaries))

	var totalSize int64
	var totalObjectCount int64
	for _, bucket := range summary.BucketSummaries {
		totalSize += bucket.Size
		totalObjectCount += bucket.ObjectCount
	}

	summary.TotalSize = totalSize
	summary.TotalObjectCount = totalObjectCount
	summary.AvgObjectCount = int64(totalObjectCount/summary.TotalCount)
	summary.AvgSize = int64(totalSize/summary.TotalCount)

	//sort Bucket summaries smallest bucket by data size to largest
	if len(summary.BucketSummaries) > 0 {
		sort.Slice(summary.BucketSummaries[:], func(i, j int) bool {
			return summary.BucketSummaries[i].Size < summary.BucketSummaries[j].Size
		})
	}

	return summary, nil
}