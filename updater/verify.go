package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

func Verify(path, expected string) error {
	f, _ := os.Open(path)
	defer f.Close()

	h := sha256.New()
	io.Copy(h, f)

	sum := hex.EncodeToString(h.Sum(nil))
	if sum != expected {
		return errors.New("checksum mismatch")
	}
	return nil
}
