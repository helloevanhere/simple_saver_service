package recommendation

import (
	"fmt"
	"strings"

	"github.com/helloevanhere/simple_saver_service/pkg/recommendation/analyze"
)

type Recommendation struct {
	Analysis      analyze.Analysis `json:"analysis"`
	TargetBuckets []string         `json:"target_buckets"`
	Recs          []Rec            `json:"recommendation"`
}

type Rec struct {
	Level string `json:"recommendation_level"`
	Text  string `json:"text"`
}

func CreateRecommendationReport(analyses []analyze.Analysis) ([]Recommendation, error) {
	recs := []Recommendation{}
	rec := Recommendation{}
	var err error

	for _, analysis := range analyses {
		rec.Analysis = analysis
		rec.TargetBuckets, err = collectTargetBuckets(analysis)
		if err != nil {
			return recs, err
		}

		rec.Recs, err = createRecommendations(analysis.Name, len(rec.TargetBuckets))
		if err != nil {
			return recs, err
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func createRecommendations(name string, bucketsImpacted int) ([]Rec, error) {
	recs := []Rec{}
	if bucketsImpacted > 0 {
		switch name {
		case "Archive Storage Analysis":
			recs = []Rec{
				{
					Level: "Simple Saver Suggestion",
					Text:  "We suggest updating the storage class of objects in these buckets to S3 Glacier Storage.",
				},
				{
					Level: "Super Saver Suggestion",
					Text:  "We suggest enabling lifecycle rules that automatically move older or infrequently accessed data to better suited storage class.",
				},
			}
		case "Bucket Versioning Analysis":
			recs = []Rec{
				{
					Level: "Simple Saver Suggestion",
					Text:  "We suggest reviewing the purpose and content of these buckets to determine if versioning is necessary.",
				},
				{
					Level: "Super Saver Suggestion",
					Text:  "We suggest enabling lifecycle rules that limit the number of versions per object.",
				},
			}

		case "Lifecycle Management Analysis":
			recs = []Rec{
				{
					Level: "Simple Saver Suggestion",
					Text:  "We suggest enabling lifecycle rules that automatically move older or infrequently accessed data to better suited storage class.",
				},
				{
					Level: "Super Saver Suggestion",
					Text:  "We suggest utilizing lifecycle filters to more precisely set lifecycle rules, such as only transitioning objects with a certain prefix.",
				},
			}

		case "Temporary Storage Analysis":
			recs = []Rec{
				{
					Level: "Simple Saver Suggestion",
					Text:  "We suggest enabling an Expiration Policy for buckets containing temporary data.",
				},
				{
					Level: "Super Saver Suggestion",
					Text:  "We suggest disabling bucket versioning for buckets containing temporary data.",
				},
			}
		case "Compressed Data Analysis":
			recs = []Rec{
				{
					Level: "Simple Saver Suggestion",
					Text:  "We suggest compressing objects in the listed buckets.",
				},
				{
					Level: "Super Saver Suggestion",
					Text:  "We suggest compressing data prior to storing in S3.",
				},
			}
		case "Duplicate Data Analysis":
			recs = []Rec{
				{
					Level: "Simple Saver Suggestion",
					Text:  "We suggest deleting duplicate objects.",
				},
				{
					Level: "Super Saver Suggestion",
					Text:  "We suggest de-duplicating data prior to storing in S3, or enabling Bucket Versioning and Lifecycle Policies.",
				},
			}
		case "Incomplete Data Analysis":
			recs = []Rec{
				{
					Level: "Simple Saver Suggestion",
					Text:  "We suggest deleting incomplete multipart uploads in the listed buckets.",
				},
				{
					Level: "Super Saver Suggestion",
					Text:  "We suggest enabling Expire Incomplete Multipart Uploads in your bucket's lifecycle policy.",
				},
			}
		}

	} else {
		recs = []Rec{
			{
				Level: "Saver Super Star",
				Text:  fmt.Sprintf("No suggestions needed! Our analysis shows that you're handling %s like a Saver Super Star", strings.ReplaceAll(name, " Analysis", "")),
			},
		}

	}

	return recs, nil

}

func collectTargetBuckets(analysis analyze.Analysis) ([]string, error) {
	targetBuckets := []string{}

	for _, result := range analysis.AnalysisResults {
		targetBuckets = append(targetBuckets, result.GetBucketSummary().Name)
	}

	if len(targetBuckets) == 0 {
		return nil, nil
	}

	return targetBuckets, nil

}
