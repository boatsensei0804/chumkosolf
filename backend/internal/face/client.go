package face

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
)

// Client เรียก face-svc (Python) เพื่อคำนวณ embedding ใบหน้า — stateless
type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 30 * time.Second}}
}

// Enabled คืน true ถ้าตั้งค่า face-svc ไว้
func (c *Client) Enabled() bool { return c.baseURL != "" }

type embedResponse struct {
	Embedding []float32 `json:"embedding"`
	Faces     int       `json:"faces"`
	Yaw       float32   `json:"yaw"`
}

// EmbedResult = embedding + สัญญาณ yaw (สำหรับตรวจ liveness จากการขยับหน้า)
type EmbedResult struct {
	Embedding []float32
	Yaw       float32
}

// Embed ส่งรูปไป face-svc แล้วคืน embedding (512 มิติ, L2-normalized) + yaw
// คืน domain.ErrNoFaceDetected ถ้าไม่พบใบหน้า, domain.ErrFaceServiceUnavailable ถ้าต่อ service ไม่ได้
func (c *Client) Embed(ctx context.Context, image []byte) (EmbedResult, error) {
	if !c.Enabled() {
		return EmbedResult{}, domain.ErrFaceServiceUnavailable
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "frame.jpg")
	if err != nil {
		return EmbedResult{}, fmt.Errorf("face: create form: %w", err)
	}
	if _, err := part.Write(image); err != nil {
		return EmbedResult{}, fmt.Errorf("face: write image: %w", err)
	}
	_ = w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embed", &body)
	if err != nil {
		return EmbedResult{}, fmt.Errorf("face: new request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return EmbedResult{}, domain.ErrFaceServiceUnavailable // ต่อ service ไม่ได้
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		var out embedResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return EmbedResult{}, fmt.Errorf("face: decode response: %w", err)
		}
		if len(out.Embedding) == 0 {
			return EmbedResult{}, domain.ErrNoFaceDetected
		}
		return EmbedResult{Embedding: out.Embedding, Yaw: out.Yaw}, nil
	case http.StatusUnprocessableEntity, http.StatusBadRequest:
		// 422 = ไม่พบใบหน้า, 400 = ภาพเสีย/อ่านไม่ได้ — สำหรับ kiosk ถือว่า "เฟรมนี้ใช้ไม่ได้" เหมือนกัน
		return EmbedResult{}, domain.ErrNoFaceDetected
	case http.StatusServiceUnavailable:
		return EmbedResult{}, domain.ErrFaceServiceUnavailable
	default:
		return EmbedResult{}, fmt.Errorf("face: service status %d", resp.StatusCode)
	}
}
