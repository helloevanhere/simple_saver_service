package v1

import (
	"net/http"
	// "os"
	"github.com/labstack/echo/v4"
	"fmt"
	"time"
	"strings"
    // "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
	// "github.com/aws/aws-sdk-go/aws/credentials"
)

type bucketsRequest struct {
    Buckets []string `json:"buckets"`
}  

func testHandler(c echo.Context) error {
	return c.HTML(http.StatusOK, "Hello, Docker! <3")
}

// // @Summary Get Storage Report
// // @Tags storage
// // @Description Get Visualizable Storage Report for the listed cloud accounts.
// // @Produce json
// // @Success 200 {object} string
// // @Failure 400 {object} api.httpError
// // @Failure 404 {object} api.httpError
// // @Param s3 buckets body []string
// // @Router /storage_report [post]
func storageReportHandler(c echo.Context) error {
	// Create a new AWS session with the credentials
	sess, err := createAWSSession()
	if err != nil {
		return fmt.Errorf("error creating AWS session: %v", err)
	}

	req := new(bucketsRequest)
    if err := c.Bind(req); err != nil {
        return c.String(http.StatusBadRequest, err.Error())
    }
    buckets := req.Buckets

	//If no buckets specified, retrieve all buckets
	if len(buckets) == 0{
		// Get List of Buckets
		buckets, err = listS3Buckets(sess)
		if err != nil {
			return fmt.Errorf("error retrieving bucket list: %v", err)
		}
	}

	bucketSummaries, err := createS3Summary(sess, buckets)
	if err != nil {
		return fmt.Errorf("error creating s3 summary csv: %v", err)
	}

	return c.JSON(http.StatusOK, bucketSummaries)
}

// // @Summary Get Storage Recommendations
// // @Tags storage
// // @Description Get Storage Recommendation List for the listed cloud accounts.
// // @Produce json
// // @Success 200 {object} []string
// // @Failure 400 {object} api.httpError
// // @Failure 404 {object} api.httpError
// // @Param accounts body []string true "Cloud Accounts"
// // @Router /storage_report [post]
func storageRecommendationHandler(c echo.Context) error {

	// Create a new AWS session with the credentials
	sess, err := createAWSSession()
	if err != nil {
		return fmt.Errorf("error creating AWS session: %v", err)
	}

	req := new(bucketsRequest)
    if err := c.Bind(req); err != nil {
        return c.String(http.StatusBadRequest, err.Error())
    }
    buckets := req.Buckets
	
	//If no buckets specified, retrieve all buckets
	if len(buckets) == 0{
		// Get List of Buckets
		buckets, err = listS3Buckets(sess)
		if err != nil {
			return fmt.Errorf("error retrieving bucket list: %v", err)
		}
	}

	scans, err := scanS3(sess, buckets)
	if err != nil {
		return fmt.Errorf("error creating s3 scans: %v", err)
	}

	bucketTidyRecs, err := bucketTidyRecs(scans)
	if err != nil {
		return fmt.Errorf("error creating bucket tidy recommendations: %v", err)
	}

	objectTidyRecs, err := objectTidyRecs(scans)
	if err != nil {
		return fmt.Errorf("error creating bucket tidy recommendations: %v", err)
	}

	report := Report{
		SaverSuggestionSummary: []RecGroup{bucketTidyRecs, objectTidyRecs},
		ScanResult: scans,
	}






	return c.JSON(http.StatusOK, report)
}

type Report struct {
	S3Status				string		`json:"s3_status"`
	SaverSuggestionSummary	[]RecGroup	`json:"saver_suggestion_summary"`
	ScanResult				[]ScanResult	`json:"complete_scan_results"`
}

type RecGroup struct {
	RecType			string	`json:"recommendation_type"`
	Recommendations []Rec	`json:"recommendations"`
}

type Rec struct {
	ScanType		string		`json:"scan_type"`
	TargetBuckets	[]string	`json:"target_buckets"`
	SimpleRec		string		`json:"simple_saver_suggestion"`
	SuperRec		string		`json:"super_saver_suggestion"`
}

