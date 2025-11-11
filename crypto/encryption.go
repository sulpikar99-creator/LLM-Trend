package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

// Service provides encryption and decryption services
type Service struct {
	masterKey []byte
	rsaKey    *rsa.PrivateKey
}

// NewService creates a new crypto service
func NewService() (*Service, error) {
	// Get master key from environment
	keyHex := os.Getenv("DATA_ENCRYPTION_KEY")
	if keyHex == "" {
		return nil, fmt.Errorf("DATA_ENCRYPTION_KEY environment variable not set")
	}

	// Decode hex key
	masterKey, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode master key: %w", err)
	}

	if len(masterKey) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes (256 bits), got %d bytes", len(masterKey))
	}

	// Load or generate RSA key
	rsaKey, err := loadOrGenerateRSAKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load RSA key: %w", err)
	}

	return &Service{
		masterKey: masterKey,
		rsaKey:    rsaKey,
	}, nil
}

// EncryptAES encrypts data using AES-256-GCM
func (s *Service) EncryptAES(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES decrypts data using AES-256-GCM
func (s *Service) DecryptAES(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptRSA encrypts data using RSA
func (s *Service) EncryptRSA(plaintext string) (string, error) {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, &s.rsaKey.PublicKey, []byte(plaintext), nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with RSA: %w", err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptRSA decrypts data using RSA
func (s *Service) DecryptRSA(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, s.rsaKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt with RSA: %w", err)
	}

	return string(plaintext), nil
}

// loadOrGenerateRSAKey loads existing RSA key or generates a new one
func loadOrGenerateRSAKey() (*rsa.PrivateKey, error) {
	keyPath := "rsa_private_key.pem"

	// Try to load existing key
	if fileInfo, err := os.Stat(keyPath); err == nil {
		// Check if it's actually a file and not a directory
		if fileInfo.IsDir() {
			return nil, fmt.Errorf("RSA key path is a directory, not a file")
		}

		// Check if file is empty
		if fileInfo.Size() == 0 {
			fmt.Println("Warning: RSA key file is empty, regenerating...")
			os.Remove(keyPath)
		} else {
			// Try to read the key
			keyData, err := os.ReadFile(keyPath)
			if err != nil {
				// If reading fails, try to delete and regenerate
				fmt.Printf("Warning: Failed to read RSA key (%v), regenerating...\n", err)
				os.Remove(keyPath)
			} else {
				block, _ := pem.Decode(keyData)
				if block == nil {
					fmt.Println("Warning: Failed to decode PEM block, regenerating...")
					os.Remove(keyPath)
				} else {
					privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
					if err != nil {
						fmt.Printf("Warning: Failed to parse RSA key (%v), regenerating...\n", err)
						os.Remove(keyPath)
					} else {
						return privateKey, nil
					}
				}
			}
		}
	}

	// Generate new key
	fmt.Println("Generating new RSA key...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Save key to file
	keyData := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyData,
	}

	// Create file with explicit permissions
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyFile.Close()

	if err := pem.Encode(keyFile, block); err != nil {
		return nil, fmt.Errorf("failed to write RSA key: %w", err)
	}

	fmt.Println("RSA key generated successfully")
	return privateKey, nil
}
