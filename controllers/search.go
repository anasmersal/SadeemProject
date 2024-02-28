package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-jwt/initializers"
	"github.com/go-jwt/models"
)

// Performs a search
func Search(c *gin.Context) {
	query := c.Query("q")

	var tags []models.Tag
	var users []models.User

	if err := initializers.DB.Where("name LIKE ?", "%"+query+"%").Preload("Users").Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search tags"})
		return
	}

	if err := initializers.DB.Where("name LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%").Preload("Tags").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}

	if len(tags) == 0 && len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No results found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags":  tags,
		"users": users,
	})
}
