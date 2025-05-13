package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Paging represents pagination parameters
type Paging struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// CalculatePagination calculates the offset and limit values for pagination
// Parameters:
//   - page: The page number (1-based indexing). If less than 1, defaults to 1
//   - limit: Number of items per page. If less than 1, defaults to 10
//
// Returns:
//   - offset: The starting position for the current page ((page-1) * limit)
//   - limit: The number of items per page
func CalculatePagination(page int, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10 // Default limit if not provided or invalid
	}
	offset := (page - 1) * limit
	return offset, limit
}

// GeneratePagingFromRequest extracts pagination parameters from the request
// Parameters:
//   - ctx: The Gin context containing the request
//
// Returns:
//   - *Paging: A pointer to the Paging struct with extracted values
func GeneratePagingFromRequest(ctx *gin.Context) *Paging {
	paging := &Paging{
		Page:  1,
		Limit: 10,
	}

	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			paging.Page = page
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			paging.Limit = limit
		}
	}

	return paging
}

// ApplyPaging applies pagination to a GORM query
// Parameters:
//   - query: The GORM query to paginate
//   - paging: The Paging struct with pagination parameters
//
// Returns:
//   - *gorm.DB: The paginated GORM query
func ApplyPaging(query *gorm.DB, paging *Paging) *gorm.DB {
	// Count total records before applying pagination
	var total int64
	query.Count(&total)
	paging.Total = int(total)

	// Calculate offset
	offset := (paging.Page - 1) * paging.Limit

	// Apply pagination
	return query.Offset(offset).Limit(paging.Limit)
}
