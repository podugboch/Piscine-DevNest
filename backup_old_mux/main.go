package main

import (
	"piscine-devnest/internal/model"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connect to database
	model.Connect()

	// Auto-migrate tables if you have models
	// model.DB.AutoMigrate(&model.User{}, &model.Post{})

	r := gin.Default()
	// app.RegisterRoutes(r)  // your routes here
	r.Run(":8080")
}
