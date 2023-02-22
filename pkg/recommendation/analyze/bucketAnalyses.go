package analyze

//This file handles bucket Analyses

import (
	"strings"
	"time"

	"github.com/helloevanhere/simple_saver_service/pkg/recommendation/estimate"
	"github.com/helloevanhere/simple_saver_service/pkg/recommendation/scan"
	"github.com/helloevanhere/simple_saver_service/pkg/summary"
)

type ArchivableAnalysisResult struct {
	BucketSummary    summary.BucketSummary     `json:"bucket_summary"`
	StorageClasses   []string                  `json:"storage_classes"`
	EstimatedSavings estimate.EstimatedSavings `json:"estimated_savings"`
}

// Statisfies AnalysisResult interface
func (r ArchivableAnalysisResult) GetBucketSummary() summary.BucketSummary {
	return r.BucketSummary
}

// Statisfies AnalysisResult interface
func (r ArchivableAnalysisResult) GetEstimates() estimate.EstimatedSavings {
	return r.EstimatedSavings
}

// Checks if a bucket contains archive/less frequently access data and
// can be moved to better suited storage class
func archiveAnalysis(scan scan.BucketScans) (ArchivableAnalysisResult, error) {
	analysisResult := ArchivableAnalysisResult{}
	if isArchivable(scan) {
		analysisResult := ArchivableAnalysisResult{
			BucketSummary:  scan.BucketSummary,
			StorageClasses: scan.Scans.BucketScan.StorageClasses,
		}
		return analysisResult, nil
	}

	return analysisResult, nil

}

// HELPER for archiveAnalysis()
func isArchivable(scan scan.BucketScans) bool {
	if isPastInactiveThreshold(scan.BucketSummary.ModifiedLastAt) {
		return true
	}

	archive_indicators := []string{"backup", "back-up", "archive"}

	for _, key := range archive_indicators {
		if strings.Contains(scan.BucketSummary.Name, key) {
			for _, class := range scan.Scans.BucketScan.StorageClasses {
				if !strings.Contains(strings.ToLower(class), "glacier") {
					return true
				}
			}
		}
	}

	return false
}

// HELPER for archiveAnalysis()
func isPastInactiveThreshold(modifiedLastAt time.Time) bool {
	tm, _ := time.Parse("YYYY-MM-DDThh:mm:ssZ", "0001-01-01T00:00:00Z")

	if modifiedLastAt != tm {
		currentTime := time.Now()
		threeMonthsAgo := currentTime.AddDate(0, -3, 0)

		return modifiedLastAt.Before(threeMonthsAgo)
	}
	return false

}

type VersioningAnalysisResult struct {
	BucketSummary    summary.BucketSummary     `json:"bucket_summary"`
	VersioningStatus string                    `json:"versioning_status"`
	EstimatedSavings estimate.EstimatedSavings `json:"estimated_savings"`
}

// Statisfies AnalysisResult interface
func (r VersioningAnalysisResult) GetBucketSummary() summary.BucketSummary {
	return r.BucketSummary
}

// Statisfies AnalysisResult interface
func (r VersioningAnalysisResult) GetEstimates() estimate.EstimatedSavings {
	return r.EstimatedSavings
}

// Takes in BucketScans and returns VersioningAnalysisResult
// for buckets that do not have versioning
func versioningAnalysis(scan scan.BucketScans) (VersioningAnalysisResult, error) {
	analysisResult := VersioningAnalysisResult{}

	//Checks if versioning is enabled with no lifecycle policies managing versions
	if scan.Scans.BucketScan.VersioningStatus == "Enabled" {
		if scan.Scans.BucketScan.LifecycleDetail.Rules == nil || len(scan.Scans.BucketScan.LifecycleDetail.Rules) == 0 || (scan.Scans.BucketScan.LifecycleDetail.Rules[0].NoncurrentVersionTransitions == nil && scan.Scans.BucketScan.LifecycleDetail.Rules[0].NoncurrentVersionExpiration == nil) {
			analysisResult = VersioningAnalysisResult{
				BucketSummary:    scan.BucketSummary,
				VersioningStatus: scan.Scans.BucketScan.VersioningStatus,
			}
			return analysisResult, nil
		}
	}

	return analysisResult, nil
}

type LifecycleAnalysisResult struct {
	BucketSummary    summary.BucketSummary     `json:"bucket_summary"`
	LifecycleDetail  scan.LifecycleDetail      `json:"lifecycle_detail"`
	EstimatedSavings estimate.EstimatedSavings `json:"estimated_savings"`
}

// Statisfies AnalysisResult interface
func (r LifecycleAnalysisResult) GetBucketSummary() summary.BucketSummary {
	return r.BucketSummary
}

// Statisfies AnalysisResult interface
func (r LifecycleAnalysisResult) GetEstimates() estimate.EstimatedSavings {
	return r.EstimatedSavings
}

// Takes in BucketScans and returns LifecycleAnalysisResult
// for buckets that do not have lifecycle policies
func lifecycleAnalysis(scan scan.BucketScans) (LifecycleAnalysisResult, error) {
	analysisResult := LifecycleAnalysisResult{}

	if scan.Scans.BucketScan.LifecycleDetail.Rules == nil {
		analysisResult := LifecycleAnalysisResult{
			BucketSummary:   scan.BucketSummary,
			LifecycleDetail: scan.Scans.BucketScan.LifecycleDetail,
		}
		return analysisResult, nil

	} else if *scan.Scans.BucketScan.LifecycleDetail.Rules[0].Status == "Suspended" {
		analysisResult := LifecycleAnalysisResult{
			BucketSummary:   scan.BucketSummary,
			LifecycleDetail: scan.Scans.BucketScan.LifecycleDetail,
		}
		return analysisResult, nil
	}

	return analysisResult, nil
}

// Takes in []AnalysisResult and returns []AnalysisResult
// Takes in the output of lifecycleAnalysis() because
// we want a list of temp storage that does not have lifecycle policies.
func temporaryStorageAnalysis(analysisResultInput []AnalysisResult) ([]AnalysisResult, error) {
	analysisResultOutput := []AnalysisResult{}

	for _, result := range analysisResultInput {
		if isTempStorage(result.GetBucketSummary().Name) {
			analysisResultOutput = append(analysisResultOutput, result)
			return analysisResultOutput, nil
		}
	}
	return analysisResultOutput, nil
}

// HELPER for temporaryStorageAnalysis()
// Checks is bucket name contains words often used to label temporary storage
func isTempStorage(bucketName string) bool {
	filenames := []string{"temp", "log", "tmp", "test"}

	for _, substr := range filenames {
		if strings.Contains(strings.ToLower(bucketName), substr) {
			return true
		}
	}

	return false
}
