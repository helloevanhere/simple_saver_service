package scan

//This package retrieves and scans information from AWS regarding a list of buckets

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/helloevanhere/simple_saver_service/pkg/awsHelpers"
	"github.com/helloevanhere/simple_saver_service/pkg/summary"
)

// Packs both scan types together (probably not necessary)
type Scans struct {
	BucketScan  BucketScan   `json:"bucket_scan"`
	ObjectScans []ObjectScan `json:"object_scans"`
}

// Contains all scans and BucketSummary information for one bucket (could probably use a better name)
type BucketScans struct {
	BucketSummary summary.BucketSummary `json:"bucket_summary"`
	Scans         Scans                 `json:"scan_results"`
}

// Takes in a session and array of bucket names and returns the BucketScans for their data
func ScanS3(sess *session.Session, buckets []string) ([]BucketScans, error) {
	results := []BucketScans{}

	//Creates []BucketSummary of all buckets provided
	bucketsSummaries, err := summary.CreateBucketSummaries(sess, buckets)
	if err != nil {
		return results, err
	}

	//Iterate through BucketSummaries to create both scan types
	for _, bucketSummary := range bucketsSummaries {

		//AWS SDK LIST CALL
		bucketObjs, err := awsHelpers.ListBucketObjects(sess, bucketSummary.Name)
		if err != nil {
			return results, err
		}

		//Create bucketScan
		bucketScan, err := bucketScan(sess, bucketSummary, bucketObjs)
		if err != nil {
			return results, err
		}

		//Create objectScan
		objectScans, err := objectScans(sess, bucketSummary, bucketObjs)
		if err != nil {
			return results, err
		}

		//Create BucketScans object for one bucket
		result := BucketScans{
			BucketSummary: bucketSummary,
			Scans: Scans{
				BucketScan:  bucketScan,
				ObjectScans: objectScans,
			},
		}

		//append BucketScans
		results = append(results, result)

	}

	return results, nil
}
