package pagination

import (
	"math"
)

// PageInfo contém informações sobre a paginação
type PageInfo struct {
	CurrentPage  int   `json:"current_page"`
	PerPage      int   `json:"per_page"`
	TotalItems   int64 `json:"total_items"`
	TotalPages   int   `json:"total_pages"`
	HasPrevious  bool  `json:"has_previous"`
	HasNext      bool  `json:"has_next"`
	PreviousPage int   `json:"previous_page"`
	NextPage     int   `json:"next_page"`
}

// Paginator gerencia a paginação de resultados
type Paginator struct {
	page    int
	perPage int
	total   int64
}

// NewPaginator cria uma nova instância do Paginator
func NewPaginator(page, perPage int) *Paginator {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	
	return &Paginator{
		page:    page,
		perPage: perPage,
	}
}

// SetTotal define o total de itens
func (p *Paginator) SetTotal(total int64) {
	p.total = total
}

// Offset retorna o offset para a query
func (p *Paginator) Offset() int {
	return (p.page - 1) * p.perPage
}

// Limit retorna o limite para a query
func (p *Paginator) Limit() int {
	return p.perPage
}

// GetInfo retorna informações sobre a paginação atual
func (p *Paginator) GetInfo() PageInfo {
	totalPages := int(math.Ceil(float64(p.total) / float64(p.perPage)))
	
	return PageInfo{
		CurrentPage:  p.page,
		PerPage:      p.perPage,
		TotalItems:   p.total,
		TotalPages:   totalPages,
		HasPrevious:  p.page > 1,
		HasNext:      p.page < totalPages,
		PreviousPage: max(1, p.page-1),
		NextPage:     min(totalPages, p.page+1),
	}
} 