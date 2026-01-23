package main

import (
	"carbon-scribe/project-portal/project-portal-backend/internal/auth"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	handler := &auth.Handler{}

	auth.RegisterRoutes(r, handler)

	r.Run(":8080") // Server on port 8080
}