type ScanResult struct {
	BucketName				string			`json:"bucket_name"`
	Size					int64			`json:"size"`
	ObjectCount				int64			`json:"object_count"`
	UncompressedDataSize	int64			`json:"uncompressed_data_size"`
	UncompressedObjectCount	int64			`json:"uncompressed_object_count"`
	IncompleteDataSize		int64			`json:"incomplete_data_size"`
	IncompleteObjectCount	int64			`json:"incomplete_object_count"`
	DuplicateDataSize		int64			`json:"duplicate_data_size"`
	DuplicateObjectCount	int64			`json:"duplicate_object_count"`
	ModifiedLastAt			time.Time		`json:"modified_last_at"`
	VersioningStatus		string			`json:"versioning_status"`
	LifecycleDetail			LifecycleDetail	`json:"lifecycle_detail"`
	TemporaryStorage		bool			`json:"temporary_storage_detected"`
	StorageClasses			[]string		`json:"storage_classes"`
}

type LifecycleDetail struct {
	Rules []*s3.LifecycleRule
}

func bucketTidyRecs(scans []ScanResult) (RecGroup, error) {
	var storageClassTargetBuckets, versioningTargetBuckets, lifecycleTargetBuckets []string

	for _, scan := range scans{
		if isArchivable(scan){
			storageClassTargetBuckets = append(storageClassTargetBuckets, scan.BucketName)
		}
		if scan.LifecycleDetail.Rules != nil {
			if scan.VersioningStatus == "Enabled" && (scan.LifecycleDetail.Rules[0].NoncurrentVersionTransitions == nil && scan.LifecycleDetail.Rules[0].NoncurrentVersionExpiration == nil){
				versioningTargetBuckets = append(versioningTargetBuckets, scan.BucketName)
			}
			if scan.LifecycleDetail.Rules == nil || *scan.LifecycleDetail.Rules[0].Status == "Suspended"{
				lifecycleTargetBuckets = append(lifecycleTargetBuckets, scan.BucketName)
			}
		}

	}

	// TO DO: Create Rec Metadata database so product/marketing can easily set language
	storageClassRec := Rec{
		ScanType: "Storage Class Scan Results",
		TargetBuckets: storageClassTargetBuckets,
		SimpleRec: "We suggest updating the storage class of objects in these buckets to S3 Glacier Storage.",
		SuperRec: "We suggest enabling lifecycle rules that automatically move older or infrequently accessed data to better suited storage class.",
	}

	versioningRec := Rec{
		ScanType: "Bucket Versioning Scan Results",
		TargetBuckets: versioningTargetBuckets,
		SimpleRec: "We suggest reviewing the purpose and content of these buckets to determine if versioning is necessary.",
		SuperRec: "We suggest enabling lifecycle rules that limit the number of versions per object.",
	}

	lifecycleRec := Rec{
		ScanType: "Lifecycle Management Scan Results",
		TargetBuckets: lifecycleTargetBuckets,
		SimpleRec: "We suggest enabling lifecycle rules that automatically move older or infrequently accessed data to better suited storage class.",
		SuperRec: "We suggest utilizing lifecycle filters to more precisely set lifecycle rules, such as only transitioning objects with a certain prefix.",
	}

	recommendations := []Rec{storageClassRec, versioningRec, lifecycleRec}

	bucketTidyRecs := RecGroup{
		RecType: "Bucket Tidy Recommendations",
		Recommendations: recommendations,
	}

	return bucketTidyRecs, nil
}

func isArchivable(scan ScanResult) bool {
	if isPastInactiveThreshold(scan.ModifiedLastAt){
		return true
	}

	archive_indicators:= []string{"backup","back-up","archive"}

	for _, key := range archive_indicators {
		if strings.Contains(scan.BucketName, key){
			for _, class := range scan.StorageClasses {
				if strings.Contains(strings.ToLower(class), "glacier") != true{
					return true
				}
			}
		}
	}

	return false
}

func isPastInactiveThreshold(modifiedLastAt time.Time) bool {
	tm, _ := time.Parse("YYYY-MM-DDThh:mm:ssZ","0001-01-01T00:00:00Z")
	
	if modifiedLastAt != tm{
		currentTime := time.Now()
		threeMonthsAgo := currentTime.AddDate(0,-3,0)
		
		return modifiedLastAt.Before(threeMonthsAgo)
	}
	return false

}

