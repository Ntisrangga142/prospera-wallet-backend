package pkg

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type HashConfig struct {
	Memory  uint32
	Time    uint32
	Thread  uint8
	KeyLen  uint32
	SaltLen uint32
}

func NewHashConfig() *HashConfig {
	return &HashConfig{}
}

func (h *HashConfig) UseRecommended() {
	h.KeyLen = 32
	h.SaltLen = 16
	h.Memory = 64 * 1024
	h.Time = 2
	h.Thread = 1
}

func (h *HashConfig) GenHash(password string) (string, error) {
	salt, err := h.genSalt()
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.Time,
		h.Memory,
		h.Thread,
		h.KeyLen,
	)
	version := argon2.Version

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", version, h.Memory, h.Time, h.Thread, b64Salt, b64Hash,
	)

	return encodedHash, nil
}

func (h *HashConfig) genSalt() ([]byte, error) {
	salt := make([]byte, h.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	return salt, nil
}

func (h *HashConfig) ComparePasswordAndHash(password, encodedHash string) (bool, error) {
	salt, hash, err := h.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, h.Time, h.Memory, h.Thread, h.KeyLen)

	if subtle.ConstantTimeCompare(hash, otherHash) == 0 {
		return false, nil
	}

	return true, nil
}

func (h *HashConfig) decodeHash(encodedHash string) (salt []byte, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, err
	}
	if vals[1] != "argon2id" {
		return nil, nil, err
	}

	var version int
	if _, err := fmt.Sscanf(vals[2], "v=%d", &version); err != nil {
		return nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, err
	}

	if _, err := fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &h.Memory, &h.Time, &h.Thread); err != nil {
		return nil, nil, err
	}

	getSalt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, nil, err
	}
	h.SaltLen = uint32(len(getSalt))

	getHash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, nil, err
	}
	h.KeyLen = uint32(len(getHash))

	return getSalt, getHash, err
}
