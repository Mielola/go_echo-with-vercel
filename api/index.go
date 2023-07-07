package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello World")
}

func Handler(w http.ResponseWriter, r *http.Request) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/api/hello", Hello)

	e.ServeHTTP(w, r)
}