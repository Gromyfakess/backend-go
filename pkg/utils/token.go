package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JwtSecret stores the secret key for signing JWT tokens
// This must be set before using any token functions
var JwtSecret []byte

// InitJWT reads JWT secret from environment and sets it up
// This must be called when the application starts
func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("ERROR: JWT_SECRET is not set in .env file. This is required for security.")
	}
	if len(secret) < 32 {
		log.Fatal("ERROR: JWT_SECRET must be at least 32 characters long for security.")
	}
	JwtSecret = []byte(secret)
}

// HashPassword takes a plain text password and returns a secure hash
// Uses bcrypt with cost factor 14 (good balance of security and speed)
func HashPassword(password string) (string, error) {
	// Generate hash from password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword checks if a plain text password matches a hash
// Returns error if password doesn't match
func VerifyPassword(hashedPassword, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return fmt.Errorf("password does not match")
	}
	return nil
}

// GenerateAllTokens creates both access token and refresh token
// Access token expires in 20 minutes, refresh token expires in 7 days
// Returns: accessToken, refreshToken, accessExpiry, refreshExpiry, error
func GenerateAllTokens(userID uint, role string, canCRUD bool) (string, string, time.Time, time.Time, error) {
	// Create access token (expires in 20 minutes)
	accessExpiry := time.Now().Add(20 * time.Minute)
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"canCRUD": canCRUD,
		"exp":     accessExpiry.Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(JwtSecret)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("failed to create access token: %w", err)
	}

	// Create refresh token (expires in 7 days)
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     refreshExpiry.Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(JwtSecret)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, accessExpiry, refreshExpiry, nil
}

// GenerateAccessTokenOnly creates only an access token (used when refreshing)
// Access token expires in 20 minutes
// Returns: accessToken, expiryTime, error
func GenerateAccessTokenOnly(userID uint, role string, canCRUD bool) (string, time.Time, error) {
	// Create access token (expires in 20 minutes)
	accessExpiry := time.Now().Add(20 * time.Minute)
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"canCRUD": canCRUD,
		"exp":     accessExpiry.Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(JwtSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create access token: %w", err)
	}

	return accessTokenString, accessExpiry, nil
}

// GenerateRefreshTokenOnly creates only a refresh token (used for token rotation)
// Refresh token expires in 7 days
// Returns: refreshToken, expiryTime, error
func GenerateRefreshTokenOnly(userID uint) (string, time.Time, error) {
	// Create refresh token (expires in 7 days)
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     refreshExpiry.Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(JwtSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return refreshTokenString, refreshExpiry, nil
}