package fetch

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/myzie/web/errors"
)

// ParsePostRequest parses a fetch.Request from a POST request body.
func ParsePostRequest(r *http.Request) (*Request, error) {
	var requestBody Request
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		return nil, errors.NewBadRequest("invalid json body")
	}
	if requestBody.URL == "" {
		return nil, errors.NewBadRequest("url is required")
	}
	return &requestBody, nil
}

// ParseGetRequest parses a fetch.Request from a GET request and its query parameters.
func ParseGetRequest(r *http.Request) (*Request, error) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		return nil, errors.NewBadRequest("path required")
	}

	var targetURL string
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		targetURL = path
	} else {
		targetURL = "https://" + path
	}

	query := r.URL.Query()

	var timeout int
	if timeoutStr := query.Get("timeout"); timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			timeout = t
		}
	}

	var waitFor int
	if waitForStr := query.Get("wait_for"); waitForStr != "" {
		if w, err := strconv.Atoi(waitForStr); err == nil && w > 0 {
			waitFor = w
		}
	}

	var onlyMainContent bool
	if value := query.Get("only_main_content"); value != "" {
		onlyMainContent = value == "true"
	}

	var excludeTags []string
	if value := query.Get("exclude_tags"); value != "" {
		excludeTags = strings.Split(value, ",")
	}

	return &Request{
		URL:             targetURL,
		Timeout:         timeout,
		WaitFor:         waitFor,
		OnlyMainContent: onlyMainContent,
		ExcludeTags:     excludeTags,
	}, nil
}
