package models

import (
	"encoding/base64"
	"html"
	"strings"

	"github.com/voidchef/devops/utils/token"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName string `gorm:"size:255;not null;" json:"firstName"`
	LastName  string `gorm:"size:255;not null;" json:"lastName"`
	Email     string `gorm:"size:255;not null;unique" json:"email"`
	Password  string `gorm:"size:255;not null;" json:"password"`
}

func (user *User) SaveUser() (*User, error) {
	err := user.BeforeSave(DB)
	if err != nil {
		return &User{}, err
	}

	err = DB.Create(&user).Error
	if err != nil {
		return &User{}, err
	}
	return user, nil
}

func (user *User) BeforeSave(tx *gorm.DB) (err error) {
	//turn password into hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	//remove spaces in username
	user.Email = html.EscapeString(strings.TrimSpace(user.Email))

	return nil
}

func VerifyPassword(hashedPassword string, password string) error {
	decodedHash, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(decodedHash, []byte(password))
}

func LoginCheck(email string, password string) (string, error) {
	user := User{}

	err := DB.Model(User{}).Where("email = ?", email).Take(&user).Error
	if err != nil {
		return "", err
	}

	err = VerifyPassword(user.Password, password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return "", err
	}

	token, err := token.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}
