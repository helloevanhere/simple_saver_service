package analyze

import (
	"github.com/helloevanhere/simple_saver_service/pkg/recommendation/scan"
	"github.com/helloevanhere/simple_saver_service/pkg/summary"
)

type AnalysisResult interface {
	GetBucketSummary() summary.BucketSummary
}

type Analysis struct {
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	AnalysisResults []AnalysisResult `json:"analysis_result"`
}

// Function that prevents empty AnalysisResults from being appended to the Analysis
func (a *Analysis) AppendAnalysisResult(result AnalysisResult) {
	if result.GetBucketSummary().Name != "" {
		a.AnalysisResults = append(a.AnalysisResults, result)
	}
}

// Takes in an array of BucketScans and returns an array of Analyses
func AnalyzeScans(scans []scan.BucketScans) ([]Analysis, error) {
	//Initalize Analysis variables
	//TO DO: Put Analysis metadata in DB and access via query
	archive := Analysis{
		Name:        "Archive Storage Analysis",
		Description: "Checks if you have buckets that archive data and if the storage class is suitable",
	}

	versioning := Analysis{
		Name:        "Bucket Versioning Analysis",
		Description: "Analyzes the Versioning Status on your buckets",
	}

	lifecycle := Analysis{
		Name:        "Lifecycle Management Analysis",
		Description: "Analyzes the Lifecycle Policies of your buckets",
	}

	tempStorage := Analysis{
		Name:        "Temporary Storage Analysis",
		Description: "Analyzes the Lifecycle Policies on buckets that have been detected to hold temporary data",
	}

	compression := Analysis{
		Name:        "Compressed Data Analysis",
		Description: "Analyzes if there are objects that can be compressed in your buckets",
	}

	duplicates := Analysis{
		Name:        "Duplicate Data Analysis",
		Description: "Analyzes if there are potentially duplicate objects in your buckets",
	}

	incomplete := Analysis{
		Name:        "Incomplete Data Analysis",
		Description: "Analyzes if there are Incomplete Multipart Uploads in your buckets and if you have the proper policies to manage them",
	}

	//Iterate through BucketScans and generate bucket analyses
	for _, scan := range scans {
		//is archivable bucket analysis
		archiveResult, err := archiveAnalysis(scan)
		if err != nil {
			return nil, err
		}
		archive.AppendAnalysisResult(archiveResult)

		//versioning status bucket analysis
		versioningResult, err := versioningAnalysis(scan)
		if err != nil {
			return nil, err
		}
		versioning.AppendAnalysisResult(versioningResult)

		//lifecycle management bucket analysis
		lifecycleResult, err := lifecycleAnalysis(scan)
		if err != nil {
			return nil, err
		}
		lifecycle.AppendAnalysisResult(lifecycleResult)

		//Iterate through objectScans to create object Analyses
		for _, objectScan := range scan.Scans.ObjectScans {
			switch objectScan.DataCategory {
			case "incomplete_multipart_upload":
				dataResult, err := ObjectAnalysis(scan.BucketSummary, scan.Scans.BucketScan, objectScan)
				if err != nil {
					return nil, err
				}
				if dataResult.BucketSummary.Name != "" {
					incomplete.AnalysisResults = append(incomplete.AnalysisResults, dataResult)
				}
			case "duplicate_objects":
				dataResult, err := ObjectAnalysis(scan.BucketSummary, scan.Scans.BucketScan, objectScan)
				if err != nil {
					return nil, err
				}
				if dataResult.BucketSummary.Name != "" {
					duplicates.AnalysisResults = append(duplicates.AnalysisResults, dataResult)
				}
			case "compressible_objects":
				dataResult, err := ObjectAnalysis(scan.BucketSummary, scan.Scans.BucketScan, objectScan)
				if err != nil {
					return nil, err
				}
				if dataResult.BucketSummary.Name != "" {
					compression.AnalysisResults = append(compression.AnalysisResults, dataResult)
				}
			}

		}

	}

	// temporary storage bucket analysis
	// Takes in lifecycle.AnalysisResults which is an array of AnalysisResult
	// for buckets that do not have lifecycle management policies
	tempStorageResults, err := temporaryStorageAnalysis(lifecycle.AnalysisResults)
	if err != nil {
		return nil, err
	}
	if len(tempStorageResults) > 0 {
		for _, r := range tempStorageResults {
			tempStorage.AppendAnalysisResult(r)
		}
	}

	analyses := []Analysis{archive, versioning, lifecycle, tempStorage, compression, duplicates, incomplete}

	return analyses, nil

}
