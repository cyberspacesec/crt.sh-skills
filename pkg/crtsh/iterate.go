// iterate.go — Pagination helper for iterating over all pages of certificate search results
package crtsh

import "context"

// IterateCertificates iterates over all pages of certificate search results,
// calling the callback function for each page of results. Iteration stops when:
//   - the callback returns false
//   - there are no more pages available
//   - the context is cancelled
//
// The params.Page field is used as the starting page (defaults to 1 if 0).
// Each subsequent page is fetched automatically based on the pagination
// information returned by crt.sh.
func (c *Client) IterateCertificates(ctx context.Context, params QueryParams, fn func(certs []Certificate, pagination *Pagination) bool) error {
	page := params.Page
	if page < 1 {
		page = 1
	}

	for {
		params.Page = page
		certs, pagination, err := c.SearchCertificates(ctx, params)
		if err != nil {
			return err
		}

		if !fn(certs, pagination) {
			return nil
		}

		if pagination == nil || pagination.NextPage <= page {
			return nil
		}
		page = pagination.NextPage
	}
}
