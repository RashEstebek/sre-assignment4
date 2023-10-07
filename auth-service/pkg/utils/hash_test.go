package utils

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestHashPassword(t *testing.T) {
	testCases := []struct {
		password           string
		expectedHashLength int
	}{
		{password: "password123", expectedHashLength: 60},
		{password: "abc123", expectedHashLength: 60},
	}

	for _, tc := range testCases {

		hash := HashPassword(tc.password)

		if len(hash) != tc.expectedHashLength {
			t.Errorf("unexpected hash length for password '%s': got %d, want %d",
				tc.password, len(hash), tc.expectedHashLength)
		}

		// Is hash valid?
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tc.password))
		if err != nil {
			t.Errorf("failed to validate hash for password '%s': %v",
				tc.password, err)
		}
	}
}

func TestCheckPasswordHash(t *testing.T) {
	testCases := []struct {
		password    string
		hash        string
		expectMatch bool
	}{
		{password: "12345678", hash: "$2a$05$Zm.DWZJACVNkoLwlQePzGekTl0fcFjiL0Fs8emVWnzLM1HO9Jf1cu", expectMatch: true},
		{password: "password123", hash: "$2a$10$invalidhash", expectMatch: false},
		{password: "wrongpassword", hash: "$2a$10$abcdef1234567890", expectMatch: false},
	}

	for _, tc := range testCases {

		result := CheckPasswordHash(tc.password, tc.hash)

		if result != tc.expectMatch {
			t.Errorf("unexpected result for password '%s' and hash '%s': got %t, want %t",
				tc.password, tc.hash, result, tc.expectMatch)
		}
	}
}
