package dto

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ListQuery struct {
	Q     string `form:"q" binding:"omitempty,max=100"`
	Page  int    `form:"page" binding:"omitempty,min=1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

func (q *ListQuery) ApplyDefaults() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	} else if q.Limit > 100 {
		q.Limit = 100
	}
}

