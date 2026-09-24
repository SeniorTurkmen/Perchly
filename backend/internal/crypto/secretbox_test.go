package crypto

import (
	"bytes"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key, err := DecodeKey("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=") // "0123456789abcdef0123456789abcdef"
	if err != nil {
		t.Fatalf("decode test key: %v", err)
	}
	return key
}

func TestSecretBoxRoundTrip(t *testing.T) {
	box, err := NewSecretBox(testKey(t))
	if err != nil {
		t.Fatalf("NewSecretBox: %v", err)
	}

	plaintext := []byte("sk-super-secret-api-key")
	ciphertext, nonce, err := box.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatalf("ciphertext must not equal plaintext")
	}

	decrypted, err := box.Decrypt(ciphertext, nonce)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestSecretBoxDecryptFailsOnTamperedCiphertext(t *testing.T) {
	box, err := NewSecretBox(testKey(t))
	if err != nil {
		t.Fatalf("NewSecretBox: %v", err)
	}

	ciphertext, nonce, err := box.Encrypt([]byte("original value"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	ciphertext[0] ^= 0xFF // flip a bit

	if _, err := box.Decrypt(ciphertext, nonce); err == nil {
		t.Fatalf("Decrypt: expected an error for tampered ciphertext, got nil")
	}
}

func TestSecretBoxDecryptFailsOnWrongKey(t *testing.T) {
	box, err := NewSecretBox(testKey(t))
	if err != nil {
		t.Fatalf("NewSecretBox: %v", err)
	}
	ciphertext, nonce, err := box.Encrypt([]byte("original value"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	otherKey, err := DecodeKey("ZmVkY2JhOTg3NjU0MzIxMGZlZGNiYTk4NzY1NDMyMTA=") // different 32-byte key
	if err != nil {
		t.Fatalf("decode other key: %v", err)
	}
	otherBox, err := NewSecretBox(otherKey)
	if err != nil {
		t.Fatalf("NewSecretBox: %v", err)
	}

	if _, err := otherBox.Decrypt(ciphertext, nonce); err == nil {
		t.Fatalf("Decrypt: expected an error when using the wrong key, got nil")
	}
}

func TestNewSecretBoxRejectsWrongKeySize(t *testing.T) {
	if _, err := NewSecretBox([]byte("too short")); err != ErrInvalidKeySize {
		t.Errorf("NewSecretBox with short key: got err %v, want ErrInvalidKeySize", err)
	}
}

func TestDecodeKeyRejectsWrongLength(t *testing.T) {
	if _, err := DecodeKey("c2hvcnQ="); err != ErrInvalidKeySize { // "short"
		t.Errorf("DecodeKey with short key: got err %v, want ErrInvalidKeySize", err)
	}
}
