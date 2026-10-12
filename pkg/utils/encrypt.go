package utils

import (
	"bytes"
	"fmt"
	"os"

	"filippo.io/age"
)

// LoadRecipients parses an age recipients file (the `age -R` format) and checks the keys can be used together.
func LoadRecipients(path string) ([]age.Recipient, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	recipients, err := age.ParseRecipients(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	// age rejects mixed post-quantum and classic keys only when encrypting: fail before the backup starts
	if _, err := Encrypt(nil, recipients); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return recipients, nil
}

// Encrypt returns data as an age file that any of the recipients can decrypt.
func Encrypt(data []byte, recipients []age.Recipient) ([]byte, error) {
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipients...)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
