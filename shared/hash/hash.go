package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64MB
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

func HashSecret(secret string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("falha ao gerar salt: %w", err)
	}

	hash := argon2.IDKey([]byte(secret), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	b64Salt  := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash  := base64.RawStdEncoding.EncodeToString(hash)
	encodedString := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argon2Memory, argon2Time, argon2Threads, b64Salt, b64Hash)
	
	return encodedString, nil
}

func VerifyArgon2Match(secret, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("formato de hash inválido")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil || version != argon2.Version {
		return false, errors.New("versão do argon2 incompatível")
	}

	var memory, time, threads uint32
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, errors.New("parâmetros do argon2 inválidos")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	actualHash := argon2.IDKey([]byte(secret), salt, time, memory, uint8(threads), uint32(len(expectedHash)))

	if subtle.ConstantTimeCompare(actualHash, expectedHash) == 1 {
		return true, nil
	}

	return false, nil
}