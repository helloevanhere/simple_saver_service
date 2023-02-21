package summary

//This package creates an S3 data snapshot summary of an AWS account

import (
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/helloevanhere/simple_saver_service/pkg/awsHelpers"
)

type S3Summary struct {
	TotalBucketCount int64           `json:"bucket_count_total"`
	TotalSize        int64           `json:"bucket_size_total"` //in bytes
	TotalObjectCount int64           `json:"object_count_total"`
	AvgObjectCount   int64           `json:"object_count_avg"`
	AvgSize          int64           `json:"bucket_size_avg"`
	BucketSummaries  []BucketSummary `json:"bucket_summaries"`
}

type BucketSummary struct {
	Name           string    `json:"bucket_name"`
	ObjectCount    int64     `json:"object_count"`
	Size           int64     `json:"bucket_size"`      //in bytes
	ModifiedLastAt time.Time `json:"modified_last_at"` //nil equivalent if empty
}

// Takes in a session obj and array of bucket names and returns an array of BucketSummary
func CreateBucketSummaries(sess *session.Session, buckets []string) ([]BucketSummary, error) {
	bucketSummaries := []BucketSummary{}

	// Loop through the buckets and get metadata
	for _, bucketName := range buckets {

		// Get bucket objects, AWS SDK LIST CALL
		objResp, err := awsHelpers.ListBucketObjects(sess, bucketName)
		if err != nil {
			return bucketSummaries, fmt.Errorf("error getting objects for bucket %s: %v", bucketName, err)
		}

		// Skip empty buckets
		if len(objResp.Contents) == 0 {
			//Create BucketSummary
			b := BucketSummary{
				Name:        bucketName,
				ObjectCount: 0,
				Size:        0,
				//TO DO: Make ModifiedLastAt the bucket's creation date, which requires ListBuckets
				ModifiedLastAt: time.Time{},
			}
			//add BucketSummary to final result
			bucketSummaries = append(bucketSummaries, b)
			continue
		}

		// Initialize variables for metadata
		var totalSize int64
		var lastModTime time.Time
		numObjects := len(objResp.Contents)

		// Loop through response and calculate metadata
		for _, obj := range objResp.Contents {
			totalSize += *obj.Size

			if obj.LastModified.After(lastModTime) {
				lastModTime = *obj.LastModified
			}
		}

		//Create BucketSummary
		b := BucketSummary{
			Name:           bucketName,
			ObjectCount:    int64(numObjects),
			Size:           totalSize,
			ModifiedLastAt: lastModTime,
		}

		//add BucketSummary to final result
		bucketSummaries = append(bucketSummaries, b)
	}

	return bucketSummaries, nil
}

// Takes in a session and array of bucket names and returns an S3Summary
func CreateS3Summary(sess *session.Session, buckets []string) (S3Summary, error) {
	summary := S3Summary{}
	var err error

	//Create [] of bucket summaries
	summary.BucketSummaries, err = CreateBucketSummaries(sess, buckets)
	if err != nil {
		return summary, fmt.Errorf("error getting bucket summaries: %v", err)
	}

	//Total Number of Buckets
	summary.TotalBucketCount = int64(len(summary.BucketSummaries))

	//Initialize variables for metadata
	var totalSize int64
	var totalObjectCount int64

	//Interate of BucketSummaries to create metadata
	for _, bucket := range summary.BucketSummaries {
		totalSize += bucket.Size
		totalObjectCount += bucket.ObjectCount
	}

	//Set metadata
	summary.TotalSize = totalSize
	summary.TotalObjectCount = totalObjectCount
	summary.AvgObjectCount = int64(totalObjectCount / summary.TotalBucketCount)
	summary.AvgSize = int64(totalSize / summary.TotalBucketCount)

	//sort BucketSummaries smallest to largest bucket by data size
	if len(summary.BucketSummaries) > 0 {
		sort.Slice(summary.BucketSummaries[:], func(i, j int) bool {
			return summary.BucketSummaries[i].Size < summary.BucketSummaries[j].Size
		})
	}

	return summary, nil
}
