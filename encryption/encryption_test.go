package encryption

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	plaintext := "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----"

	encrypted, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if encrypted[:4] != "ENC:" {
		t.Errorf("Encrypted string should start with ENC: prefix")
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted text doesn't match original: got %s, want %s", decrypted, plaintext)
	}
}

func TestDecryptPlaintext(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := "not-encrypted-value"

	result, err := Decrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Decrypt should not fail on plaintext: %v", err)
	}

	if result != plaintext {
		t.Errorf("Plaintext should pass through unchanged: got %s, want %s", result, plaintext)
	}
}

func TestEncryptNoKey(t *testing.T) {
	plaintext := "test-value"

	result, err := Encrypt(nil, plaintext)
	if err != nil {
		t.Fatalf("Encrypt with nil key should not error: %v", err)
	}

	if result != plaintext {
		t.Errorf("Plaintext should pass through unchanged when no key: got %s, want %s", result, plaintext)
	}
}

func TestDecryptNoKey(t *testing.T) {
	encrypted := "ENC:someencryptedvalue"

	result, err := Decrypt(nil, encrypted)
	if err != nil {
		t.Fatalf("Decrypt with nil key should not error: %v", err)
	}

	if result != encrypted {
		t.Errorf("Encrypted value should pass through unchanged when no key")
	}
}
