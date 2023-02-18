package v1

import (
	"net/http"
	"os"
	"github.com/labstack/echo/v4"
	"fmt"
	"time"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func testHandler(c echo.Context) error {
	return c.HTML(http.StatusOK, "Hello, Docker! <3")
}

func getS3SummarysHandler(c echo.Context) error {
	// Create a new AWS session with the credentials
	sess, err := createAWSSession()
	if err != nil {
		return fmt.Errorf("error creating AWS session: %v", err)
	}

	// Get List of Buckets
	buckets, err := listS3Buckets(sess)
	if err != nil {
		return fmt.Errorf("error retrieving bucket list: %v", err)
	}

	csvContent, err := createS3SummaryCSV(sess, buckets)
	if err != nil {
		return fmt.Errorf("error creating s3 summary csv: %v", err)
	}

	// Set the content type and disposition headers to force download
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=s3report.csv")

	// Write the CSV data to the response
	c.Response().Write([]byte(csvContent))

	return nil
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

func getBucketObjects(sess *session.Session, bucketName string) (*s3.ListObjectsV2Output, error) {
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

func createS3SummaryCSV(sess *session.Session, buckets []string) (string, error) {
	// Create CSV header
	csvContent := "Bucket Name,Object Count,Bucket Size,Largest Object Size,Smallest Object Size,Date Last Modified\n"

	// Loop through the buckets and get metadata
	for _, bucketName := range buckets {
		// Get bucket objects
		objResp, err := getBucketObjects(sess, bucketName)
		if err != nil {
			return "", fmt.Errorf("error getting objects for bucket %s: %v", bucketName, err)
		}

		// Skip empty buckets
		if len(objResp.Contents) == 0 {
			csvContent += fmt.Sprintf("%s,%d,%d,,-,-\n", bucketName, 0, 0)
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

		// Add metadata to CSV content
		csvContent += fmt.Sprintf("%s,%d,%d,%d,%d,%s\n", bucketName, numObjects, totalSize, largestSize, smallestSize, lastModTime.Format(time.RFC3339))
	}

	return csvContent, nil
}


// // @Summary Get Storage Report
// // @Tags storage
// // @Description Get Visualizable Storage Report for the listed cloud accounts.
// // @Produce json
// // @Success 200 {object} string
// // @Failure 400 {object} api.httpError
// // @Failure 404 {object} api.httpError
// // @Param accounts body []string true "Cloud Accounts"
// // @Router /storage_report [post]
// func storageReportHandler(c echo.Context) error {
// 	return c.JSON(http.StatusOK, "Report Contents")
// }

// // @Summary Get Storage Recommendations
// // @Tags storage
// // @Description Get Storage Recommendation List for the listed cloud accounts.
// // @Produce json
// // @Success 200 {object} []string
// // @Failure 400 {object} api.httpError
// // @Failure 404 {object} api.httpError
// // @Param accounts body []string true "Cloud Accounts"
// // @Router /storage_report [post]
// func storageRecommendationHandler(c echo.Context) error {
// 	return c.JSON(http.StatusOK, []string{"Recommendation 1", "Recommendation 2"})
// }