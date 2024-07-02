package routes

import (
	"inventory-management-system/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/items", controllers.GetItems)
	

	return r
}