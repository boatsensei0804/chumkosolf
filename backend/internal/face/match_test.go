package face

import "testing"

func TestCosine(t *testing.T) {
	if got := Cosine([]float32{1, 0}, []float32{1, 0}); got < 0.999 {
		t.Errorf("เวกเตอร์เดียวกัน cosine = %v, want ~1", got)
	}
	if got := Cosine([]float32{1, 0}, []float32{0, 1}); got > 0.001 || got < -0.001 {
		t.Errorf("ตั้งฉาก cosine = %v, want ~0", got)
	}
	if got := Cosine([]float32{1, 2, 3}, []float32{1, 2}); got != 0 {
		t.Errorf("ขนาดไม่เท่ากัน ต้องคืน 0, ได้ %v", got)
	}
}

func TestBestMatch_PicksHighestAboveThreshold(t *testing.T) {
	samples := []Sample{
		{StudentID: "s1", Vector: []float32{1, 0, 0}},
		{StudentID: "s2", Vector: []float32{0, 1, 0}},
		{StudentID: "s1", Vector: []float32{0.9, 0.1, 0}}, // s1 มีหลายรูป
	}
	id, score, ok := BestMatch([]float32{1, 0, 0}, samples, 0.4)
	if !ok || id != "s1" {
		t.Fatalf("match = %q ok=%v (score %v), want s1", id, ok, score)
	}
}

func TestBestMatch_BelowThreshold(t *testing.T) {
	samples := []Sample{{StudentID: "s1", Vector: []float32{0, 1, 0}}}
	_, _, ok := BestMatch([]float32{1, 0, 0}, samples, 0.4)
	if ok {
		t.Error("คนละทิศ (cosine 0) ต้องไม่ match")
	}
}

func TestBestMatch_Empty(t *testing.T) {
	if _, _, ok := BestMatch([]float32{1, 0}, nil, 0.4); ok {
		t.Error("ไม่มี sample ต้องไม่ match")
	}
}
