package osv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/osv-scanner/pkg/lockfile"
	"github.com/google/osv-scanner/pkg/models"

	"golang.org/x/sync/errgroup"
)

const (
	// QueryEndpoint is the URL for posting queries to OSV.
	QueryEndpoint = "https://api.osv.dev/v1/querybatch"
	// GetEndpoint is the URL for getting vulenrabilities from OSV.
	GetEndpoint = "https://api.osv.dev/v1/vulns"
	// BaseVulnerabilityURL is the base URL for detailed vulnerability views.
	BaseVulnerabilityURL = "https://osv.dev/"
	// maxQueriesPerRequest splits up querybatch into multiple requests if
	// number of queries exceed this number
	maxQueriesPerRequest  = 1000
	maxConcurrentRequests = 25
	maxRetryAttempts      = 4
	// jitterMultiplier is multiplied to the retry delay multiplied by rand(0, 1.0)
	jitterMultiplier = 2
)

var RequestUserAgent = ""

// Package represents a package identifier for OSV.
type Package struct {
	PURL      string `json:"purl,omitempty"`
	Name      string `json:"name,omitempty"`
	Ecosystem string `json:"ecosystem,omitempty"`
}

// Query represents a query to OSV.
type Query struct {
	Commit   string            `json:"commit,omitempty"`
	Package  Package           `json:"package,omitempty"`
	Version  string            `json:"version,omitempty"`
	Source   models.SourceInfo `json:"-"` // TODO: Move this into Info struct in v2
	Metadata models.Metadata   `json:"-"`
}

// BatchedQuery represents a batched query to OSV.
type BatchedQuery struct {
	Queries []*Query `json:"queries"`
}

// MinimalVulnerability represents an unhydrated vulnerability entry from OSV.
type MinimalVulnerability struct {
	ID string `json:"id"`
}

// Response represents a full response from OSV.
type Response struct {
	Vulns []models.Vulnerability `json:"vulns"`
}

// MinimalResponse represents an unhydrated response from OSV.
type MinimalResponse struct {
	Vulns []MinimalVulnerability `json:"vulns"`
}

// BatchedResponse represents an unhydrated batched response from OSV.
type BatchedResponse struct {
	Results []MinimalResponse `json:"results"`
}

// HydratedBatchedResponse represents a hydrated batched response from OSV.
type HydratedBatchedResponse struct {
	Results []Response `json:"results"`
}

// MakeCommitRequest makes a commit hash request.
func MakeCommitRequest(commit string) *Query {
	return &Query{
		Commit: commit,
	}
}

// MakePURLRequest makes a PURL request.
func MakePURLRequest(purl string) *Query {
	return &Query{
		Package: Package{
			PURL: purl,
		},
	}
}

func MakePkgRequest(pkgDetails lockfile.PackageDetails) *Query {
	// API has trouble parsing requests with both commit and Package details filled in
	if pkgDetails.Ecosystem == "" && pkgDetails.Commit != "" {
		return &Query{
			Metadata: models.Metadata{
				RepoURL:   pkgDetails.Name,
				DepGroups: pkgDetails.DepGroups,
			},
			Commit: pkgDetails.Commit,
		}
	}

	return &Query{
		Version: pkgDetails.Version,
		Package: Package{
			Name:      pkgDetails.Name,
			Ecosystem: string(pkgDetails.Ecosystem),
		},
		Metadata: models.Metadata{
			DepGroups: pkgDetails.DepGroups,
		},
	}
}

// From: https://stackoverflow.com/a/72408490
func chunkBy[T any](items []T, chunkSize int) [][]T {
	chunks := make([][]T, 0, (len(items)/chunkSize)+1)
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}

	return append(chunks, items)
}

// checkResponseError checks if the response has an error.
func checkResponseError(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	respBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response from server: %w", err)
	}
	defer resp.Body.Close()

	return fmt.Errorf("server response error: %s", string(respBuf))
}

// MakeRequest sends a batched query to osv.dev
func MakeRequest(request BatchedQuery) (*BatchedResponse, error) {
	return MakeRequestWithClient(request, http.DefaultClient)
}

