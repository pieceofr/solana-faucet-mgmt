package main

import (
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
)

func showErrorPage(c *gin.Context, errorMessage string) {
	tmpl := template.Must(template.New("error").Parse(`
		<!doctype html>
		<html>
			<head>
				<title>Error</title>
			</head>
			<body>
				<h1>Error</h1>
				<p>{{ . }}</p>
			</body>
		</html>
	`))

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusBadRequest, "")
	err := tmpl.Execute(c.Writer, errorMessage)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}
