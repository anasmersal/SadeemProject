package controllers

import (
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-jwt/initializers"
	"github.com/go-jwt/models"
	"golang.org/x/crypto/bcrypt"
)

// Retrieve user data
func UserData(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User assertion failed"})
		return
	}

	if err := initializers.DB.Model(&user).Association("Tags").Find(&user.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preload tags"})
		return
	}

	baseURL := os.Getenv("BASE_URL")

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"ID":    user.ID,
			"Email": user.Email,
			"Name":  user.Name,
			"Admin": user.Admin,
			"Image": baseURL + "image/user/" + user.Image,
			"Tags":  user.Tags,
		},
	})
}

// Edits user data
func EditUserData(c *gin.Context) {
	var body struct {
		Email    string
		Password string
		Name     string
	}
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found in context"})
		return
	}

	dbUser := user.(models.User)
	oldImagePath := dbUser.Image

	_, err := c.FormFile("Image")
	if err != nil {
		if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		filename, err := SaveUploadedImage(c, "Image", "images/users/", []string{".jpeg", ".png"})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		dbUser.Image = filename

		if oldImagePath != "" {
			err := os.Remove("images/users/" + oldImagePath)
			if err != nil {
				c.JSON(http.StatusBadRequest, err.Error())
			}
		}
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Email != "" {
		dbUser.Email = body.Email
	}
	if body.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to hash password"})
			return
		}
		dbUser.Password = string(hash)
	}
	if body.Name != "" {
		dbUser.Name = body.Name
	}

	if !IsValidEmail(dbUser.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format."})
		return
	}

	if err := initializers.DB.Save(&dbUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	baseURL := os.Getenv("BASE_URL")

	dbUser.Image = baseURL + "image/user/" + dbUser.Image

	c.JSON(http.StatusOK, dbUser)
}

// Retrieves all users data
func UsersData(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var users []models.User
	var totalUsersCount int64
	var err error
	offset := (page - 1) * pageSize

	query := c.Query("q")
	db := initializers.DB.Model(&models.User{})

	// Apply search criteria if query is provided
	if query != "" {
		db = db.Where("name LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%")
	}

	// Count total users matching the search criteria
	if err = db.Count(&totalUsersCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Fetch users with pagination and preload their tags
	if err = db.Preload("Tags").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Load base URL from environment variable
	baseURL := os.Getenv("BASE_URL")

	// Customize the user data for response
	var responseData []gin.H
	for _, user := range users {
		var tagsData []gin.H
		for _, tag := range user.Tags {
			tagsData = append(tagsData, gin.H{
				"id":    tag.ID,
				"name":  tag.Name,
				"image": baseURL + "image/tag/" + tag.Image,
			})
		}
		userData := gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"image": baseURL + "image/user/" + user.Image,
			"admin": user.Admin,
			"tags":  tagsData,
		}
		responseData = append(responseData, userData)
	}

	pagination := gin.H{
		"total_items":  totalUsersCount,
		"total_pages":  int(math.Ceil(float64(totalUsersCount) / float64(pageSize))),
		"current_page": page,
		"page_size":    pageSize,
	}

	c.JSON(http.StatusOK, gin.H{"users": responseData, "pagination": pagination})
}

func EditUserDataByID(c *gin.Context) {
	userID := c.Param("id")

	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var dbUser models.User
	if err := initializers.DB.First(&dbUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	oldImagePath := dbUser.Image

	if _, err := c.FormFile("Image"); err != nil {
		if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		filename, err := SaveUploadedImage(c, "Image", "images/users/", []string{".jpeg", ".png"})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		dbUser.Image = filename

		if oldImagePath != "" {
			err := os.Remove("images/users/" + oldImagePath)
			if err != nil {
				c.JSON(http.StatusBadRequest, err.Error())
			}
		}
	}

	var body struct {
		Email    string
		Password string
		Name     string
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Email != "" {
		dbUser.Email = body.Email
	}
	if body.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to hash password"})
			return
		}
		dbUser.Password = string(hash)
	}
	if body.Name != "" {
		dbUser.Name = body.Name
	}

	if !IsValidEmail(dbUser.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format."})
		return
	}

	if err := initializers.DB.Save(&dbUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	baseURL := os.Getenv("BASE_URL")

	dbUser.Image = baseURL + "image/user/" + dbUser.Image

	c.JSON(http.StatusOK, dbUser)
}
