// Package crypto จัดการการเข้ารหัสข้อมูลอ่อนไหวตาม PDPA
// (เลขบัตรประชาชน, เลขบัตรประจำตัวราชการ) — เข้ารหัสตอนเก็บ + hash สำหรับ dedup/ค้นหา
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// Cipher เข้ารหัส/ถอดรหัสข้อมูลอ่อนไหวด้วย AES-256-GCM
// และสร้าง hash (HMAC-SHA256) สำหรับ unique constraint / ค้นหาแบบไม่ต้องถอดรหัส
type Cipher struct {
	gcm     cipher.AEAD
	hmacKey []byte
}

// NewCipher สร้าง Cipher จาก key แบบ hex ขนาด 32 ไบต์ (64 ตัวอักษร)
func NewCipher(hexKey string) (*Cipher, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("crypto: key ต้องเป็น hex: %w", err)
	}
	if len(key) != 32 {
		return nil, errors.New("crypto: ENCRYPTION_KEY ต้องเป็น hex ขนาด 32 ไบต์ (64 ตัวอักษร)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}

	// derive hmac key แยกจาก encryption key เพื่อไม่ใช้ key เดียวสองหน้าที่
	h := sha256.Sum256(append([]byte("hmac:"), key...))

	return &Cipher{gcm: gcm, hmacKey: h[:]}, nil
}

// Encrypt เข้ารหัส plaintext → ciphertext (nonce นำหน้า) สำหรับเก็บลงคอลัมน์ BYTEA
func (c *Cipher) Encrypt(plaintext string) ([]byte, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: nonce: %w", err)
	}
	return c.gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

// Decrypt ถอดรหัส ciphertext กลับเป็น plaintext
func (c *Cipher) Decrypt(ciphertext []byte) (string, error) {
	ns := c.gcm.NonceSize()
	if len(ciphertext) < ns {
		return "", errors.New("crypto: ciphertext สั้นเกินไป")
	}
	nonce, data := ciphertext[:ns], ciphertext[ns:]
	plain, err := c.gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: decrypt: %w", err)
	}
	return string(plain), nil
}

// Hash สร้าง HMAC-SHA256 (hex) สำหรับเก็บลงคอลัมน์ *_hash ใช้ทำ unique/ค้นหา
func (c *Cipher) Hash(value string) string {
	mac := hmac.New(sha256.New, c.hmacKey)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

// Mask ปิดบังเลขบัตรประชาชน 13 หลักเป็นรูปแบบ 1-2345-xxxxx-xx-1
func Mask(nationalID string) string {
	if len(nationalID) != 13 {
		return "xxxxxxxxxxxxx"
	}
	return fmt.Sprintf("%s-%s-xxxxx-xx-%s",
		nationalID[0:1], nationalID[1:5], nationalID[12:13])
}
