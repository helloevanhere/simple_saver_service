package scan

//This file scans for information about data in a particular category

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/helloevanhere/simple_saver_service/pkg/recommendation/estimate"
	"github.com/helloevanhere/simple_saver_service/pkg/summary"
)

// ObjectScan contains information about data in a particular category
type ObjectScan struct {
	DataCategory     string                    `json:"data_category"`
	DataSize         int64                     `json:"data_size"`
	ObjectCount      int64                     `json:"object_count"`
	EstimatedSavings estimate.EstimatedSavings `json:"estimated_savings"`
}

// Takes in a session, a bucket, and the objects in the bucket and returns a ObjectScan
// which contains information about data in a particular category
// Currently scan categories are: incomplete multipart uploads, potential duplicate objects, and uncompressed objects
func objectScans(sess *session.Session, bucket summary.BucketSummary, bucketObjs *s3.ListObjectsV2Output) ([]ObjectScan, error) {
	objectScan := ObjectScan{}
	objectScans := []ObjectScan{}
	var err error

	// Currently scans for incomplete multipart uploads, potential duplicate objects, and uncompressed objects
	scanTypes := []string{"incomplete_multipart_upload", "duplicate_objects", "compressible_objects"}

	for _, scanType := range scanTypes {
		switch scanType {
		case "incomplete_multipart_upload":
			objectScan.DataCategory = scanType
			objectScan.ObjectCount, objectScan.DataSize, objectScan.EstimatedSavings, err = incompleteMultipartUploadScan(sess, bucket.Name, bucketObjs)
			if err != nil {
				return objectScans, err
			}
		case "duplicate_objects":
			objectScan.DataCategory = scanType
			objectScan.ObjectCount, objectScan.DataSize, objectScan.EstimatedSavings, err = duplicateObjectsScan(bucketObjs)
			if err != nil {
				return objectScans, err
			}
		case "compressible_objects":
			objectScan.DataCategory = scanType
			objectScan.ObjectCount, objectScan.DataSize, objectScan.EstimatedSavings, err = uncompressedObjectsScan(bucketObjs)
			if err != nil {
				return objectScans, err
			}
		}

		objectScans = append(objectScans, objectScan)

	}

	return objectScans, nil
}

// Scans for incomplete multipart uploads
// Takes in a session and bucketname and returns count of incomplete partial upload objects and the size of those objects
func incompleteMultipartUploadScan(sess *session.Session, bucketName string, bucketObjs *s3.ListObjectsV2Output) (int64, int64, estimate.EstimatedSavings, error) {
	var totalSize int64
	var totalSavings float64
	svc := s3.New(sess)

	//AWS SDK LIST CALL
	// Retrieve list of multipart uploads
	resp, err := svc.ListMultipartUploads(&s3.ListMultipartUploadsInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return 0, 0, estimate.EstimatedSavings{}, err
	}
	//Count of incomplete multipart uploads
	incompleteObjectCount := int64(len(resp.Uploads))

	keysToFilterBy := []string{}
	//Get total size of incomplete multipart uploads
	for _, upload := range resp.Uploads {
		keysToFilterBy = append(keysToFilterBy, *upload.Key)
	}

	filteredObjects := filterObjects(bucketObjs, keysToFilterBy)
	for _, object := range filteredObjects.Contents {
		savings := estimate.SavingsForBytesDeletedByStorageClass(*object.Size, *object.StorageClass)
		totalSavings += savings
		totalSize += *object.Size
	}

	savings := estimate.EstimatedSavings{
		CalculatedMonthlylSavingsMin: totalSavings,
		CalculatedMonthlySavingsMax:  totalSavings,
	}

	return incompleteObjectCount, totalSize, savings, nil
}

// HELPER for incompleteMultipartUploadScan
func filterObjects(objects *s3.ListObjectsV2Output, keys []string) *s3.ListObjectsV2Output {
	filteredObjects := make([]*s3.Object, 0)

	for _, obj := range objects.Contents {
		for _, key := range keys {
			if *obj.Key == key {
				filteredObjects = append(filteredObjects, obj)
				break
			}
		}
	}

	return &s3.ListObjectsV2Output{
		Contents: filteredObjects,
	}
}

