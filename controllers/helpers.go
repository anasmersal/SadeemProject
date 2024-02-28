package controllers

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// IsValidEmail checks if the provided email string is in a valid format
func IsValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(emailRegex).MatchString(email)
}

// SaveUploadedImage saves the uploaded image and returns the filename or an error
func SaveUploadedImage(c *gin.Context, formFileKey, savePath string, allowedExtensions []string) (string, error) {
	file, err := c.FormFile(formFileKey)
	if err != nil {
		var err error = fmt.Errorf("failed to retrieve image")
		return "", err
	}

	ext := filepath.Ext(file.Filename)
	var isValidExtension bool
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			isValidExtension = true
			break
		}
	}
	if !isValidExtension {
		var err error = fmt.Errorf("invalid image format. only jpeg and png are supported")
		return "", err
	}

	filename := uuid.New().String() + ext

	if err := c.SaveUploadedFile(file, savePath+filename); err != nil {
		var err error = fmt.Errorf("failed to save image")
		return "", err
	}

	return filename, nil
}
