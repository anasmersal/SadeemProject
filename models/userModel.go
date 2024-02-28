package models

import "gorm.io/gorm"

// User represents the user model
type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Name     string
	Password string
	Image    string
	Admin    bool
	Tags     []Tag `gorm:"many2many:user_tags;"`
}

// Tag represents the tag model
type Tag struct {
	gorm.Model
	Name  string `gorm:"unique"`
	Image string
	Users []User `gorm:"many2many:user_tags;"`
}

// UserTag represents the many-to-many relationship between users and tags
type UserTag struct {
	gorm.Model
	UserID uint `gorm:"constraint:OnDelete:CASCADE;"`
	TagID  uint `gorm:"constraint:OnDelete:CASCADE;"`
}