// Checks for potentially duplicate objects based on the Etag hash and size
// Returns the count of duplicates and size of duplicates
func duplicateObjectsScan(bucketObjs *s3.ListObjectsV2Output) (int64, int64, estimate.EstimatedSavings, error) {
	sizeMap := make(map[string]int64)
	var totalCount, totalSize int64
	var totalSavings float64

	//Bucket needs at least two objects to have duplicate objects
	if len(bucketObjs.Contents) < 2 {
		return 0, 0, estimate.EstimatedSavings{}, nil
	}

	for _, object := range bucketObjs.Contents {
		// Check if the object's size and ETag have already been seen
		key := fmt.Sprintf("%d-%s", *object.Size, *object.ETag)
		if _, ok := sizeMap[key]; ok {
			totalCount++
			totalSize += *object.Size
			savings := estimate.SavingsForBytesDeletedByStorageClass(*object.Size, *object.StorageClass)
			totalSavings += savings
		} else {
			sizeMap[key] = *object.Size
		}
	}
	savings := estimate.EstimatedSavings{
		CalculatedMonthlylSavingsMin: totalSavings,
		CalculatedMonthlySavingsMax:  totalSavings,
	}

	return totalCount, totalSize, savings, nil
}

// Scans for data that isn't compressed but could be
// Takes in a ptr to a list of objects in a bucket
// Returns the count of uncompressed/compressible objects and the size of those objects
func uncompressedObjectsScan(bucketObjs *s3.ListObjectsV2Output) (int64, int64, estimate.EstimatedSavings, error) {
	var totalCount, totalSize int64
	var totalMinSavings, totalMaxSavings float64

	for _, object := range bucketObjs.Contents {
		if isCompressed(*object.Key) {
			continue
		}
		ext := isCompressable(filepath.Ext(*object.Key))
		if ext != "" {
			totalCount++
			totalSize += *object.Size
			minSavings, maxSavings, err := estimate.SavingsForBytesCompressedByStorageClass(*object.Size, ext, *object.StorageClass)
			totalMinSavings += minSavings
			totalMaxSavings += maxSavings
			if err != nil {
				return 0, 0, estimate.EstimatedSavings{}, err
			}
		}
	}
	savings := estimate.EstimatedSavings{
		CalculatedMonthlylSavingsMin: totalMinSavings,
		CalculatedMonthlySavingsMax:  totalMaxSavings,
	}

	return totalCount, totalSize, savings, nil
}

// HELPER for uncompressedObjectsScan()
// isCompressed checks if a given object key ends in a compressed file extension
func isCompressed(filename string) bool {
	extensions := []string{".h264", ".zstd", ".7zip", ".gz", ".gzip", ".zip", ".tar", ".rar", ".bz2", ".bzip2", ".tgz", ".snappy", ".jpeg", ".mp3"}

	for _, ext := range extensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// HELPER for uncompressedObjectsScan()
// isCompressed returns the appropriate compression type for a given file extension
func isCompressable(extension string) string {
	switch extension {
	case ".gz", ".txt", ".log", ".md", ".yml", ".yaml", ".xml", ".json", ".csv", ".conf", ".py", ".java", ".go", ".js", ".rb", ".pl", ".php", ".html", ".css", ".scss", ".less", ".svg", ".pdf", ".par":
		return ".gzip"
	case ".png", ".gif", ".bmp", ".heif", ".heic":
		return ".jpeg"
	case ".wav", ".aac", ".ogg", ".wma":
		return ".mp3"
	case ".mp4", ".mov", ".avi", ".mkv":
		return ".h264"
	case ".avro", ".parquet", ".orc":
		return ".snappy"
	case ".zst":
		return ".zstd"
	case ".7z":
		return ".7zip"
	default:
		return ""
	}
}
