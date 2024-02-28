package controllers

import (
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

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"ID":    user.ID,
			"Email": user.Email,
			"Name":  user.Name,
			"Admin": user.Admin,
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

	c.JSON(http.StatusOK, dbUser)
}

// Retrieves all users data
func UsersData(c *gin.Context) {
	var users []map[string]interface{}
	if err := initializers.DB.Model(&models.User{}).Select("id, name, email, admin").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
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

	c.JSON(http.StatusOK, dbUser)
}
