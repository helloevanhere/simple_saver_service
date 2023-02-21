package main

import (
	"os"

	"github.com/helloevanhere/simple_saver_service/pkg/api/v1"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	v1.Register(e)

	// e.GET("/ping", func(c echo.Context) error {
	// 	return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
	// })

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	e.Logger.Fatal(e.Start(":" + httpPort))
}
