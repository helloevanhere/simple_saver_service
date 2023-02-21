package v1

import (
	"fmt"
	"net/http"

	"github.com/helloevanhere/simple_saver_service/pkg/awsHelpers"
	"github.com/helloevanhere/simple_saver_service/pkg/recommendation"
	"github.com/helloevanhere/simple_saver_service/pkg/recommendation/analyze"
	"github.com/helloevanhere/simple_saver_service/pkg/recommendation/scan"
	"github.com/helloevanhere/simple_saver_service/pkg/summary"
	"github.com/labstack/echo/v4"
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
	sess, err := awsHelpers.CreateAWSSession()
	if err != nil {
		return fmt.Errorf("error creating AWS session: %v", err)
	}

	req := new(bucketsRequest)
	if err := c.Bind(req); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	buckets := req.Buckets

	//If no buckets specified, retrieve all buckets
	if buckets[0] == "*" {
		// Get List of Buckets
		//AWS SDK LIST CALL
		buckets, err = awsHelpers.ListS3Buckets(sess)
		if err != nil {
			return fmt.Errorf("error retrieving bucket list: %v", err)
		}
	}

	bucketSummaries, err := summary.CreateS3Summary(sess, buckets)
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
	sess, err := awsHelpers.CreateAWSSession()
	if err != nil {
		return fmt.Errorf("error creating AWS session: %v", err)
	}

	req := new(bucketsRequest)
	if err := c.Bind(req); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	buckets := req.Buckets

	//If no buckets specified, retrieve all buckets
	if buckets[0] == "*" {
		// Get List of Buckets
		//AWS SDK LIST CALL
		buckets, err = awsHelpers.ListS3Buckets(sess)
		if err != nil {
			return fmt.Errorf("error retrieving bucket list: %v", err)
		}
	}

	scans, err := scan.ScanS3(sess, buckets)
	if err != nil {
		return fmt.Errorf("error creating s3 scans: %v", err)
	}

	analyses, err := analyze.AnalyzeScans(scans)
	if err != nil {
		return fmt.Errorf("error creating analyses: %v", err)
	}

	recommendations, err := recommendation.CreateRecommendationReport(analyses)
	if err != nil {
		return fmt.Errorf("error creating recommendations: %v", err)
	}

	return c.JSON(http.StatusOK, recommendations)
}

type Report struct {
	S3Status               string                          `json:"s3_status"`
	SaverSuggestionSummary []recommendation.Recommendation `json:"saver_suggestion_summary"`
	ScanResults            []scan.BucketScans              `json:"complete_scan_results"`
	TotalPotentialSavings  float64                         `json:"total_potential_savings"`
}
