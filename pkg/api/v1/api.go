package v1

import (
	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) {
	e.GET("/", testHandler)
	e.GET("/s3summary", getS3SummarysHandler)
	// e.POST("/storage_report", getS3SummarysHandler)
	// e.POST("/storage_recommendation", storageRecommendationHandler)
}
