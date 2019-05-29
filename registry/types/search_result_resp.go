package types

import "github.com/alibaba/pouch/apis/types"

// SearchResultResp response of search images from specific registry
type SearchResultResp struct {

	// NumResults indicates the number of results the query return
	NumResults int64 `json:"num_results,omitempty"`

	// query contains the query string that generated the search results
	Query string `json:"query,omitempty"`

	// Results is a slice containing the actual results for the search
	Results []*types.SearchResultItem `json:"results"`
}
