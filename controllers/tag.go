package controllers

import (
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-jwt/initializers"
	"github.com/go-jwt/models"
)

// Creates a new tag
func CreateTag(c *gin.Context) {
	var body struct {
		Name string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}
	filename, err := SaveUploadedImage(c, "Image", "images/tags/", []string{".jpeg", ".png"})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	tag := models.Tag{
		Name:  body.Name,
		Image: filename,
	}

	result := initializers.DB.Create(&tag)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The tag name already exists",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag created successfully",
	})
}

// Edits a tag
func EditTag(c *gin.Context) {
	tagID := c.Param("id")

	id, err := strconv.ParseUint(tagID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	var dbTag models.Tag
	if err := initializers.DB.First(&dbTag, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	oldImagePath := dbTag.Image

	if _, err := c.FormFile("Image"); err != nil {
		if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		filename, err := SaveUploadedImage(c, "Image", "images/tags/", []string{".jpeg", ".png"})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		dbTag.Image = filename

		if oldImagePath != "" {
			err := os.Remove("images/tag/" + oldImagePath)
			if err != nil {
				c.JSON(http.StatusBadRequest, err.Error())
			}
		}
	}

	var body struct {
		Name string
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Name != "" {
		dbTag.Name = body.Name
	}

	if err := initializers.DB.Save(&dbTag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag"})
		return
	}

	baseURL := os.Getenv("BASE_URL")

	dbTag.Image = baseURL + "image/tag/" + dbTag.Image

	c.JSON(http.StatusOK, dbTag)
}

// Adds a tag to a user
func AddTagToUser(c *gin.Context) {
	userID := c.Param("id")

	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var dbUser models.User
	if err := initializers.DB.Preload("Tags").First(&dbUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var tag models.Tag
	if err := c.Bind(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, existingTag := range dbUser.Tags {
		if existingTag.Name == tag.Name {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tag already associated with user"})
			return
		}
	}

	var existingTag models.Tag
	if err := initializers.DB.Where("name = ?", tag.Name).First(&existingTag).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag not found"})
		return
	}

	dbUser.Tags = append(dbUser.Tags, existingTag)

	if err := initializers.DB.Save(&dbUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag added to user successfully",
	})
}

// Retrieve essential user data using the UserSummary struct.
type UserSummary struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"Email"`
	Image string `json:"Image"`
}

// Retrieves all tags from the database with pagination
func GetAllTags(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	sort := c.DefaultQuery("sort", "asc")
	query := c.Query("q")

	if sort != "asc" && sort != "desc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort order"})
		return
	}

	var totalTagsCount int64
	var err error
	db := initializers.DB.Model(&models.Tag{})

	// Apply search criteria if query is provided
	if query != "" {
		db = db.Where("name LIKE ?", "%"+query+"%")
	}

	if err = db.Count(&totalTagsCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}

	offset := (page - 1) * pageSize

	var tags []models.Tag
	dbQuery := initializers.DB.Preload("Users").Offset(offset).Limit(pageSize)

	if query != "" {
		dbQuery = dbQuery.Where("name LIKE ?", "%"+query+"%")
	}

	if sort == "asc" {
		dbQuery = dbQuery.Order("name asc")
	} else {
		dbQuery = dbQuery.Order("name desc")
	}

	if err = dbQuery.Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}

	var tagData []gin.H
	baseURL := os.Getenv("BASE_URL")

	for _, tag := range tags {
		var userSummaries []UserSummary
		for _, user := range tag.Users {
			userSummaries = append(userSummaries, UserSummary{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
				Image: baseURL + "image/user/" + user.Image,
			})
		}
		tagData = append(tagData, gin.H{
			"id":    tag.ID,
			"name":  tag.Name,
			"image": baseURL + "image/tag/" + tag.Image,
			"users": userSummaries,
		})
	}

	pagination := gin.H{
		"total_items":  totalTagsCount,
		"total_pages":  int(math.Ceil(float64(totalTagsCount) / float64(pageSize))),
		"current_page": page,
		"page_size":    pageSize,
	}

	c.JSON(http.StatusOK, gin.H{"tags": tagData, "pagination": pagination})
}
