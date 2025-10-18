package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type CryptService struct{}

func NewCryptService() *CryptService {
	return &CryptService{}
}

func (c *CryptService) HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
	}
	return string(hash)
}

func (c *CryptService) ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
