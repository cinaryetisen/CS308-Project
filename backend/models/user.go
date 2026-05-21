package models

import (
	"time"

	"medieval-store/security"

	"gorm.io/gorm"
)

// User represents a customer or manager in the store
type User struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"not null" json:"name"`
	TaxID       string    `json:"tax_id"`
	Email       string    `gorm:"unique;not null" json:"email"`
	HomeAddress string    `json:"home_address"`
	Password    string    `gorm:"not null" json:"password"` // Stores the bcrypt hash
	Role        string    `gorm:"default:'customer'" json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeSave encrypts PII before it lands in the DB column.
func (u *User) BeforeSave(tx *gorm.DB) error {
	enc, err := security.Encrypt(u.TaxID)
	if err != nil {
		return err
	}
	u.TaxID = enc

	enc, err = security.Encrypt(u.HomeAddress)
	if err != nil {
		return err
	}
	u.HomeAddress = enc
	return nil
}

// AfterSave decrypts back so the in-memory struct keeps holding plaintext
// for the caller (e.g., so UpdateProfile can c.JSON the response).
func (u *User) AfterSave(tx *gorm.DB) error {
	return decryptUserPII(u)
}

// AfterFind decrypts on read so callers never see ciphertext.
func (u *User) AfterFind(tx *gorm.DB) error {
	return decryptUserPII(u)
}

func decryptUserPII(u *User) error {
	dec, err := security.Decrypt(u.TaxID)
	if err != nil {
		return err
	}
	u.TaxID = dec

	dec, err = security.Decrypt(u.HomeAddress)
	if err != nil {
		return err
	}
	u.HomeAddress = dec
	return nil
}
