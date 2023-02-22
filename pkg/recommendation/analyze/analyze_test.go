package analyze

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/helloevanhere/simple_saver_service/pkg/recommendation/scan"
	"github.com/helloevanhere/simple_saver_service/pkg/summary"
)

// Test AnalyzeScans
func TestAnalyzeScans(t *testing.T) {
	// Create some test data
	scans := []scan.BucketScans{
		{
			BucketSummary: summary.BucketSummary{
				Name: "test-bucket-1",
			},
			Scans: scan.Scans{
				ObjectScans: []scan.ObjectScan{
					{
						DataCategory: "incomplete_multipart_upload",
					},
				},
			},
		},
	}

	// Call the AnalyzeScans function
	results, err := AnalyzeScans(scans)
	if err != nil {
		t.Fatalf("AnalyzeScans returned an error: %v", err)
	}

	// Check that the results contain the expected data
	if len(results) != 7 {
		t.Fatalf("AnalyzeScans did not return the expected number of analyses")
	}
	if results[0].Name != "Archive Storage Analysis" {
		t.Fatalf("AnalyzeScans did not return the expected analysis results")
	}
	if results[1].Name != "Bucket Versioning Analysis" {
		t.Fatalf("AnalyzeScans did not return the expected analysis results")
	}
	if results[2].Name != "Lifecycle Management Analysis" {
		t.Fatalf("AnalyzeScans did not return the expected analysis results")
	}
	if results[3].Name != "Temporary Storage Analysis" {
		t.Fatalf("AnalyzeScans did not return the expected analysis results")
	}
	if results[4].Name != "Compressed Data Analysis" {
		t.Fatalf("AnalyzeScans did not return the expected analysis results")
	}
	if results[5].Name != "Duplicate Data Analysis" {
		t.Fatalf("AnalyzeScans did not return the expected analysis results")
	}
	if results[6].Name != "Incomplete Data Analysis" {
		t.Fatalf("AnalyzeScans did not return the expected analysis results")
	}
}

// Test Object Analysis
func TestObjectAnalysis(t *testing.T) {
	// Test incomplete_multipart_upload scenario
	bucketSummary := summary.BucketSummary{}
	bucketScan := scan.BucketScan{
		LifecycleDetail: scan.LifecycleDetail{
			Rules: []*s3.LifecycleRule{
				{
					AbortIncompleteMultipartUpload: nil,
				},
			},
		},
	}
	objectScan := scan.ObjectScan{
		DataCategory: "incomplete_multipart_upload",
		ObjectCount:  1,
	}
	result, err := ObjectAnalysis(bucketSummary, bucketScan, objectScan)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.BucketSummary != bucketSummary {
		t.Errorf("expected BucketSummary to be %v, but got %v", bucketSummary, result.BucketSummary)
	}
	if result.Data != objectScan {
		t.Errorf("expected Data to be %v, but got %v", objectScan, result.Data)
	}

	// Test duplicate_objects scenario
	objectScan = scan.ObjectScan{
		DataCategory: "duplicate_objects",
		ObjectCount:  1,
	}
	result, err = ObjectAnalysis(bucketSummary, bucketScan, objectScan)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.BucketSummary != bucketSummary {
		t.Errorf("expected BucketSummary to be %v, but got %v", bucketSummary, result.BucketSummary)
	}
	if result.Data != objectScan {
		t.Errorf("expected Data to be %v, but got %v", objectScan, result.Data)
	}

	// Test compressible_objects scenario
	objectScan = scan.ObjectScan{
		DataCategory: "compressible_objects",
		ObjectCount:  1,
	}
	result, err = ObjectAnalysis(bucketSummary, bucketScan, objectScan)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.BucketSummary != bucketSummary {
		t.Errorf("expected BucketSummary to be %v, but got %v", bucketSummary, result.BucketSummary)
	}
	if result.Data != objectScan {
		t.Errorf("expected Data to be %v, but got %v", objectScan, result.Data)
	}
}
