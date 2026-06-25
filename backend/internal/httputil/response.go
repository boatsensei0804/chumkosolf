// Package httputil มี helper สำหรับ response format มาตรฐานของทั้งระบบ
// รูปแบบ: { "success": bool, "data": any, "error": {code,message}, "meta": {...} }
package httputil

import "github.com/gofiber/fiber/v2"

// APIResponse คือ envelope มาตรฐานของทุก endpoint
type APIResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data"`
	Error   *APIError `json:"error"`
	Meta    *Meta     `json:"meta,omitempty"`
}

// APIError มี code (machine-readable) + message ภาษาไทย (ผู้ใช้เข้าใจได้)
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta สำหรับข้อมูล pagination ฯลฯ
type Meta struct {
	Page  int `json:"page,omitempty"`
	Total int `json:"total,omitempty"`
}

// OK ตอบ 200 พร้อม data
func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{Success: true, Data: data})
}

// Created ตอบ 201 พร้อม data
func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{Success: true, Data: data})
}

// OKWithMeta ตอบ 200 พร้อม data + meta (เช่น list ที่มี pagination)
func OKWithMeta(c *fiber.Ctx, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{Success: true, Data: data, Meta: meta})
}

// Error ตอบ error ด้วย status, code, ข้อความไทย
func Error(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(APIResponse{
		Success: false,
		Error:   &APIError{Code: code, Message: message},
	})
}
