package domain

// CurrentTerm คือปีการศึกษา + ภาคเรียนปัจจุบันของโรงเรียน (is_current)
type CurrentTerm struct {
	SemesterID   string
	AcademicYear int // ปีการศึกษา พ.ศ. เช่น 2568
	Term         int // ภาคเรียน 1 หรือ 2
}
