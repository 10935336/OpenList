package model

type UserGroup struct {
	ID          uint   `json:"id" gorm:"primaryKey"`                  // unique key
	Name        string `json:"name" gorm:"unique" binding:"required"` // group name
	Description string `json:"description"`
}
