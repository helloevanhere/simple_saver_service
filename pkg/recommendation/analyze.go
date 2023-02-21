package recommendation

import (
	"strings"
	"time"

	"github.com/helloevanhere/simple_saver_service/pkg/summary"
)

type Analysis struct {
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	AnalysisResults []interface{} `json:"analysis_result"`
}

func AnalyzeScans(scans []BucketScans) ([]Analysis, error) {
	analyses := []Analysis{}
	archive := Analysis{
		Name: "Archive Storage Analysis",
	}

	versioning := Analysis{
		Name: "Bucket Versioning Analysis",
	}

	lifecycle := Analysis{
		Name: "Lifecycle Management Analysis",
	}
	lifecycleResults := []LifecycleAnalysisResult{}

	compression := Analysis{
		Name: "Compressed Data Analysis",
	}

	duplicates := Analysis{
		Name: "Duplicate Data Analysis",
	}

	tempStorage := Analysis{
		Name: "Temporary Storage Analysis",
	}

	incomplete := Analysis{
		Name: "Incomplete Data Analysis",
	}

	for _, scan := range scans {
		//is archivable analysis
		archiveResult, err := archiveAnalysis(scan)
		if err != nil {
			return analyses, err
		}
		if archiveResult.BucketSummary.Name != "" {
			archive.AnalysisResults = append(archive.AnalysisResults, archiveResult)
		}
		//bucket versioning analysis
		versioningResult, err := versioningAnalysis(scan)
		if err != nil {
			return analyses, err
		}
		if versioningResult.BucketSummary.Name != "" {
			versioning.AnalysisResults = append(versioning.AnalysisResults, versioningResult)
		}
		//lifecycle management analysis
		lifecycleResult, err := lifecycleAnalysis(scan)
		if err != nil {
			return analyses, err
		}
		if lifecycleResult.BucketSummary.Name != "" {
			lifecycleResults = append(lifecycleResults, lifecycleResult)
		}
		for _, objectScan := range scan.Scans.ObjectScans {
			switch objectScan.DataCategory {
			case "incomplete_multipart_upload":
				dataResult, err := dataAnalysis(scan.BucketSummary, scan.Scans.BucketScan, objectScan)
				if err != nil {
					return analyses, err
				}
				if dataResult.BucketSummary.Name != "" {
					incomplete.AnalysisResults = append(incomplete.AnalysisResults, dataResult)
				}
			case "duplicate_objects":
				dataResult, err := dataAnalysis(scan.BucketSummary, scan.Scans.BucketScan, objectScan)
				if err != nil {
					return analyses, err
				}
				if dataResult.BucketSummary.Name != "" {
					duplicates.AnalysisResults = append(duplicates.AnalysisResults, dataResult)
				}
			case "compressible_objects":
				dataResult, err := dataAnalysis(scan.BucketSummary, scan.Scans.BucketScan, objectScan)
				if err != nil {
					return analyses, err
				}
				if dataResult.BucketSummary.Name != "" {
					compression.AnalysisResults = append(compression.AnalysisResults, dataResult)
				}
			}

		}

	}

	if len(lifecycleResults) > 0 {
		for _, r := range lifecycleResults {
			lifecycle.AnalysisResults = append(lifecycle.AnalysisResults, r)
		}
	}

	//temporary storage analysis
	tempStorageResults, err := temporaryStorageAnalysis(lifecycleResults)
	if err != nil {
		return analyses, err
	}
	if len(tempStorageResults) > 0 {
		for _, r := range tempStorageResults {
			tempStorage.AnalysisResults = append(tempStorage.AnalysisResults, r)
		}
	}

	analyses = []Analysis{archive, versioning, lifecycle, tempStorage, compression, duplicates, incomplete}

	return noActionRequired(analyses), nil

}

func noActionRequired(analyses []Analysis) []Analysis {
	updatedAnalyses := []Analysis{}
	for _, analysis := range analyses {
		if len(analysis.AnalysisResults) == 0 {
			analysis.AnalysisResults = append(analysis.AnalysisResults, "No Action Required")
		}
		updatedAnalyses = append(updatedAnalyses, analysis)
	}
	return updatedAnalyses
}

