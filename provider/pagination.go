package provider

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Link represents a single pagination link.
type Link struct {
	URL    *string `json:"url"` // Make nullable
	Label  string  `json:"label"`
	Active bool    `json:"active"`
}

// PaginationResponse represents the common pagination data structure.
type PaginationResponse struct {
	CurrentPage  int     `json:"current_page"`
	FirstPageURL *string `json:"first_page_url"` // Make nullable
	From         *int    `json:"from"`           // Make nullable
	LastPage     int     `json:"last_page"`
	LastPageURL  *string `json:"last_page_url"` // Make nullable
	Links        []Link  `json:"links"`
	NextPageURL  *string `json:"next_page_url"` // Make nullable
	Path         string  `json:"path"`
	PerPage      int     `json:"per_page"`
	PrevPageURL  *string `json:"prev_page_url"` // Make nullable
	To           *int    `json:"to"`            // Make nullable
	Total        int     `json:"total"`
}

// GeneratePaginationData generates the pagination response for a list.
//
// `path` should be a relative API path, e.g., "/api/exercise"
// This function will prepend APP_BASE_URL to it.
func GeneratePaginationData(totalCount, page, limit int, path string, queryParams url.Values) PaginationResponse {
	appBaseURL := os.Getenv("APP_BASE_URL") // e.g., https://api2.rtglabs.net
	if appBaseURL == "" {
		appBaseURL = "http://localhost:8080" // Fallback
	}

	// Clean up slashes to avoid malformed URLs
	baseURL := strings.TrimRight(appBaseURL, "/") + "/" + strings.TrimLeft(path, "/")

	// Calculate last page
	lastPage := int(math.Ceil(float64(totalCount) / float64(limit)))
	if lastPage == 0 && totalCount == 0 {
		lastPage = 1 // Laravel-like: if total is 0, there's still 1 page
	}

	// Ensure page is within bounds
	if page < 1 {
		page = 1
	}
	if page > lastPage {
		page = lastPage
	}

	// Helper to build a full URL for a given page, returning *string (nullable)
	buildPageURL := func(p int) *string {
		if p < 1 || p > lastPage {
			return nil
		}
		q := make(url.Values)
		for k, v := range queryParams {
			if k != "page" && k != "limit" {
				q[k] = v
			}
		}
		q.Set("page", strconv.Itoa(p))
		q.Set("limit", strconv.Itoa(limit))

		fullURL := fmt.Sprintf("%s?%s", baseURL, q.Encode())
		return &fullURL
	}

	// --- Calculate From and To ---
	var from, to *int
	if totalCount > 0 {
		tempFrom := (page-1)*limit + 1
		tempTo := page * limit
		if tempTo > totalCount {
			tempTo = totalCount
		}
		from = &tempFrom
		to = &tempTo
	}

	// --- Generate Links ---
	links := make([]Link, 0)

	// Previous
	prevPageURL := buildPageURL(page - 1)
	links = append(links, Link{
		URL:    prevPageURL,
		Label:  "&laquo; Previous",
		Active: false,
	})

	// Pages
	startPage := int(math.Max(1, float64(page-2)))
	endPage := int(math.Min(float64(lastPage), float64(page+2)))

	if endPage-startPage+1 < 5 {
		if startPage == 1 {
			endPage = int(math.Min(float64(lastPage), float64(startPage+4)))
		} else if endPage == lastPage {
			startPage = int(math.Max(1, float64(endPage-4)))
		}
	}

	pagesToShow := make(map[int]bool)
	if lastPage > 0 {
		for i := startPage; i <= endPage; i++ {
			pagesToShow[i] = true
		}
		pagesToShow[1] = true
		pagesToShow[lastPage] = true
	}

	// Sort pages
	var sortedPages []int
	for p := range pagesToShow {
		sortedPages = append(sortedPages, p)
	}
	for i := 0; i < len(sortedPages)-1; i++ {
		for j := i + 1; j < len(sortedPages); j++ {
			if sortedPages[i] > sortedPages[j] {
				sortedPages[i], sortedPages[j] = sortedPages[j], sortedPages[i]
			}
		}
	}

	lastAddedPage := 0
	for _, p := range sortedPages {
		if p > lastAddedPage+1 {
			links = append(links, Link{URL: nil, Label: "...", Active: false})
		}
		links = append(links, Link{
			URL:    buildPageURL(p),
			Label:  strconv.Itoa(p),
			Active: p == page,
		})
		lastAddedPage = p
	}

	// Next
	nextPageURL := buildPageURL(page + 1)
	links = append(links, Link{
		URL:    nextPageURL,
		Label:  "Next &raquo;",
		Active: false,
	})

	// First and Last
	firstPageURL := buildPageURL(1)
	lastPageURL := buildPageURL(lastPage)

	return PaginationResponse{
		CurrentPage:  page,
		FirstPageURL: firstPageURL,
		From:         from,
		LastPage:     lastPage,
		LastPageURL:  lastPageURL,
		Links:        links,
		NextPageURL:  nextPageURL,
		Path:         baseURL, // full path with host
		PerPage:      limit,
		PrevPageURL:  prevPageURL,
		To:           to,
		Total:        totalCount,
	}
}
