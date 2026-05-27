package model

type User struct {
	BaseModel
	Username string `gorm:"size:64;not null;uniqueIndex" json:"username"`
	Password string `gorm:"size:256;not null" json:"-"`
	Email    string `gorm:"size:128" json:"email"`
	Phone    string `gorm:"size:32" json:"phone"`
	Status   int    `gorm:"default:1" json:"status"` // 1: active, 0: disabled
}
