package summary

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

func CreateBucketSummaries(sess *session.Session, buckets []string) ([]BucketSummary, error) {
	bucketSummaries := []BucketSummary{}

	// Loop through the buckets and get metadata
	for _, bucketName := range buckets {
		// Get bucket objects
		objResp, err := awsHelpers.ListBucketObjects(sess, bucketName)
		if err != nil {
			return bucketSummaries, fmt.Errorf("error getting objects for bucket %s: %v", bucketName, err)
		}

		// Skip empty buckets
		if len(objResp.Contents) == 0 {
			b := BucketSummary{
				Name:        bucketName,
				ObjectCount: 0,
				Size:        0,
				//TO DO: Make ModifiedLastAt the bucket's creation date
				ModifiedLastAt: time.Time{},
			}
			bucketSummaries = append(bucketSummaries, b)
			continue
		}

		// Initialize variables for metadata
		var totalSize int64
		var lastModTime time.Time
		numObjects := len(objResp.Contents)

		// Loop through objects and calculate metadata
		for _, obj := range objResp.Contents {
			totalSize += *obj.Size

			if obj.LastModified.After(lastModTime) {
				lastModTime = *obj.LastModified
			}
		}

		b := BucketSummary{
			Name:           bucketName,
			ObjectCount:    int64(numObjects),
			Size:           totalSize,
			ModifiedLastAt: lastModTime,
		}

		bucketSummaries = append(bucketSummaries, b)
	}

	return bucketSummaries, nil
}

func CreateS3Summary(sess *session.Session, buckets []string) (S3Summary, error) {
	summary := S3Summary{}
	var err error

	summary.BucketSummaries, err = CreateBucketSummaries(sess, buckets)
	if err != nil {
		return summary, fmt.Errorf("error getting bucket summaries: %v", err)
	}

	summary.TotalBucketCount = int64(len(summary.BucketSummaries))

	var totalSize int64
	var totalObjectCount int64
	for _, bucket := range summary.BucketSummaries {
		totalSize += bucket.Size
		totalObjectCount += bucket.ObjectCount
	}

	summary.TotalSize = totalSize
	summary.TotalObjectCount = totalObjectCount
	summary.AvgObjectCount = int64(totalObjectCount / summary.TotalBucketCount)
	summary.AvgSize = int64(totalSize / summary.TotalBucketCount)

	//sort Bucket summaries smallest bucket by data size to largest
	if len(summary.BucketSummaries) > 0 {
		sort.Slice(summary.BucketSummaries[:], func(i, j int) bool {
			return summary.BucketSummaries[i].Size < summary.BucketSummaries[j].Size
		})
	}

	return summary, nil
}
