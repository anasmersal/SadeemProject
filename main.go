package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-jwt/controllers"
	"github.com/go-jwt/initializers"
	"github.com/go-jwt/middleware"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {
	r := gin.Default()
	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)

	r.GET("/user", middleware.RequireAuth, controllers.UserData)
	r.PATCH("/user/edit", middleware.RequireAuth, controllers.EditUserData)
	r.GET("/users", middleware.RequireAuth, middleware.RequireAdmin(), controllers.UsersData)
	r.PATCH("/user/edit/:id", middleware.RequireAuth, middleware.RequireAdmin(), controllers.EditUserDataByID)

	r.GET("/tags", middleware.RequireAuth, controllers.GetAllTags)
	r.POST("/tag", middleware.RequireAuth, middleware.RequireAdmin(), controllers.CreateTag)
	r.PATCH("/tag/edit/:id", middleware.RequireAuth, middleware.RequireAdmin(), controllers.EditTag)
	r.POST("/user/tag/:id", middleware.RequireAuth, middleware.RequireAdmin(), controllers.AddTagToUser)

	r.GET("/search", middleware.RequireAuth, controllers.Search)

	r.Run()
}
