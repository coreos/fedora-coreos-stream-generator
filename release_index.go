package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

// ReleaseIndex for accessing Release Index metadata
type ReleaseIndex struct {
	Releases []Index `json:"releases"`
}

// Index - details for single release
type Index struct {
	Metadata string `json:"metadata"`
}

// ReleaseURL gets path for latest release.json available
func ReleaseURL(releaseIndexURL string) (string, error) {
	var relIndex ReleaseIndex
	parsedURL, err := url.Parse(releaseIndexURL)
	if err != nil {
		return "", err
	}

	var decoder *json.Decoder
	if parsedURL.Scheme == "" {
		// It is most likely a local file.
		releases, err := os.Open(releaseIndexURL)
		if err != nil {
			return "", err
		}

		defer releases.Close()
		decoder = json.NewDecoder(releases)
	} else {
		resp, err := http.Get(releaseIndexURL)
		if err != nil {
			return "", err
		}

		defer resp.Body.Close()
		decoder = json.NewDecoder(resp.Body)
	}

	if err := decoder.Decode(&relIndex); err != nil {
		return "", err
	}
	if len(relIndex.Releases) < 1 {
		return "", fmt.Errorf("No release available to process")
	}

	return relIndex.Releases[len(relIndex.Releases)-1].Metadata, nil
}
