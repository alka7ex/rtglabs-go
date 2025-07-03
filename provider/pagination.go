package provider

import (
	"fmt"
	"net/url"
	"strconv"
)

// Link represents a single pagination link.
type Link struct {
	URL    *string `json:"url"`
	Label  string  `json:"label"`
	Active bool    `json:"active"`
}

// PaginationResponse represents the common pagination data structure.
type PaginationResponse struct {
	CurrentPage  int     `json:"current_page"`
	FirstPageURL string  `json:"first_page_url"`
	From         *int    `json:"from"`
	LastPage     int     `json:"last_page"`
	LastPageURL  string  `json:"last_page_url"`
	Links        []Link  `json:"links"`
	NextPageURL  *string `json:"next_page_url"`
	Path         string  `json:"path"`
	PerPage      int     `json:"per_page"`
	PrevPageURL  *string `json:"prev_page_url"`
	To           *int    `json:"to"`
	Total        int     `json:"total"`
}

// GeneratePaginationData generates the pagination response for a list.
func GeneratePaginationData(totalCount, page, limit int, baseURL string, queryParams url.Values) PaginationResponse {
	lastPage := (totalCount + limit - 1) / limit
	offset := (page - 1) * limit
	to := offset + len(queryParams) // This 'to' needs to be calculated based on the actual number of items on the current page, not just offset.
	if offset+limit < totalCount {
		tempTo := offset + limit
		to = tempTo
	} else {
		to = totalCount
	}

	buildPageURL := func(p int) string {
		q := make(url.Values)
		for k, v := range queryParams {
			q[k] = v
		}
		q.Set("page", strconv.Itoa(p))
		q.Set("limit", strconv.Itoa(limit))
		return fmt.Sprintf("%s?%s", baseURL, q.Encode())
	}

	var links []Link
	for i := 1; i <= lastPage; i++ {
		url := buildPageURL(i)
		links = append(links, Link{
			URL:    &url,
			Label:  strconv.Itoa(i),
			Active: i == page,
		})
	}

	var prevURL, nextURL *string
	if page > 1 {
		url := buildPageURL(page - 1)
		prevURL = &url
	}
	if page < lastPage {
		url := buildPageURL(page + 1)
		nextURL = &url
	}

	// Adjust 'From' and 'To' based on actual data
	from := 0
	if totalCount > 0 {
		from = offset + 1
	}

	return PaginationResponse{
		CurrentPage:  page,
		FirstPageURL: buildPageURL(1),
		From:         &from,
		LastPage:     lastPage,
		LastPageURL:  buildPageURL(lastPage),
		Links:        links,
		NextPageURL:  nextURL,
		Path:         baseURL,
		PerPage:      limit,
		PrevPageURL:  prevURL,
		To:           &to,
		Total:        totalCount,
	}
}
