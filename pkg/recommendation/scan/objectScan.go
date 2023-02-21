package scan

//This file scans for information about data in a particular category

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/helloevanhere/simple_saver_service/pkg/summary"
)

// ObjectScan contains information about data in a particular category
type ObjectScan struct {
	DataCategory string `json:"data_category"`
	DataSize     int64  `json:"data_size"`
	ObjectCount  int64  `json:"object_count"`
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

// Scans for incomplete multipart uploads
// Takes in a session and bucketname and returns count of incomplete partial upload objects and the size of those objects
func incompleteMultipartUploadScan(sess *session.Session, bucketName string) (int64, int64, error) {
	var totalSize int64
	svc := s3.New(sess)

	//AWS SDK LIST CALL
	// Retrieve list of multipart uploads
	resp, err := svc.ListMultipartUploads(&s3.ListMultipartUploadsInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return 0, 0, err
	}
	//Count of incomplete multipart uploads
	incompleteObjectCount := int64(len(resp.Uploads))

	//Get total size of incomplete multipart uploads
	for _, upload := range resp.Uploads {
		// List information about the parts of the upload
		//AWS SDK LIST CALL
		parts, err := svc.ListParts(&s3.ListPartsInput{
			Bucket:   &bucketName,
			Key:      upload.Key,
			UploadId: upload.UploadId,
		})
		if err != nil {
			return 0, 0, err
		}

		// Sum the sizes of all incomplete multipart uploads
		var uploadSize int64
		for _, part := range parts.Parts {
			uploadSize += *part.Size
		}

		totalSize += uploadSize
	}

	return incompleteObjectCount, totalSize, nil
}

// Checks for potentially duplicate objects based on the Etag hash and size
// Returns the count of duplicates and size of duplicates
func duplicateObjectsScan(bucketObjs *s3.ListObjectsV2Output) (int64, int64, error) {
	sizeMap := make(map[string]int64)
	var totalCount, totalSize int64

	//Bucket needs at least two objects to have duplicate objects
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

// Scans for data that isn't compressed but could be
// Takes in a ptr to a list of objects in a bucket
// Returns the count of uncompressed/compressible objects and the size of those objects
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

// HELPER for uncompressedObjectsScan()
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

// HELPER for uncompressedObjectsScan()
// isCompressed checks if a given object key ends in a file extension that cannot be impactfully compressed
func isCompressable(filename string) bool {
	extensions := []string{".jpeg", ".mp3", ".mpeg", ".png", ".pdf", ".txt"}

	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return false
		}
	}
	return true
}
