package main

// ReleaseIndex for accessing Release Index metadata
type ReleaseIndex struct {
	Releases []Index `json:"releases"`
}

// Index - details for single release
type Index struct {
	Metadata string `json:"metadata"`
}
