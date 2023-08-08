package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/voidchef/devops/controllers"
	"github.com/voidchef/devops/middleware"
	"github.com/voidchef/devops/models"
	"github.com/voidchef/devops/utils"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// Connect to database
	err = models.ConnectDatabase()
	if err != nil {
		fmt.Printf("Error connecting to database -> %s \n", err)
	} else {
		fmt.Println("Database Connected!")
	}

	// Connect to docker
	err = utils.ConnectToDocker()
	if err != nil {
		fmt.Printf("Error connecting to docker daemon -> %s \n", err)
	} else {
		fmt.Println("Docker Daemon Connected!")
	}

	router := gin.Default()

	public := router.Group("/api")

	// Auth routes
	public.POST("/auth/login", controllers.Login)
	public.POST("/auth/register", controllers.Register)

	private := router.Group("/api/docker")
	// Use the JwtAuthMiddleware middleware for private routes
	private.Use(middleware.JwtAuthMiddleware())

	// Docker routes
	private.GET("/containers", controllers.GetContainers)
	private.GET("/stats/:containerID", controllers.GetStats)
	private.POST("/startContainer/:containerID", controllers.StartContainer)
	private.POST("/stopContainer/:containerID", controllers.StopContainer)
	private.POST("/updateContainer/:containerID", controllers.UpdateContainer)
	private.DELETE("/deleteContainer/:containerID", controllers.DeleteContainer)

	// Run the server
	router.Run(":" + os.Getenv("PORT"))
}