func objectTidyRecs(scans []ScanResult) (RecGroup, error) {
	var compressedDataTargetBuckets, duplicateDataTargetBuckets, incompleteDataTargetBuckets, temporaryStorageTargetBuckets []string

	for _, scan := range scans{
		if scan.UncompressedObjectCount > 0 {
			compressedDataTargetBuckets = append(compressedDataTargetBuckets, scan.BucketName)
		}
		if scan.DuplicateObjectCount > 0 {
			duplicateDataTargetBuckets = append(duplicateDataTargetBuckets, scan.BucketName)
		}
		if scan.TemporaryStorage {
			temporaryStorageTargetBuckets = append(temporaryStorageTargetBuckets, scan.BucketName)
		}
		if scan.LifecycleDetail.Rules != nil{
			if scan.IncompleteObjectCount > 0 && scan.LifecycleDetail.Rules[0].AbortIncompleteMultipartUpload == nil {
				incompleteDataTargetBuckets = append(incompleteDataTargetBuckets, scan.BucketName)
			}
		}
	}

	// TO DO: Create Rec Metadata database so product/marketing can easily set language
	compressionRec := Rec{
		ScanType: "Compressed DataScan Results",
		TargetBuckets: compressedDataTargetBuckets,
		SimpleRec: "We suggest compressing bucket objects where feasible.",
		SuperRec: "We suggest updating your data pipelines to compress data prior to storing in S3.",
	}

	duplicateDataRec := Rec{
		ScanType: "Duplicate Data Scan Results",
		TargetBuckets: duplicateDataTargetBuckets,
		SimpleRec: "We suggest deleting duplicate objects where feasible.",
		SuperRec: "We suggest updating your data pipelines to handle de-duplicating data prior to storing in S3, or Enable Bucket Versioning and Lifecycle Policies.",
	}

	incompleteDataRec := Rec{
		ScanType: "Incomplete Data Scan Results",
		TargetBuckets: incompleteDataTargetBuckets,
		SuperRec: "We suggest enabling Expire Incomplete Multipart Uploads in your lifecycle policy.",
	}

	tempStorageRec := Rec{
		ScanType: "Temporary Scan Results",
		TargetBuckets: duplicateDataTargetBuckets,
		SimpleRec: "We suggest enabling an Expiration Policy for buckets containing temporary data.",
		SuperRec: "We suggest disabling bucket versioning for buckets containing temporary data.",
	}

	recommendations := []Rec{compressionRec, duplicateDataRec, incompleteDataRec, tempStorageRec}

	objectTidyRecs := RecGroup{
		RecType: "Object Tidy Recommendations",
		Recommendations: recommendations,
	}

	return objectTidyRecs, nil
}

func scanS3(sess *session.Session, buckets []string) ([]ScanResult, error) {
	result := ScanResult{}
	results := []ScanResult{}

	bucketsSummaries, err := createBucketSummaries(sess, buckets)
	if err != nil {
		return results, err
	}

	for _, bucket := range(bucketsSummaries){

		result = ScanResult{
			BucketName:	bucket.Name,
			Size: bucket.Size,
			ObjectCount: bucket.ObjectCount,
			ModifiedLastAt: bucket.ModifiedLastAt,
		}

		result.IncompleteObjectCount, result.IncompleteDataSize, err = incompleteMultipartUploadScan(sess, bucket.Name)
		if err != nil {
			return results, err
		}

		bucketObjs, err := listBucketObjects(sess, bucket.Name)
		result.UncompressedObjectCount, result.UncompressedDataSize, err = uncompressedObjectsScan(bucketObjs)
	    if err != nil {
			return results, err
		}

		result.TemporaryStorage, err = tempStorageScan(bucket.Name)
		if err != nil {
			return results, err
		}

		result.StorageClasses, err = storageClassScan(bucketObjs)
		if err != nil {
			return results, err
		}

		result.VersioningStatus, err = versioningEnabledScan(sess, bucket.Name)
		if err != nil {
			return results, err
		}

		result.LifecycleDetail, err = lifecycleScan(sess, bucket.Name)
		if err != nil {
			return results, err
		}

		result.DuplicateObjectCount, result.DuplicateDataSize, err =  duplicateObjectsScan(bucketObjs)
		if err != nil {
			return results, err
		}

		results = append(results, result)
	}

	return results, nil
}

func incompleteMultipartUploadScan(sess *session.Session, bucketName string) (int64, int64, error) {
	var totalSize int64
    svc := s3.New(sess)

    // Retrieve list of multipart uploads
    resp, err := svc.ListMultipartUploads(&s3.ListMultipartUploadsInput{
        Bucket: &bucketName,
    })
    if err != nil {
        return 0,0, err
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
			return 0,0, err
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
		if isCompressable(*obj.Key){
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

func tempStorageScan(bucketName string) (bool, error) {
    filenames := []string{"temp","log","tmp","test"}

	for _, substr := range filenames {
		if strings.Contains(strings.ToLower(bucketName), substr) {
			return true, nil
		}
	}
   
    return false, nil
}

//Scans the Objects and returns a unique list of storage classes
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
		return 0,0,nil
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

	if versioningConfig.Status != nil{
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