package vault

import (
	"crypto/rand"
	"os"
	"path/filepath"
)

// Vault persists local, on-disk state for the CLI (signing key, session token).
type Vault struct {
	dir string
}

func New() (*Vault, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	dir := filepath.Join(home, ".clilogin")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	return &Vault{dir: dir}, nil
}

func (v *Vault) tokenPath() string { return filepath.Join(v.dir, "token") }
func (v *Vault) keyPath() string   { return filepath.Join(v.dir, ".jwtkey") }

// LoadOrCreateSigningKey persists a random HS256 key on first run so tokens
// signed before a restart are still verifiable after it.
// if no signing key file exists yet, generate one and persist it
// if a signing key file already exists, reuse it instead of generating a new one
func (v *Vault) LoadOrCreateSigningKey() ([]byte, error) {
	if b, err := os.ReadFile(v.keyPath()); err == nil && len(b) == 32 {
		return b, nil
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(v.keyPath(), key, 0600); err != nil {
		return nil, err
	}
	return key, nil
}

func (v *Vault) SaveToken(token string) error {
	return os.WriteFile(v.tokenPath(), []byte(token), 0600)
}

// LoadToken from the file.
func (v *Vault) LoadToken() (string, error) {
	b, err := os.ReadFile(v.tokenPath())
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (v *Vault) ClearToken() error {
	err := os.Remove(v.tokenPath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
