/*
 * Watchman API
 *
 * Moov Watchman offers download, parse, and search functions over numerous U.S. trade sanction lists for complying with regional laws. Also included is a web UI and async webhook notification service to initiate processes on remote systems.
 *
 * API version: v1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package client

// CaptaList struct for CaptaList
type CaptaList struct {
	EntityID       string   `json:"entityID,omitempty"`
	EntityNumber   string   `json:"entityNumber,omitempty"`
	Type           string   `json:"type,omitempty"`
	Programs       []string `json:"programs,omitempty"`
	Name           string   `json:"name,omitempty"`
	Addresses      []string `json:"addresses,omitempty"`
	Remarks        []string `json:"remarks,omitempty"`
	SourceListURL  string   `json:"sourceListURL,omitempty"`
	AlternateNames []string `json:"alternateNames,omitempty"`
	SourceInfoURL  string   `json:"sourceInfoURL,omitempty"`
	IDs            []string `json:"IDs,omitempty"`
	// Match percentage of search query
	Match float32 `json:"match,omitempty"`
	// The highest scoring term from the search query. This term is the precomputed indexed value used by the search algorithm.
	MatchedName string `json:"matchedName,omitempty"`
}
