package apputils

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Pagination struct {
	Page    int
	PerPage int
	Offset  int
	Limit   int
}

type PaginationResponse[T any] struct {
	Items *[]T          `json:"items"`
	Meta  *MetaResponse `json:"meta"`
}

type MetaResponse struct {
	Page      int `json:"page"`
	PerPage   int `json:"per_page"`
	Total     int `json:"total"`
	TotalPage int `json:"total_page"`
}

func Paginate(c *fiber.Ctx) *Pagination {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	switch {
	case perPage > 100:
		perPage = 100
	case perPage <= 0:
		perPage = 10
	}

	offset := (page - 1) * perPage

	return &Pagination{
		Page:    page,
		PerPage: perPage,
		Offset:  offset,
		Limit:   perPage,
	}
}

func GetPage(c *fiber.Ctx) int {
	pageStr := c.Query("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}
	return page
}

func GetPageSize(c *fiber.Ctx) int {
	perPageStr := c.Query("per_page", "10")
	perPage, _ := strconv.Atoi(perPageStr)
	switch {
	case perPage > 100:
		perPage = 100
	case perPage <= 0:
		perPage = 10
	}
	return perPage
}

func PaginationBuilder[T any](items []T, meta MetaResponse) *PaginationResponse[T] {
	return &(PaginationResponse[T]{
		Items: &items,
		Meta:  &meta,
	})
}

func PaginationMetaBuilder(c *fiber.Ctx, total int) *MetaResponse {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))

	totalPage := int(math.Ceil(float64(total) / float64(perPage)))

	return &(MetaResponse{
		Page:      page,
		PerPage:   perPage,
		Total:     total,
		TotalPage: totalPage,
	})
}
