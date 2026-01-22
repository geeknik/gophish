package encryption

import (
	"bytes"
	"strings"
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

func TestEncryptDecryptEmptyString(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	encrypted, err := Encrypt(key, "")
	if err != nil {
		t.Fatalf("Failed to encrypt empty string: %v", err)
	}

	if !strings.HasPrefix(encrypted, "ENC:") {
		t.Errorf("Encrypted empty string should still have ENC: prefix")
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt empty string: %v", err)
	}

	if decrypted != "" {
		t.Errorf("Decrypted empty string should be empty, got: %s", decrypted)
	}
}

func TestEncryptDecryptLargeData(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	largeData := strings.Repeat("A", 100000)

	encrypted, err := Encrypt(key, largeData)
	if err != nil {
		t.Fatalf("Failed to encrypt large data: %v", err)
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt large data: %v", err)
	}

	if decrypted != largeData {
		t.Errorf("Decrypted large data doesn't match original")
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	key, _ := GenerateKey()

	_, err := Decrypt(key, "ENC:not-valid-base64!!!")
	if err == nil {
		t.Errorf("Decrypt should fail on invalid base64")
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	key, _ := GenerateKey()

	encrypted, err := Encrypt(key, "secret data")
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	tampered := encrypted[:len(encrypted)-4] + "XXXX"

	_, err = Decrypt(key, tampered)
	if err == nil {
		t.Errorf("Decrypt should fail on tampered ciphertext")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	encrypted, err := Encrypt(key1, "secret data")
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	_, err = Decrypt(key2, encrypted)
	if err == nil {
		t.Errorf("Decrypt should fail with wrong key")
	}
}

func TestDecryptCiphertextTooShort(t *testing.T) {
	key, _ := GenerateKey()

	_, err := Decrypt(key, "ENC:YWJj")
	if err != ErrCipherTextTooShort {
		t.Errorf("Expected ErrCipherTextTooShort, got: %v", err)
	}
}

func TestEncryptInvalidKeySize(t *testing.T) {
	invalidKey := []byte("short")

	_, err := Encrypt(invalidKey, "test")
	if err == nil {
		t.Errorf("Encrypt should fail with invalid key size")
	}
}

func TestGenerateKeyLength(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Generated key should be 32 bytes, got %d", len(key))
	}
}

func TestGenerateKeyUniqueness(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	if bytes.Equal(key1, key2) {
		t.Errorf("Generated keys should be unique")
	}
}

func TestEncryptDifferentCiphertexts(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := "same input"

	encrypted1, _ := Encrypt(key, plaintext)
	encrypted2, _ := Encrypt(key, plaintext)

	if encrypted1 == encrypted2 {
		t.Errorf("Same plaintext should produce different ciphertexts due to random nonce")
	}
}

func TestEncryptSpecialCharacters(t *testing.T) {
	key, _ := GenerateKey()
	special := "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?\nNewlines\tTabs\r\nCRLF"

	encrypted, err := Encrypt(key, special)
	if err != nil {
		t.Fatalf("Failed to encrypt special characters: %v", err)
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt special characters: %v", err)
	}

	if decrypted != special {
		t.Errorf("Special characters not preserved: got %q, want %q", decrypted, special)
	}
}

func TestEncryptUnicode(t *testing.T) {
	key, _ := GenerateKey()
	unicode := "Unicode: 日本語 中文 한국어 العربية 🔐🔑"

	encrypted, err := Encrypt(key, unicode)
	if err != nil {
		t.Fatalf("Failed to encrypt unicode: %v", err)
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt unicode: %v", err)
	}

	if decrypted != unicode {
		t.Errorf("Unicode not preserved: got %q, want %q", decrypted, unicode)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	key, _ := GenerateKey()
	plaintext := strings.Repeat("benchmark data ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encrypt(key, plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key, _ := GenerateKey()
	plaintext := strings.Repeat("benchmark data ", 100)
	encrypted, _ := Encrypt(key, plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decrypt(key, encrypted)
	}
}
