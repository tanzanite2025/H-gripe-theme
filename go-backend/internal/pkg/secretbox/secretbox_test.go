package secretbox

import "testing"

func TestEncryptStringRoundTrip(t *testing.T) {
	encrypted, err := EncryptString("refresh-token-value", "test-master-key")
	if err != nil {
		t.Fatalf("EncryptString() error = %v", err)
	}
	if encrypted == "" || encrypted == "refresh-token-value" {
		t.Fatalf("encrypted value exposed plaintext: %q", encrypted)
	}

	decrypted, err := DecryptString(encrypted, "test-master-key")
	if err != nil {
		t.Fatalf("DecryptString() error = %v", err)
	}
	if decrypted != "refresh-token-value" {
		t.Fatalf("DecryptString() = %q, want refresh-token-value", decrypted)
	}
}

func TestDecryptStringRejectsWrongKey(t *testing.T) {
	encrypted, err := EncryptString("refresh-token-value", "test-master-key")
	if err != nil {
		t.Fatalf("EncryptString() error = %v", err)
	}
	if _, err := DecryptString(encrypted, "wrong-key"); err == nil {
		t.Fatal("DecryptString() should reject a wrong key")
	}
}
