package marketplace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	galleryAPI    = "https://marketplace.visualstudio.com/_apis/public/gallery/extensionquery"
	apiVersion    = "7.1-preview.1"
	maxRetries    = 3
	retryBaseWait = 2 * time.Second
)

type Client struct {
	http    *http.Client
	verbose bool
}

func NewClient(verbose bool) *Client {
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		verbose: verbose,
	}
}

func (c *Client) QueryExtension(name string, pinned bool) (*ExtensionResult, error) {
	flags := FlagsLatest
	if pinned {
		flags = FlagsAllVersions
	}

	req := QueryRequest{
		Filters: []Filter{{
			Criteria: []Criterion{{
				FilterType: 7, // extension name
				Value:      name,
			}},
		}},
		Flags: flags,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling query: %w", err)
	}

	var resp QueryResponse
	if err := c.doWithRetry("POST", galleryAPI, body, &resp); err != nil {
		return nil, fmt.Errorf("querying %s: %w", name, err)
	}

	if len(resp.Results) == 0 || len(resp.Results[0].Extensions) == 0 {
		return nil, fmt.Errorf("extension %q not found", name)
	}

	return &resp.Results[0].Extensions[0], nil
}

func (c *Client) doWithRetry(method, url string, reqBody []byte, out any) error {
	for attempt := range maxRetries {
		err := c.do(method, url, reqBody, out)
		if err == nil {
			return nil
		}

		if attempt < maxRetries-1 {
			wait := retryBaseWait * time.Duration(1<<attempt)
			if c.verbose {
				log.Printf("retry %d/%d for %s %s: %v (waiting %s)", attempt+1, maxRetries, method, url, err, wait)
			}
			time.Sleep(wait)
			continue
		}
		return err
	}
	return nil // unreachable
}

func (c *Client) do(method, url string, reqBody []byte, out any) error {
	var bodyReader io.Reader
	if reqBody != nil {
		bodyReader = bytes.NewReader(reqBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json;api-version="+apiVersion)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
