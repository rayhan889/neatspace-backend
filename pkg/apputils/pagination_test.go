package apputils

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginate(t *testing.T) {
	tests := []struct {
		name            string
		queryParams     string
		expectedPage    int
		expectedPerPage int
		expectedOffset  int
		expectedLimit   int
	}{
		{
			name:            "default values when no query params",
			queryParams:     "",
			expectedPage:    1,
			expectedPerPage: 10,
			expectedOffset:  0,
			expectedLimit:   10,
		},
		{
			name:            "valid page and per_page",
			queryParams:     "?page=2&per_page=20",
			expectedPage:    2,
			expectedPerPage: 20,
			expectedOffset:  20,
			expectedLimit:   20,
		},
		{
			name:            "page below 1 defaults to 1",
			queryParams:     "?page=0&per_page=15",
			expectedPage:    1,
			expectedPerPage: 15,
			expectedOffset:  0,
			expectedLimit:   15,
		},
		{
			name:            "negative page defaults to 1",
			queryParams:     "?page=-5&per_page=10",
			expectedPage:    1,
			expectedPerPage: 10,
			expectedOffset:  0,
			expectedLimit:   10,
		},
		{
			name:            "per_page over 100 caps at 100",
			queryParams:     "?page=1&per_page=150",
			expectedPage:    1,
			expectedPerPage: 100,
			expectedOffset:  0,
			expectedLimit:   100,
		},
		{
			name:            "per_page below 1 defaults to 10",
			queryParams:     "?page=3&per_page=0",
			expectedPage:    3,
			expectedPerPage: 10,
			expectedOffset:  20,
			expectedLimit:   10,
		},
		{
			name:            "negative per_page defaults to 10",
			queryParams:     "?page=2&per_page=-5",
			expectedPage:    2,
			expectedPerPage: 10,
			expectedOffset:  10,
			expectedLimit:   10,
		},
		{
			name:            "large page number",
			queryParams:     "?page=100&per_page=25",
			expectedPage:    100,
			expectedPerPage: 25,
			expectedOffset:  2475,
			expectedLimit:   25,
		},
		{
			name:            "invalid page value defaults to 1",
			queryParams:     "?page=abc&per_page=10",
			expectedPage:    1,
			expectedPerPage: 10,
			expectedOffset:  0,
			expectedLimit:   10,
		},
		{
			name:            "invalid per_page value defaults to 10",
			queryParams:     "?page=2&per_page=xyz",
			expectedPage:    2,
			expectedPerPage: 10,
			expectedOffset:  10,
			expectedLimit:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				p := Paginate(c)
				assert.Equal(t, tt.expectedPage, p.Page)
				assert.Equal(t, tt.expectedPerPage, p.PerPage)
				assert.Equal(t, tt.expectedOffset, p.Offset)
				assert.Equal(t, tt.expectedLimit, p.Limit)
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test"+tt.queryParams, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPage(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  string
		expectedPage int
	}{
		{
			name:         "default page is 1",
			queryParams:  "",
			expectedPage: 1,
		},
		{
			name:         "valid page number",
			queryParams:  "?page=5",
			expectedPage: 5,
		},
		{
			name:         "zero page defaults to 1",
			queryParams:  "?page=0",
			expectedPage: 1,
		},
		{
			name:         "negative page defaults to 1",
			queryParams:  "?page=-10",
			expectedPage: 1,
		},
		{
			name:         "invalid page value defaults to 1",
			queryParams:  "?page=invalid",
			expectedPage: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				page := GetPage(c)
				assert.Equal(t, tt.expectedPage, page)
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test"+tt.queryParams, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPageSize(t *testing.T) {
	tests := []struct {
		name             string
		queryParams      string
		expectedPageSize int
	}{
		{
			name:             "default page size is 10",
			queryParams:      "",
			expectedPageSize: 10,
		},
		{
			name:             "valid page size",
			queryParams:      "?per_page=25",
			expectedPageSize: 25,
		},
		{
			name:             "page size over 100 caps at 100",
			queryParams:      "?per_page=200",
			expectedPageSize: 100,
		},
		{
			name:             "zero page size defaults to 10",
			queryParams:      "?per_page=0",
			expectedPageSize: 10,
		},
		{
			name:             "negative page size defaults to 10",
			queryParams:      "?per_page=-5",
			expectedPageSize: 10,
		},
		{
			name:             "invalid page size defaults to 10",
			queryParams:      "?per_page=abc",
			expectedPageSize: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				pageSize := GetPageSize(c)
				assert.Equal(t, tt.expectedPageSize, pageSize)
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test"+tt.queryParams, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode)
		})
	}
}

func TestPaginationBuilder(t *testing.T) {
	t.Run("build pagination response with items and meta", func(t *testing.T) {
		items := []string{"item1", "item2", "item3"}
		meta := MetaResponse{
			Page:      1,
			PerPage:   10,
			Total:     3,
			TotalPage: 1,
		}

		result := PaginationBuilder(items, meta)

		require.NotNil(t, result)
		require.NotNil(t, result.Items)
		require.NotNil(t, result.Meta)
		assert.Equal(t, &items, result.Items)
		assert.Equal(t, 3, len(*result.Items))
		assert.Equal(t, &meta, result.Meta)
		assert.Equal(t, 1, result.Meta.Page)
		assert.Equal(t, 10, result.Meta.PerPage)
		assert.Equal(t, 3, result.Meta.Total)
		assert.Equal(t, 1, result.Meta.TotalPage)
	})

	t.Run("build pagination response with empty items", func(t *testing.T) {
		items := []int{}
		meta := MetaResponse{
			Page:      1,
			PerPage:   10,
			Total:     0,
			TotalPage: 0,
		}

		result := PaginationBuilder(items, meta)

		require.NotNil(t, result)
		require.NotNil(t, result.Items)
		assert.Equal(t, 0, len(*result.Items))
		assert.Equal(t, 0, result.Meta.Total)
		assert.Equal(t, 0, result.Meta.TotalPage)
	})

	t.Run("build pagination response with struct items", func(t *testing.T) {
		type User struct {
			ID   int
			Name string
		}
		items := []User{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
		}
		meta := MetaResponse{
			Page:      2,
			PerPage:   2,
			Total:     10,
			TotalPage: 5,
		}

		result := PaginationBuilder(items, meta)

		require.NotNil(t, result)
		assert.Equal(t, 2, len(*result.Items))
		assert.Equal(t, "Alice", (*result.Items)[0].Name)
		assert.Equal(t, 2, result.Meta.Page)
		assert.Equal(t, 5, result.Meta.TotalPage)
	})
}

func TestPaginationMetaBuilder(t *testing.T) {
	tests := []struct {
		name              string
		queryParams       string
		total             int
		expectedPage      int
		expectedPerPage   int
		expectedTotal     int
		expectedTotalPage int
	}{
		{
			name:              "default pagination with 25 total items",
			queryParams:       "",
			total:             25,
			expectedPage:      1,
			expectedPerPage:   10,
			expectedTotal:     25,
			expectedTotalPage: 3,
		},
		{
			name:              "page 2 with 100 items",
			queryParams:       "?page=2&per_page=20",
			total:             100,
			expectedPage:      2,
			expectedPerPage:   20,
			expectedTotal:     100,
			expectedTotalPage: 5,
		},
		{
			name:              "total items less than per_page",
			queryParams:       "?page=1&per_page=10",
			total:             5,
			expectedPage:      1,
			expectedPerPage:   10,
			expectedTotal:     5,
			expectedTotalPage: 1,
		},
		{
			name:              "exact multiple of per_page",
			queryParams:       "?page=1&per_page=10",
			total:             50,
			expectedPage:      1,
			expectedPerPage:   10,
			expectedTotal:     50,
			expectedTotalPage: 5,
		},
		{
			name:              "zero total items",
			queryParams:       "?page=1&per_page=10",
			total:             0,
			expectedPage:      1,
			expectedPerPage:   10,
			expectedTotal:     0,
			expectedTotalPage: 0,
		},
		{
			name:              "large dataset with custom page size",
			queryParams:       "?page=5&per_page=50",
			total:             1000,
			expectedPage:      5,
			expectedPerPage:   50,
			expectedTotal:     1000,
			expectedTotalPage: 20,
		},
		{
			name:              "total not evenly divisible by per_page",
			queryParams:       "?page=1&per_page=7",
			total:             23,
			expectedPage:      1,
			expectedPerPage:   7,
			expectedTotal:     23,
			expectedTotalPage: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				meta := PaginationMetaBuilder(c, tt.total)

				assert.Equal(t, tt.expectedPage, meta.Page)
				assert.Equal(t, tt.expectedPerPage, meta.PerPage)
				assert.Equal(t, tt.expectedTotal, meta.Total)
				assert.Equal(t, tt.expectedTotalPage, meta.TotalPage)

				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test"+tt.queryParams, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode)
			defer resp.Body.Close()
			_, _ = io.ReadAll(resp.Body)
		})
	}
}

func TestPaginationIntegration(t *testing.T) {
	t.Run("full pagination flow with all functions", func(t *testing.T) {
		app := fiber.New()
		app.Get("/users", func(c *fiber.Ctx) error {
			// Simulate user data
			allUsers := []string{
				"user1", "user2", "user3", "user4", "user5",
				"user6", "user7", "user8", "user9", "user10",
				"user11", "user12", "user13", "user14", "user15",
			}

			// Use Paginate
			p := Paginate(c)

			// Simulate fetching paginated data
			start := p.Offset
			end := start + p.Limit
			if end > len(allUsers) {
				end = len(allUsers)
			}

			var items []string
			if start < len(allUsers) {
				items = allUsers[start:end]
			} else {
				items = []string{}
			}

			// Build meta
			meta := PaginationMetaBuilder(c, len(allUsers))

			// Build response
			response := PaginationBuilder(items, *meta)

			return c.JSON(response)
		})

		// Test page 1
		req := httptest.NewRequest("GET", "/users?page=1&per_page=5", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Test page 2
		req = httptest.NewRequest("GET", "/users?page=2&per_page=5", nil)
		resp, err = app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Test out of range page
		req = httptest.NewRequest("GET", "/users?page=100&per_page=5", nil)
		resp, err = app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}
