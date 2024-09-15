package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const TokenTableName = "token"

type Token struct {
	ID        int       `gorm:"primaryKey;column:id"`
	UserID    uuid.UUID `gorm:"column:user_id"`
	Hash      string    `gorm:"column:hash"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Token) TableName() string {
	return TokenTableName
}

func (m *Token) BeforeCreate(_ *gorm.DB) (err error) {
	curTime := time.Now()

	m.CreatedAt = curTime
	m.UpdatedAt = curTime
	return nil
}

func (m *Token) BeforeUpdate(_ *gorm.DB) (err error) {
	m.UpdatedAt = time.Now()
	return nil
}
