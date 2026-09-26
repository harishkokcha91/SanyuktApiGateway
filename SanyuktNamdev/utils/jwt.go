package utils

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"github.com/joho/godotenv"
)

var (
	jwtSecret []byte
	jwtSecretMu sync.RWMutex
)

func init() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load()
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	
	// In test mode, use a default secret if not provided
	// This allows tests to run without requiring the env var
	if len(jwtSecret) == 0 {
		// Check if we're running tests
		if isTestMode() {
			jwtSecret = []byte("test-secret-key-that-is-at-least-32-bytes-long")
		} else {
			log.Fatal("JWT_SECRET environment variable is required but not set")
		}
	}
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 bytes (256 bits) for HS256")
	}
}

// isTestMode checks if the program is running in test mode
func isTestMode() bool {
	// Check for test flags in os.Args
	for _, arg := range os.Args {
		if arg == "-test.v" || arg == "-test.run" || arg == "-test.bench" {
			return true
		}
		if len(arg) > 5 && arg[:6] == "-test." {
			return true
		}
	}
	return false
}

// JWTSecretForTest returns the current JWT secret (for test helpers)
func JWTSecretForTest() []byte {
	jwtSecretMu.RLock()
	defer jwtSecretMu.RUnlock()
	return jwtSecret
}

// SetTestJWTSecret sets a test JWT secret (for testing only)
func SetTestJWTSecret(secret string) {
	jwtSecretMu.Lock()
	defer jwtSecretMu.Unlock()
	jwtSecret = []byte(secret)
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userId string, role string) (string, error) {
	if role == "" {
		role = "user"
	}
	claims := jwt.MapClaims{
		"userid": userId,
		"role":   role,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	jwtSecretMu.RLock()
	secret := jwtSecret
	jwtSecretMu.RUnlock()
	
	return token.SignedString(secret)
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		
		jwtSecretMu.RLock()
		secret := jwtSecret
		jwtSecretMu.RUnlock()
		
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// HashPassword hashes a plain-text password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a hashed password with a plain password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
