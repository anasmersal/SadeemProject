package initializers

import "github.com/go-jwt/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{}, &models.Tag{}, &models.UserTag{})

}
