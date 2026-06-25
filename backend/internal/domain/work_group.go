package domain

// WorkGroup คือกลุ่มงานของโรงเรียน (personnel, general_affairs, academic, budget_plan)
type WorkGroup struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
