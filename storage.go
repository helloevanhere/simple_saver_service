package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

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
// // @Param accounts body []string true "Cloud Accounts"
// // @Router /storage_report [post]
// func storageReportHandler(c echo.Context) error {
// 	return c.JSON(http.StatusOK, "Report Contents")
// }

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
