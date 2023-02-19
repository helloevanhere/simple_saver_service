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

type bucketsRequest struct {
    Buckets []string `json:"buckets"`
}  

type bucketSummary struct {
	Name				string 		`json:"bucket_name"`
	ObjectCount 		int64 		`json:"object_count"`
	Size 				int64		`json:"bucket_size"`
	LargestObjectSize 	int64		`json:"largest_object_size"`
	SmallestObjectSize	int64		`json:"smallest_object_size"`
	ModifiedLastAt		time.Time 	`json:"modified_last_at"`	
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

func createS3Summary(sess *session.Session, buckets []string) ([]bucketSummary, error) {
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
			}
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