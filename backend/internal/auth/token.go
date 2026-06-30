// Package auth จัดการ JWT access token และการสร้าง refresh token
// (cross-cutting — ไม่พึ่ง DB/framework เพื่อให้ test ง่ายและ reuse ได้)
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "chumkosoft"

// Claims คือข้อมูล scope/ตัวตนที่ฝังใน access token
// (พอสำหรับ middleware สร้าง tenant.Identity โดยไม่ต้องแตะ DB ทุก request)
type Claims struct {
	SchoolID      string `json:"school_id"`
	Role          string `json:"role"`
	SemesterID    string `json:"semester_id,omitempty"`
	IsSchoolAdmin bool   `json:"is_school_admin"`
	jwt.RegisteredClaims
}

// TokenManager ออก/ตรวจสอบ access token ด้วย HS256
type TokenManager struct {
	secret    []byte
	accessTTL time.Duration
}

// NewTokenManager สร้าง manager จาก secret และอายุ access token
func NewTokenManager(secret string, accessTTL time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), accessTTL: accessTTL}
}

// AccessTTL คืนอายุของ access token (ใช้ตอบ expires_in)
func (m *TokenManager) AccessTTL() time.Duration { return m.accessTTL }

// Issue สร้าง access token สำหรับ userID พร้อม claims ที่กำหนด
func (m *TokenManager) Issue(userID string, c Claims) (string, error) {
	now := time.Now()
	c.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &c)
	signed, err := tok.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("auth: sign token: %w", err)
	}
	return signed, nil
}

// Parse ตรวจสอบลายเซ็น/อายุ/issuer แล้วคืน claims
func (m *TokenManager) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("auth: unexpected signing method")
		}
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(issuer),
	)
	if err != nil {
		return nil, fmt.Errorf("auth: parse token: %w", err)
	}
	return claims, nil
}

// NewRefreshToken สร้าง opaque refresh token (สุ่ม 32 ไบต์ → hex)
// เก็บฝั่ง server (Redis) เพื่อให้เพิกถอนได้ ต่างจาก access token ที่ stateless
func NewRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: gen refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
