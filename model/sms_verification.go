package model

import "time"

type SmsVerification struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Phone     string    `json:"phone" gorm:"column:phone;size:20;not null;index:idx_phone_code,priority:1"`
	Code      string    `json:"code" gorm:"column:code;size:6;not null;index:idx_phone_code,priority:2"`
	Purpose   string    `json:"purpose" gorm:"column:purpose;size:16;default:register"`
	ExpiresAt time.Time `json:"expires_at" gorm:"column:expires_at;not null;index:idx_expires"`
	Verified  bool      `json:"verified" gorm:"column:verified;default:false"`
	CreatedAt time.Time `json:"created_at"`
}

func (SmsVerification) TableName() string {
	return "sms_verifications"
}
