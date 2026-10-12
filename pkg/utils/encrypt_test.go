package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
)

func writeRecipients(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "recipients.txt")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestEncryptRoundTrip(t *testing.T) {
	classic, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	classic2, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	pq, err := age.GenerateHybridIdentity()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		recipients string
		identities []age.Identity
	}{
		{"classic", classic.Recipient().String() + "\n", []age.Identity{classic}},
		{"post-quantum", pq.Recipient().String() + "\n", []age.Identity{pq}},
		{"two recipients, comments, second key decrypts",
			"# admin\n" + classic.Recipient().String() + "\n\n# break-glass\n" + classic2.Recipient().String() + "\n",
			[]age.Identity{classic2}},
	}

	plain := []byte("apiVersion: v1\nkind: Secret\nmetadata:\n  name: db\ndata:\n  password: c2VjcmV0\n")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipients, err := LoadRecipients(writeRecipients(t, tt.recipients))
			if err != nil {
				t.Fatalf("LoadRecipients: %v", err)
			}
			enc, err := Encrypt(plain, recipients)
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}
			if bytes.Contains(enc, []byte("c2VjcmV0")) || bytes.Contains(enc, []byte("kind: Secret")) {
				t.Fatal("ciphertext contains plain text")
			}
			r, err := age.Decrypt(bytes.NewReader(enc), tt.identities...)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}
			got, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, plain) {
				t.Fatalf("round trip mismatch: got %q", got)
			}
		})
	}
}

func TestLoadRecipientsRejectsUnusableFiles(t *testing.T) {
	classic, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	pq, err := age.GenerateHybridIdentity()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		content string
	}{
		{"empty", ""},
		{"only comments", "# nothing here\n\n"},
		{"invalid key", "age1notakey\n"},
		{"private key by mistake", classic.String() + "\n"},
		{"ssh key", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIH9pO5pz22JZEas user@host\n"},
		{"mixed post-quantum and classic", pq.Recipient().String() + "\n" + classic.Recipient().String() + "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadRecipients(writeRecipients(t, tt.content)); err == nil {
				t.Fatal("expected an error")
			} else if strings.Contains(err.Error(), "AGE-SECRET-KEY") {
				t.Fatalf("error leaks the private key: %v", err)
			}
		})
	}

	t.Run("missing file", func(t *testing.T) {
		if _, err := LoadRecipients(filepath.Join(t.TempDir(), "nope.txt")); err == nil {
			t.Fatal("expected an error")
		}
	})
}
