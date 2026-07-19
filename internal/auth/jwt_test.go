package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWTAndValidateJWT(t *testing.T) {
	tokenSecret := "my-secret-key"
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000") // Example UUID
	// Create a JWT token
	token, err := MakeJWT(userID, tokenSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	// Validate the JWT token
	validatedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if validatedUserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, validatedUserID.String())
	}

}

func TestGetBearerToken(t *testing.T) {
	// Test with a valid Authorization header
	req := &http.Request{
		Header: http.Header{
			"Authorization": []string{"Bearer my-valid-token"},
		},
	}
	token, err := GetBearerToken(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if token != "my-valid-token" {
		t.Errorf("Expected token 'my-valid-token', got '%s'", token)
	}

	// Test with a missing Authorization header
	req = &http.Request{
		Header: http.Header{},
	}
	_, err = GetBearerToken(req)
	if err == nil {
		t.Fatal("Expected an error for missing Authorization header, got nil")
	}

	// Test with an invalid Authorization header format
	req = &http.Request{
		Header: http.Header{
			"Authorization": []string{"InvalidFormat"},
		},
	}
	_, err = GetBearerToken(req)
	if err == nil {
		t.Fatal("Expected an error for invalid Authorization header format, got nil")
	}
}
