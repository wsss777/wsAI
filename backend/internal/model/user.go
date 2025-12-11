package model

import "time"
import "gorm.io/gorm"

type User struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	Name      string         `gorm:"type:varchar(50)" json:"name"`
	Email     string         `gorm:"type:varchar(100);index" json:"email"`
	Username  string         `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	Password  string         `gorm:"type:varchar(255)" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
