package v1

import (
	"net/http"
	// "os"
	"github.com/labstack/echo/v4"
	"fmt"
	// "time"
    // "github.com/aws/aws-sdk-go/aws"
    // "github.com/aws/aws-sdk-go/aws/session"
    // "github.com/aws/aws-sdk-go/service/s3"
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
// func storageRecommendationHandler(c echo.Context) error {
// 	return c.JSON(http.StatusOK, []string{"Recommendation 1", "Recommendation 2"})
// }


// func createRecommendations(sess *session.Session, buckets []string) ([]recommendationSummary, error) {

// }