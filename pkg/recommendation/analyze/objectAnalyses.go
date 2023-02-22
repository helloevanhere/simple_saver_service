package analyze

import (
	"github.com/helloevanhere/simple_saver_service/pkg/recommendation/estimate"
	"github.com/helloevanhere/simple_saver_service/pkg/recommendation/scan"
	"github.com/helloevanhere/simple_saver_service/pkg/summary"
)

type ObjectAnalysisResult struct {
	BucketSummary summary.BucketSummary `json:"bucket_summary"`
	Data          scan.ObjectScan       `json:"data"`
}

// Statisfies AnalysisResult interface
func (r ObjectAnalysisResult) GetBucketSummary() summary.BucketSummary {
	return r.BucketSummary
}

// Statisfies AnalysisResult interface
func (r ObjectAnalysisResult) GetEstimates() estimate.EstimatedSavings {
	return r.Data.EstimatedSavings
}

// Takes in BucketSummary, BucketScan and ObjectScan and returns ObjectAnalysisResult
func ObjectAnalysis(bucketSummary summary.BucketSummary, bucketScan scan.BucketScan, objectScan scan.ObjectScan) (ObjectAnalysisResult, error) {
	analysisResult := ObjectAnalysisResult{}

	// currentS3Cost, err := estimate.CurrentStorageCost(bucketSummary.Size, bucketScan.StorageClasses[0])
	// if err != nil {
	// 	return ObjectAnalysisResult{}, err
	// }

	switch objectScan.DataCategory {
	case "incomplete_multipart_upload":
		//no null dereference
		if objectScan.ObjectCount > 0 && (bucketScan.LifecycleDetail.Rules == nil || len(bucketScan.LifecycleDetail.Rules) == 0 || bucketScan.LifecycleDetail.Rules[0].AbortIncompleteMultipartUpload == nil) {
			analysisResult := ObjectAnalysisResult{
				BucketSummary: bucketSummary,
				Data:          objectScan,
			}
			return analysisResult, nil
		}

	case "duplicate_objects":
		//don't return empty result
		if objectScan.ObjectCount > 0 {
			analysisResult := ObjectAnalysisResult{
				BucketSummary: bucketSummary,
				Data:          objectScan,
			}
			return analysisResult, nil
		}
	case "compressible_objects":
		//don't return empty result
		if objectScan.ObjectCount > 0 {
			analysisResult := ObjectAnalysisResult{
				BucketSummary: bucketSummary,
				Data:          objectScan,
			}
			return analysisResult, nil
		}

	}
	return analysisResult, nil
}