// MakeRequestWithClient sends a batched query to osv.dev with the provided
// http client.
func MakeRequestWithClient(request BatchedQuery, client *http.Client) (*BatchedResponse, error) {
	// API has a limit of 1000 bulk query per request
	queryChunks := chunkBy(request.Queries, maxQueriesPerRequest)
	var totalOsvResp BatchedResponse
	for _, queries := range queryChunks {
		requestBytes, err := json.Marshal(BatchedQuery{Queries: queries})
		if err != nil {
			return nil, err
		}

		resp, err := makeRetryRequest(func() (*http.Response, error) {
			// Make sure request buffer is inside retry, if outside
			// http request would finish the buffer, and retried requests would be empty
			requestBuf := bytes.NewBuffer(requestBytes)
			// We do not need a specific context
			//nolint:noctx
			req, err := http.NewRequest(http.MethodPost, QueryEndpoint, requestBuf)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/json")
			if RequestUserAgent != "" {
				req.Header.Set("User-Agent", RequestUserAgent)
			}

			return client.Do(req)
		})
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var osvResp BatchedResponse
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&osvResp)
		if err != nil {
			return nil, err
		}

		totalOsvResp.Results = append(totalOsvResp.Results, osvResp.Results...)
	}

	return &totalOsvResp, nil
}

// Get a Vulnerability for the given ID.
func Get(id string) (*models.Vulnerability, error) {
	return GetWithClient(id, http.DefaultClient)
}

// GetWithClient gets a Vulnerability for the given ID with the provided http
// client.
func GetWithClient(id string, client *http.Client) (*models.Vulnerability, error) {
	resp, err := makeRetryRequest(func() (*http.Response, error) {
		// We do not need a specific context
		//nolint:noctx
		req, err := http.NewRequest(http.MethodGet, GetEndpoint+"/"+id, nil)
		if err != nil {
			return nil, err
		}
		if RequestUserAgent != "" {
			req.Header.Set("User-Agent", RequestUserAgent)
		}

		return client.Do(req)
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var vuln models.Vulnerability
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&vuln)
	if err != nil {
		return nil, err
	}

	return &vuln, nil
}

// Hydrate fills the results of the batched response with the full
// Vulnerability details.
func Hydrate(resp *BatchedResponse) (*HydratedBatchedResponse, error) {
	return HydrateWithClient(resp, http.DefaultClient)
}

// HydrateWithClient fills the results of the batched response with the full
// Vulnerability details using the provided http client.
func HydrateWithClient(resp *BatchedResponse, client *http.Client) (*HydratedBatchedResponse, error) {
	hydrated := HydratedBatchedResponse{}
	// Preallocate the array to avoid slice reallocations when inserting later
	hydrated.Results = make([]Response, len(resp.Results))
	for idx := range hydrated.Results {
		hydrated.Results[idx].Vulns = make([]models.Vulnerability, len(resp.Results[idx].Vulns))
	}

	g, ctx := errgroup.WithContext(context.TODO())
	g.SetLimit(maxConcurrentRequests)
	for batchIdx, response := range resp.Results {
		for resultIdx, vuln := range response.Vulns {
			id := vuln.ID
			g.Go(func() error {
				// exit early if another hydration request has already failed
				// results are thrown away later, so avoid needless work
				if ctx.Err() != nil {
					return nil //nolint:nilerr // this value doesn't matter to errgroup.Wait()
				}
				vuln, err := GetWithClient(id, client)
				if err != nil {
					return err
				}
				hydrated.Results[batchIdx].Vulns[resultIdx] = *vuln

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &hydrated, nil
}

// makeRetryRequest will return an error on both network errors, and if the response is not 200
func makeRetryRequest(action func() (*http.Response, error)) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := range maxRetryAttempts {
		// rand is initialized with a random number (since go1.20), and is also safe to use concurrently
		// we do not need to use a cryptographically secure random jitter, this is just to spread out the retry requests
		// #nosec G404
		jitterAmount := (rand.Float64() * float64(jitterMultiplier) * float64(i))
		time.Sleep(time.Duration(i*i)*time.Second + time.Duration(jitterAmount*1000)*time.Millisecond)

		resp, err = action()
		if err == nil {
			// Check the response for HTTP errors
			err = checkResponseError(resp)
			if err == nil {
				break
			}
		}
	}

	return resp, err
}
