package recommendation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/helloevanhere/simple_saver_service/pkg/awsHelpers"
	"github.com/helloevanhere/simple_saver_service/pkg/summary"
)

type LifecycleDetail struct {
	Rules []*s3.LifecycleRule
}

type BucketScan struct {
	LifecycleDetail  LifecycleDetail `json:"lifecycle_detail"`
	VersioningStatus string          `json:"versioning_status"`
	StorageClasses   []string        `json:"storage_classes"`
}

type ObjectScan struct {
	DataCategory string `json:"data_category"`
	DataSize     int64  `json:"data_size"`
	ObjectCount  int64  `json:"object_count"`
}

type Scans struct {
	BucketScan  BucketScan   `json:"bucket_scan"`
	ObjectScans []ObjectScan `json:"object_scans"`
}

type BucketScans struct {
	BucketSummary summary.BucketSummary `json:"bucket_summary"`
	Scans         Scans                 `json:"scan_results"`
}

func ScanS3(sess *session.Session, buckets []string) ([]BucketScans, error) {
	results := []BucketScans{}

	bucketsSummaries, err := summary.CreateBucketSummaries(sess, buckets)
	if err != nil {
		return results, err
	}

	for _, bucketSummary := range bucketsSummaries {
		bucketObjs, err := awsHelpers.ListBucketObjects(sess, bucketSummary.Name)
		if err != nil {
			return results, err
		}

		bucketScan, err := bucketScan(sess, bucketSummary, bucketObjs)
		if err != nil {
			return results, err
		}

		objectScans, err := objectScans(sess, bucketSummary, bucketObjs)
		if err != nil {
			return results, err
		}

		result := BucketScans{
			BucketSummary: bucketSummary,
			Scans: Scans{
				BucketScan:  bucketScan,
				ObjectScans: objectScans,
			},
		}
		results = append(results, result)

	}

	return results, nil
}

func bucketScan(sess *session.Session, bucket summary.BucketSummary, bucketObjs *s3.ListObjectsV2Output) (BucketScan, error) {
	bucketScan := BucketScan{}
	var err error

	bucketScan.LifecycleDetail, err = lifecycleScan(sess, bucket.Name)
	if err != nil {
		return bucketScan, err
	}
	bucketScan.VersioningStatus, err = versioningEnabledScan(sess, bucket.Name)
	if err != nil {
		return bucketScan, err
	}
	bucketScan.StorageClasses, err = storageClassScan(bucketObjs)
	if err != nil {
		return bucketScan, err
	}

	return bucketScan, nil
}

func objectScans(sess *session.Session, bucket summary.BucketSummary, bucketObjs *s3.ListObjectsV2Output) ([]ObjectScan, error) {
	objectScan := ObjectScan{}
	objectScans := []ObjectScan{}
	var err error

	scanTypes := []string{"incomplete_multipart_upload", "duplicate_objects", "compressible_objects"}

	for _, scanType := range scanTypes {
		switch scanType {
		case "incomplete_multipart_upload":
			objectScan.DataCategory = scanType
			objectScan.ObjectCount, objectScan.DataSize, err = incompleteMultipartUploadScan(sess, bucket.Name)
			if err != nil {
				return objectScans, err
			}
		case "duplicate_objects":
			objectScan.DataCategory = scanType
			objectScan.ObjectCount, objectScan.DataSize, err = duplicateObjectsScan(bucketObjs)
			if err != nil {
				return objectScans, err
			}
		case "compressible_objects":
			objectScan.DataCategory = scanType
			objectScan.ObjectCount, objectScan.DataSize, err = uncompressedObjectsScan(bucketObjs)
			if err != nil {
				return objectScans, err
			}
		}

		objectScans = append(objectScans, objectScan)

	}

	return objectScans, nil
}

func incompleteMultipartUploadScan(sess *session.Session, bucketName string) (int64, int64, error) {
	var totalSize int64
	svc := s3.New(sess)

	// Retrieve list of multipart uploads
	resp, err := svc.ListMultipartUploads(&s3.ListMultipartUploadsInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return 0, 0, err
	}
	//count of incomplete multipart uploads
	incompleteObjectCount := int64(len(resp.Uploads))

	//get total size of incomplete multipart uploads
	for _, upload := range resp.Uploads {
		// retrieve information about the parts of the upload
		parts, err := svc.ListParts(&s3.ListPartsInput{
			Bucket:   &bucketName,
			Key:      upload.Key,
			UploadId: upload.UploadId,
		})
		if err != nil {
			return 0, 0, err
		}

		// sum the sizes of all parts to calculate the upload size
		var uploadSize int64

		for _, part := range parts.Parts {
			uploadSize += *part.Size
		}

		// add the upload size to the total size
		totalSize += uploadSize
	}

	return incompleteObjectCount, totalSize, nil
}

func uncompressedObjectsScan(bucketObjs *s3.ListObjectsV2Output) (int64, int64, error) {
	var totalCount, totalSize int64

	for _, obj := range bucketObjs.Contents {
		if isCompressed(*obj.Key) {
			continue
		}
		if isCompressable(*obj.Key) {
			totalCount++
			totalSize += *obj.Size
		}
	}

	return totalCount, totalSize, nil
}

// isCompressed checks if a given object key ends in a compressed file extension
func isCompressed(filename string) bool {
	extensions := []string{".gz", ".zip", ".tar", ".bz2", ".tgz", ".snappy"}

	for _, ext := range extensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// isCompressed checks if a given object key ends in a compressed file extension
func isCompressable(filename string) bool {
	extensions := []string{".jpeg", ".mp3", ".mpeg", ".png", ".pdf", ".txt"}

	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return false
		}
	}
	return true
}

// Scans the Objects and returns a unique list of storage classes
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

func duplicateObjectsScan(bucketObjs *s3.ListObjectsV2Output) (int64, int64, error) {
	sizeMap := make(map[string]int64)
	var totalCount, totalSize int64

	if len(bucketObjs.Contents) < 2 {
		return 0, 0, nil
	}

	for _, obj := range bucketObjs.Contents {
		// Check if the object's size and ETag have already been seen
		key := fmt.Sprintf("%d-%s", *obj.Size, *obj.ETag)
		if _, ok := sizeMap[key]; ok {
			totalCount++
			totalSize += *obj.Size
		} else {
			sizeMap[key] = *obj.Size
		}
	}

	return totalCount, totalSize, nil
}

func versioningEnabledScan(sess *session.Session, bucketName string) (string, error) {
	versioningStatus := "Not Enabled"
	svc := s3.New(sess)

	// Call the GetBucketVersioning API to get the versioning configuration of the bucket
	versioningConfig, err := svc.GetBucketVersioning(&s3.GetBucketVersioningInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return versioningStatus, err
	}

	if versioningConfig.Status != nil {
		versioningStatus = *versioningConfig.Status
	}
	return versioningStatus, nil
}

func lifecycleScan(sess *session.Session, bucketName string) (LifecycleDetail, error) {
	svc := s3.New(sess)

	// Call the GetBucketLifecycleConfiguration API to retrieve the lifecycle configuration of the bucket
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
