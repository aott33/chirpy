package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	testSecret := "test-secret-key"
	testUserID := uuid.New()
	testDuration := time.Hour

	token, err := MakeJWT(testUserID, testSecret, testDuration)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Verify token can be parsed
	parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse generated token: %v", err)
	}

	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		t.Fatal("failed to extract claims")
	}

	if claims.Issuer != "chirpy" {
		t.Errorf("expected issuer 'chirpy', got '%s'", claims.Issuer)
	}

	if claims.Subject != testUserID.String() {
		t.Errorf("expected subject %s, got %s", testUserID.String(), claims.Subject)
	}
}

func TestValidateJWT(t *testing.T) {
	testSecret := "test-secret-key" // Note: your ValidateJWT hardcodes this
	testUserID := uuid.New()
	testDuration := time.Hour

	// Create a valid token
	token, err := MakeJWT(testUserID, testSecret, testDuration)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Test valid token
	userID, err := ValidateJWT(token, testSecret)
	if err != nil {
		t.Fatalf("ValidateJWT failed on valid token: %v", err)
	}

	if userID != testUserID {
		t.Errorf("expected userID %s, got %s", testUserID, userID)
	}

	// Test invalid token
	_, err = ValidateJWT("invalid.token.string", testSecret)
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}

	// Test expired token
	expiredToken, err := MakeJWT(testUserID, testSecret, -time.Hour)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = ValidateJWT(expiredToken, testSecret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}
