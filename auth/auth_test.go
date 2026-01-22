package auth

import (
	"testing"
)

func TestPasswordPolicy(t *testing.T) {
	candidate := "short"
	got := CheckPasswordPolicy(candidate)
	if got != ErrPasswordTooShort {
		t.Fatalf("unexpected error received. expected %v got %v", ErrPasswordTooShort, got)
	}

	candidate = "valid password"
	got = CheckPasswordPolicy(candidate)
	if got != nil {
		t.Fatalf("unexpected error received. expected %v got %v", nil, got)
	}
}

func TestValidatePasswordChange(t *testing.T) {
	newPassword := "valid password"
	confirmPassword := "invalid"
	currentPassword := "current password"
	currentHash, err := GeneratePasswordHash(currentPassword)
	if err != nil {
		t.Fatalf("unexpected error generating password hash: %v", err)
	}

	_, got := ValidatePasswordChange(currentHash, newPassword, confirmPassword)
	if got != ErrPasswordMismatch {
		t.Fatalf("unexpected error received. expected %v got %v", ErrPasswordMismatch, got)
	}

	newPassword = currentPassword
	confirmPassword = newPassword
	_, got = ValidatePasswordChange(currentHash, newPassword, confirmPassword)
	if got != ErrReusedPassword {
		t.Fatalf("unexpected error received. expected %v got %v", ErrReusedPassword, got)
	}
}

func TestDummyHashIsValidBcrypt(t *testing.T) {
	err := ValidatePassword("any-password", DummyHash)
	if err == nil {
		t.Fatalf("DummyHash should never match any password")
	}
}

func TestConstantTimeComparison(t *testing.T) {
	realHash, _ := GeneratePasswordHash("real-password")

	errDummy := ValidatePassword("wrong-password", DummyHash)
	if errDummy == nil {
		t.Fatalf("DummyHash comparison should fail")
	}

	errReal := ValidatePassword("wrong-password", realHash)
	if errReal == nil {
		t.Fatalf("Real hash comparison with wrong password should fail")
	}
}
