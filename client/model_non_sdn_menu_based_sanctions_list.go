/*
 * Watchman API
 *
 * Moov Watchman offers download, parse, and search functions over numerous U.S. trade sanction lists for complying with regional laws. Also included is a web UI and async webhook notification service to initiate processes on remote systems.
 *
 * API version: v1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package client

// NonSdnMenuBasedSanctionsList struct for NonSdnMenuBasedSanctionsList
type NonSdnMenuBasedSanctionsList struct {
	EntityID       string   `json:"EntityID,omitempty"`
	EntityNumber   string   `json:"EntityNumber,omitempty"`
	Type           string   `json:"Type,omitempty"`
	Programs       []string `json:"Programs,omitempty"`
	Name           string   `json:"Name,omitempty"`
	Addresses      []string `json:"Addresses,omitempty"`
	Remarks        []string `json:"Remarks,omitempty"`
	AlternateNames []string `json:"AlternateNames,omitempty"`
	SourceInfoURL  string   `json:"SourceInfoURL,omitempty"`
	IDs            []string `json:"IDs,omitempty"`
	// Match percentage of search query
	Match float32 `json:"match,omitempty"`
	// The highest scoring term from the search query. This term is the precomputed indexed value used by the search algorithm.
	MatchedName string `json:"matchedName,omitempty"`
}
