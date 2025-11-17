package acvs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SimpleToSFetcher implements ToSFetcher with basic HTTP fetching.
type SimpleToSFetcher struct {
	client  *http.Client
	timeout time.Duration
}

// NewSimpleToSFetcher creates a new simple ToS fetcher.
func NewSimpleToSFetcher() *SimpleToSFetcher {
	return &SimpleToSFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// FetchToS fetches ToS content from a URL.
func (f *SimpleToSFetcher) FetchToS(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "ACM-ACVS/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// DiscoverToSURL attempts to discover the ToS URL for a site.
// Phase II: Uses simple heuristics.
// Phase III: Could use sitemap parsing or link crawling.
func (f *SimpleToSFetcher) DiscoverToSURL(ctx context.Context, site string) (string, error) {
	// Common ToS URL patterns
	patterns := []string{
		"https://%s/terms",
		"https://%s/terms-of-service",
		"https://%s/tos",
		"https://%s/legal/terms",
		"https://%s/legal/tos",
		"https://www.%s/terms",
		"https://www.%s/terms-of-service",
	}

	// Try common patterns
	for _, pattern := range patterns {
		url := fmt.Sprintf(pattern, site)

		// Quick HEAD request to check if URL exists
		req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("User-Agent", "ACM-ACVS/1.0")

		resp, err := f.client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return url, nil
		}
	}

	// If all patterns failed, return default
	return fmt.Sprintf("https://%s/terms", site), fmt.Errorf("could not discover ToS URL for %s", site)
}

// SetTimeout sets the HTTP client timeout.
func (f *SimpleToSFetcher) SetTimeout(timeout time.Duration) {
	f.timeout = timeout
	f.client.Timeout = timeout
}

// extractTextFromHTML is a simple HTML-to-text converter.
// Phase III: Use a proper HTML parser.
func extractTextFromHTML(html string) string {
	// Very basic HTML tag stripping
	// TODO: Use proper HTML parser in Phase III
	text := html

	// Remove script and style tags
	text = removeTagContent(text, "script")
	text = removeTagContent(text, "style")

	// Remove all HTML tags
	text = strings.ReplaceAll(text, "<", " <")
	text = strings.ReplaceAll(text, ">", "> ")

	// Simple tag removal (not perfect, but good enough for stub)
	for strings.Contains(text, "<") && strings.Contains(text, ">") {
		start := strings.Index(text, "<")
		end := strings.Index(text[start:], ">")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+1:]
	}

	// Normalize whitespace
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}

// removeTagContent removes content between opening and closing tags.
func removeTagContent(html, tag string) string {
	openTag := "<" + tag
	closeTag := "</" + tag + ">"

	for {
		start := strings.Index(strings.ToLower(html), strings.ToLower(openTag))
		if start == -1 {
			break
		}

		end := strings.Index(strings.ToLower(html[start:]), strings.ToLower(closeTag))
		if end == -1 {
			break
		}

		html = html[:start] + html[start+end+len(closeTag):]
	}

	return html
}
