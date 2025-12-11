package utils

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var JwtSecret []byte

func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is not set in .env file")
	}
	JwtSecret = []byte(secret)
}

func HashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), 14)
	return string(b), err
}

func VerifyPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func GenerateAllTokens(userID uint, role string, canCRUD bool) (string, string, time.Time, time.Time, error) {
	atExp := time.Now().Add(time.Minute * 20)
	atClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"canCRUD": canCRUD,
		"exp":     atExp.Unix(),
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString(JwtSecret)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	rtExp := time.Now().Add(time.Hour * 24 * 7)
	rtClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     rtExp.Unix(),
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString(JwtSecret)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	return accessToken, refreshToken, atExp, rtExp, nil
}

func GenerateAccessTokenOnly(userID uint, role string, canCRUD bool) (string, time.Time, error) {
	atExp := time.Now().Add(time.Minute * 20)
	atClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"canCRUD": canCRUD,
		"exp":     atExp.Unix(),
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString(JwtSecret)
	return accessToken, atExp, err
}
