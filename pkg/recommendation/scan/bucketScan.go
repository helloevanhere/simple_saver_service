package scan

//This file contains scans for information on rules and policies that impact the entire bucket
//Currently those scans are: Lifecycle Rules, Versioning Status, and Object Storage Classes
import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/helloevanhere/simple_saver_service/pkg/summary"
)

type LifecycleDetail struct {
	Rules []*s3.LifecycleRule
}

// BucketScan contains information on rules and policies that impact the entire bucket
type BucketScan struct {
	LifecycleDetail  LifecycleDetail `json:"lifecycle_detail"`
	VersioningStatus string          `json:"versioning_status"`
	StorageClasses   []string        `json:"storage_classes"`
}

// Takes in a session, a bucket, and the objects in the bucket and returns a BucketScan
// which contains information on rules and policies that impact the entire bucket
func bucketScan(sess *session.Session, bucket summary.BucketSummary, bucketObjs *s3.ListObjectsV2Output) (BucketScan, error) {
	bucketScan := BucketScan{}
	var err error

	//Gets information about the buckets lifecycle policies
	bucketScan.LifecycleDetail, err = lifecycleScan(sess, bucket.Name)
	if err != nil {
		return bucketScan, err
	}
	//Gets the buckets versioning status
	bucketScan.VersioningStatus, err = versioningEnabledScan(sess, bucket.Name)
	if err != nil {
		return bucketScan, err
	}
	//Creates a unique array of the storage classes of the objects in the bucket
	bucketScan.StorageClasses, err = storageClassScan(bucketObjs)
	if err != nil {
		return bucketScan, err
	}

	return bucketScan, nil
}

// Takes in a session and bucket name and retrieves the Lifecycle Policy details
func lifecycleScan(sess *session.Session, bucketName string) (LifecycleDetail, error) {
	svc := s3.New(sess)

	// Call the GetBucketLifecycleConfiguration API to retrieve the lifecycle configuration of the bucket
	//AWS SDK LIST CALL
	output, err := svc.GetBucketLifecycleConfiguration(&s3.GetBucketLifecycleConfigurationInput{
		Bucket: &bucketName,
	})
	if err != nil {
		fmt.Println("Error getting bucket lifecycle configuration:", err)
		return LifecycleDetail{}, nil
	}

	// Convert the output to a local struct for easier handling
	lifecycleConfig := LifecycleDetail{
		Rules: output.Rules,
	}
	return lifecycleConfig, nil
}

// Takes a session and bucket name and returns the versioning Status of type string
func versioningEnabledScan(sess *session.Session, bucketName string) (string, error) {
	versioningStatus := "Not Enabled"
	svc := s3.New(sess)

	// Call the GetBucketVersioning API to get the versioning configuration of the bucket
	//AWS SDK LIST CALL
	versioningConfig, err := svc.GetBucketVersioning(&s3.GetBucketVersioningInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return versioningStatus, err
	}

	//No null dereference
	if versioningConfig.Status != nil {
		versioningStatus = *versioningConfig.Status
	}
	return versioningStatus, nil
}

// Takes in a ptr to an array of objects in a bucket
// Scans the objects in a bucket and returns a unique list of storage classes
func storageClassScan(bucketObjs *s3.ListObjectsV2Output) ([]string, error) {
	var classes []string
	unique := make(map[string]bool)

	for _, item := range bucketObjs.Contents {
		storageClass := *item.StorageClass
		if !unique[storageClass] {
			unique[storageClass] = true
			classes = append(classes, storageClass)
		}
	}

	return classes, nil
}
