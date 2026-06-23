package model

import "time"

type EmailVerification struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Email     string    `json:"email" gorm:"column:email;size:128;not null;index:idx_email_code,priority:1"`
	Code      string    `json:"code" gorm:"column:code;size:6;not null;index:idx_email_code,priority:2"`
	Purpose   string    `json:"purpose" gorm:"column:purpose;size:16;default:register"`
	ExpiresAt time.Time `json:"expires_at" gorm:"column:expires_at;not null;index:idx_expires"`
	Verified  bool      `json:"verified" gorm:"column:verified;default:false"`
	CreatedAt time.Time `json:"created_at"`
}

func (EmailVerification) TableName() string {
	return "email_verifications"
}
