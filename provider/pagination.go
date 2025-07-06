package provider

import (
	"fmt"
	"math" // Needed for math.Ceil
	"net/url"
	"strconv"
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
// `baseURL` should be the full base URL including scheme and host (e.g., "http://localhost:8000/api/workouts")
func GeneratePaginationData(totalCount, page, limit int, baseURL string, queryParams url.Values) PaginationResponse {
	// Calculate last page
	lastPage := int(math.Ceil(float64(totalCount) / float64(limit)))
	if lastPage == 0 && totalCount == 0 {
		lastPage = 1 // Laravel-like: if total is 0, there's still 1 page
	}

	// Ensure page is within valid bounds after calculating lastPage
	if page < 1 {
		page = 1
	}
	if page > lastPage {
		page = lastPage
	}

	// Helper to build a full URL for a given page, returning *string (nullable)
	buildPageURL := func(p int) *string {
		if p < 1 || p > lastPage {
			return nil // Return nil if page is out of bounds
		}
		q := make(url.Values)
		for k, v := range queryParams {
			// Don't include 'page' and 'limit' from existing queryParams,
			// as they will be explicitly set. Other filters should be carried over.
			if k != "page" && k != "limit" {
				q[k] = v
			}
		}
		q.Set("page", strconv.Itoa(p))
		q.Set("limit", strconv.Itoa(limit)) // Ensure limit is always in the URL

		// Construct the URL without the baseURL's path, then add query
		// If baseURL already contains the path, just append query
		fullURL := fmt.Sprintf("%s?%s", baseURL, q.Encode())
		return &fullURL
	}

	// --- Calculate From and To (Pointers for nullability) ---
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
	// If totalCount is 0, from and to remain nil, which serializes to null.

	// --- Generate Links ---
	links := make([]Link, 0)

	// Add "Previous" link
	prevPageURL := buildPageURL(page - 1)
	links = append(links, Link{
		URL:    prevPageURL,
		Label:  "&laquo; Previous",
		Active: false,
	})

	// Add individual page links
	// Laravel usually shows a limited set of page numbers, e.g., current, 2 before, 2 after.
	// For exact Laravel replica, this logic can be more complex (ellipsis, etc.).
	// For now, let's show all pages or a reasonable subset (e.g., up to 5 surrounding pages).
	// Let's implement a basic range: show current page, and 2 pages before/after if available.
	startPage := int(math.Max(1, float64(page-2)))
	endPage := int(math.Min(float64(lastPage), float64(page+2)))

	// Adjust range if it's too small at the ends
	if endPage-startPage+1 < 5 { // If fewer than 5 pages in current window
		if startPage == 1 { // If at the beginning, expand end
			endPage = int(math.Min(float64(lastPage), float64(startPage+4)))
		} else if endPage == lastPage { // If at the end, expand start
			startPage = int(math.Max(1, float64(endPage-4)))
		}
	}

	// Always show at least page 1 and last page if they are outside the range
	pagesToShow := make(map[int]bool)
	if lastPage > 0 { // Ensure lastPage is at least 1
		for i := startPage; i <= endPage; i++ {
			pagesToShow[i] = true
		}
		pagesToShow[1] = true        // Always include page 1
		pagesToShow[lastPage] = true // Always include last page
	}

	// Sort the pages to display
	var sortedPages []int
	for p := range pagesToShow {
		sortedPages = append(sortedPages, p)
	}
	// Sort them numerically
	for i := 0; i < len(sortedPages)-1; i++ {
		for j := i + 1; j < len(sortedPages); j++ {
			if sortedPages[i] > sortedPages[j] {
				sortedPages[i], sortedPages[j] = sortedPages[j], sortedPages[i]
			}
		}
	}

	// Add page links, including ellipsis logic if desired (simplified for now)
	lastAddedPage := 0
	for _, p := range sortedPages {
		if p > lastAddedPage+1 { // Add ellipsis if there's a gap
			links = append(links, Link{URL: nil, Label: "...", Active: false})
		}
		links = append(links, Link{
			URL:    buildPageURL(p),
			Label:  strconv.Itoa(p),
			Active: p == page,
		})
		lastAddedPage = p
	}

	// Add "Next" link
	nextPageURL := buildPageURL(page + 1)
	links = append(links, Link{
		URL:    nextPageURL,
		Label:  "Next &raquo;",
		Active: false,
	})

	// Set FirstPageURL and LastPageURL
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
		Path:         baseURL, // Path is the base URL without query params
		PerPage:      limit,
		PrevPageURL:  prevPageURL,
		To:           to,
		Total:        totalCount,
	}
}
