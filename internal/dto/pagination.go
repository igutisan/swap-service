package dto

// PaginationMeta contiene los metadatos de paginación
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	PerPage     int   `json:"per_page"`
	TotalItems  int64 `json:"total_items"`
	HasNext     bool  `json:"has_next"`
	HasPrev     bool  `json:"has_prev"`
	NextPage    int   `json:"next_page,omitempty"`
	PrevPage    int   `json:"prev_page,omitempty"`
}

// PaginationRequest representa los parámetros de paginación en las requests
type PaginationRequest struct {
	Page    int `form:"page" json:"page" validate:"omitempty,min=1"`
	PerPage int `form:"per_page" json:"per_page" validate:"omitempty,min=1,max=100"`
}

// DefaultPagination retorna los valores por defecto de paginación
func DefaultPagination() PaginationRequest {
	return PaginationRequest{
		Page:    1,
		PerPage: 20,
	}
}

// GetOffset calcula el offset para la base de datos
func (p PaginationRequest) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PerPage <= 0 {
		p.PerPage = 20
	}
	return (p.Page - 1) * p.PerPage
}

// GetLimit retorna el límite para la base de datos
func (p PaginationRequest) GetLimit() int {
	if p.PerPage <= 0 {
		return 20
	}
	if p.PerPage > 100 {
		return 100
	}
	return p.PerPage
}

// PaginatedData es una estructura genérica para datos paginados
type PaginatedData struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// NewPaginationMeta crea metadatos de paginación
func NewPaginationMeta(currentPage, perPage int, totalItems int64) PaginationMeta {
	totalPages := int((totalItems + int64(perPage) - 1) / int64(perPage))
	if totalPages < 1 {
		totalPages = 1
	}

	meta := PaginationMeta{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		PerPage:     perPage,
		TotalItems:  totalItems,
		HasNext:     currentPage < totalPages,
		HasPrev:     currentPage > 1,
	}

	if meta.HasNext {
		meta.NextPage = currentPage + 1
	}
	if meta.HasPrev {
		meta.PrevPage = currentPage - 1
	}

	return meta
}
