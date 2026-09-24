package pagination

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 10
	MaxPerPage     = 100
)

type Query struct {
	Page    int
	PerPage int
	Search  string
}

func FromContext(c *gin.Context) Query {
	page := parsePositive(c.Query("page"), DefaultPage)
	perPage := parsePositive(c.Query("per_page"), DefaultPerPage)
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	return Query{Page: page, PerPage: perPage, Search: strings.TrimSpace(c.Query("search"))}
}

func (query Query) Offset() int {
	return (query.Page - 1) * query.PerPage
}

type Result[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewResult[T any](items []T, query Query, total int) Result[T] {
	totalPages := 0
	if total > 0 {
		totalPages = (total + query.PerPage - 1) / query.PerPage
	}
	return Result[T]{Items: items, Page: query.Page, PerPage: query.PerPage, Total: total, TotalPages: totalPages}
}

func parsePositive(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
