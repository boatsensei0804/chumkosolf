// Package face มี logic จับคู่ใบหน้า (cosine similarity) แบบ pure — ทดสอบได้โดยไม่พึ่ง DB/HTTP
package face

import "math"

// DefaultThreshold คือเกณฑ์ cosine ขั้นต่ำที่จะถือว่า "ตรงคน" (ArcFace ปกติ ~0.35–0.5)
const DefaultThreshold = 0.40

// Sample คือ embedding ของรูปหนึ่งใบที่ enroll ไว้ (ผูกกับนักเรียน)
type Sample struct {
	StudentID string
	Vector    []float32
}

// Cosine คืน cosine similarity ของเวกเตอร์สองตัว (0 ถ้าขนาดไม่เท่ากันหรือเป็นศูนย์)
func Cosine(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(na) * math.Sqrt(nb)))
}

// BestMatch หา student ที่ใกล้ที่สุดกับ query ที่ similarity >= threshold
// นักเรียน 1 คนมีได้หลาย embedding (หลายรูป) — ใช้คะแนนสูงสุดต่อคน
func BestMatch(query []float32, samples []Sample, threshold float32) (studentID string, score float32, ok bool) {
	best := float32(-1)
	for i := range samples {
		s := Cosine(query, samples[i].Vector)
		if s > best {
			best = s
			studentID = samples[i].StudentID
		}
	}
	if studentID == "" || best < threshold {
		return "", best, false
	}
	return studentID, best, true
}