type DataAnalysisResult struct {
	BucketSummary summary.BucketSummary `json:"bucket_summary"`
	Data          ObjectScan            `json:"data"`
}

func dataAnalysis(bucketSummary summary.BucketSummary, bucketScan BucketScan, objectScan ObjectScan) (DataAnalysisResult, error) {
	analysisResult := DataAnalysisResult{}

	switch objectScan.DataCategory {
	case "incomplete_multipart_upload":
		if bucketScan.LifecycleDetail.Rules != nil {
			if objectScan.ObjectCount > 0 && bucketScan.LifecycleDetail.Rules[0].AbortIncompleteMultipartUpload == nil {
				analysisResult := DataAnalysisResult{
					BucketSummary: bucketSummary,
					Data:          objectScan,
				}
				return analysisResult, nil
			}
		}
	case "duplicate_objects":
		if objectScan.ObjectCount > 0 {
			analysisResult := DataAnalysisResult{
				BucketSummary: bucketSummary,
				Data:          objectScan,
			}
			return analysisResult, nil
		}
	case "compressible_objects":
		if objectScan.ObjectCount > 0 {
			analysisResult := DataAnalysisResult{
				BucketSummary: bucketSummary,
				Data:          objectScan,
			}
			return analysisResult, nil
		}

	}
	return analysisResult, nil
}

type ArchivableAnalysisResult struct {
	BucketSummary  summary.BucketSummary `json:"bucket_summary"`
	StorageClasses []string              `json:"storage_classes"`
}

func archiveAnalysis(scan BucketScans) (ArchivableAnalysisResult, error) {
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

func isArchivable(scan BucketScans) bool {
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
	BucketSummary    summary.BucketSummary `json:"bucket_summary"`
	VersioningStatus string                `json:"versioning_status"`
}

func versioningAnalysis(scan BucketScans) (VersioningAnalysisResult, error) {
	analysisResult := VersioningAnalysisResult{}
	if scan.Scans.BucketScan.LifecycleDetail.Rules != nil {
		if scan.Scans.BucketScan.VersioningStatus == "Enabled" && (scan.Scans.BucketScan.LifecycleDetail.Rules[0].NoncurrentVersionTransitions == nil && scan.Scans.BucketScan.LifecycleDetail.Rules[0].NoncurrentVersionExpiration == nil) {
			analysisResult := VersioningAnalysisResult{
				BucketSummary:    scan.BucketSummary,
				VersioningStatus: scan.Scans.BucketScan.VersioningStatus,
			}
			return analysisResult, nil
		}
	}

	return analysisResult, nil

}

type LifecycleAnalysisResult struct {
	BucketSummary   summary.BucketSummary `json:"bucket_summary"`
	LifecycleDetail LifecycleDetail       `json:"lifecycle_detail"`
}

func lifecycleAnalysis(scan BucketScans) (LifecycleAnalysisResult, error) {
	analysisResult := LifecycleAnalysisResult{}

	if scan.Scans.BucketScan.LifecycleDetail.Rules != nil {
		if scan.Scans.BucketScan.LifecycleDetail.Rules == nil || *scan.Scans.BucketScan.LifecycleDetail.Rules[0].Status == "Suspended" {
			analysisResult := LifecycleAnalysisResult{
				BucketSummary:   scan.BucketSummary,
				LifecycleDetail: scan.Scans.BucketScan.LifecycleDetail,
			}
			return analysisResult, nil

		}
	}

	return analysisResult, nil

}

func temporaryStorageAnalysis(analysisResultInput []LifecycleAnalysisResult) ([]LifecycleAnalysisResult, error) {
	analysisResultOutput := []LifecycleAnalysisResult{}

	for _, result := range analysisResultInput {
		if isTempStorage(result.BucketSummary.Name) {
			analysisResultOutput = append(analysisResultOutput, result)
			return analysisResultOutput, nil
		}
	}
	return analysisResultOutput, nil
}

func isTempStorage(bucketName string) bool {
	filenames := []string{"temp", "log", "tmp", "test"}

	for _, substr := range filenames {
		if strings.Contains(strings.ToLower(bucketName), substr) {
			return true
		}
	}

	return false
}
