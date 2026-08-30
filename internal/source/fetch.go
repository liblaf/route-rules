package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxSourceBytes = 64 << 20

type FetchMetadata struct {
	URL          string `json:"url"`
	Bytes        int    `json:"bytes"`
	ETag         string `json:"etag,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
}

type fetchedDocument struct {
	payload  []byte
	metadata FetchMetadata
}

type Fetcher struct {
	client *http.Client
	cache  map[string]fetchedDocument
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{Timeout: 45 * time.Second},
		cache:  make(map[string]fetchedDocument),
	}
}

func (fetcher *Fetcher) Fetch(ctx context.Context, url string) ([]byte, FetchMetadata, error) {
	if cached, exists := fetcher.cache[url]; exists {
		return append([]byte(nil), cached.payload...), cached.metadata, nil
	}

	var lastError error
	for attempt := 1; attempt <= 3; attempt++ {
		payload, metadata, err := fetcher.fetchOnce(ctx, url)
		if err == nil {
			fetcher.cache[url] = fetchedDocument{payload: payload, metadata: metadata}
			return append([]byte(nil), payload...), metadata, nil
		}
		lastError = err
		if ctx.Err() != nil {
			break
		}
		if attempt < 3 {
			timer := time.NewTimer(time.Duration(attempt) * 500 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, FetchMetadata{}, ctx.Err()
			case <-timer.C:
			}
		}
	}
	return nil, FetchMetadata{}, fmt.Errorf("fetch %s after 3 attempts: %w", url, lastError)
}

func (fetcher *Fetcher) fetchOnce(ctx context.Context, url string) ([]byte, FetchMetadata, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, FetchMetadata{}, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("User-Agent", "liblaf/route-rules")
	request.Header.Set("Accept", "text/plain, application/yaml, text/csv, */*")

	response, err := fetcher.client.Do(request)
	if err != nil {
		return nil, FetchMetadata{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return nil, FetchMetadata{}, fmt.Errorf("unexpected HTTP status %s", response.Status)
	}

	limited := io.LimitReader(response.Body, maxSourceBytes+1)
	payload, err := io.ReadAll(limited)
	if err != nil {
		return nil, FetchMetadata{}, fmt.Errorf("read response: %w", err)
	}
	if len(payload) == 0 {
		return nil, FetchMetadata{}, fmt.Errorf("empty response")
	}
	if len(payload) > maxSourceBytes {
		return nil, FetchMetadata{}, fmt.Errorf("response exceeds %d bytes", maxSourceBytes)
	}
	metadata := FetchMetadata{
		URL:          url,
		Bytes:        len(payload),
		ETag:         response.Header.Get("ETag"),
		LastModified: response.Header.Get("Last-Modified"),
	}
	return payload, metadata, nil
}
