package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/coreos/stream-metadata-go/release"
	"github.com/coreos/stream-metadata-go/stream"
)

var errReleaseIndexMissing = errors.New("Please specify release index url or release override")

// getReleaseURL gets path for latest release.json available
func getReleaseURL(releaseIndexURL string) (string, error) {
	var relIndex release.Index
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

	return relIndex.Releases[len(relIndex.Releases)-1].MetadataURL, nil
}

func overrideData(original, override interface{}) interface{} {
	switch override1 := override.(type) {
	case map[string]interface{}:
		original1, ok := original.(map[string]interface{})
		if !ok {
			return override1
		}
		for key, value1 := range original1 {
			if value2, ok := override1[key]; ok {
				override1[key] = overrideData(value1, value2)
			} else {
				override1[key] = value1
			}
		}
	case nil:
		original1, ok := original.(map[string]interface{})
		if ok {
			return original1
		}
	}
	return override
}

func run() error {
	var releasesURL string
	flag.StringVar(&releasesURL, "releases", "", "Release index location for the required stream")
	var overrideReleasePath string
	flag.StringVar(&overrideReleasePath, "release", "", "Override release metadata location")
	var overrideFilename string
	flag.StringVar(&overrideFilename, "override", "", "Override file location for the required stream")
	var outputFile string
	flag.StringVar(&outputFile, "output-file", "", "Save output into a file")
	var prettyPrint bool
	flag.BoolVar(&prettyPrint, "pretty-print", false, "Pretty-print output")

	flag.Parse()

	var releasePath string
	if releasesURL == "" && overrideReleasePath == "" {
		return errReleaseIndexMissing
	} else if releasesURL != "" && overrideReleasePath != "" {
		return fmt.Errorf("Can't specify both -releases and -release")
	} else if overrideReleasePath != "" {
		releasePath = overrideReleasePath
	} else {
		var err error
		releasePath, err = getReleaseURL(releasesURL)
		if err != nil {
			return fmt.Errorf("Error with Release Index: %v", err)
		}
	}

	parsedURL, err := url.Parse(releasePath)
	if err != nil {
		return fmt.Errorf("Error while parsing release path: %v", err)
	}

	var decoder *json.Decoder
	if parsedURL.Scheme == "" {
		// It is most likely a local file.
		releaseMetadataFile, err := os.Open(releasePath)
		if err != nil {
			return fmt.Errorf("Error opening file: %v", err)
		}

		defer releaseMetadataFile.Close()
		decoder = json.NewDecoder(releaseMetadataFile)
	} else {
		resp, err := http.Get(releasePath)
		if err != nil {
			return fmt.Errorf("Error while fetching: %v", err)
		}

		defer resp.Body.Close()
		decoder = json.NewDecoder(resp.Body)
	}

	var rel release.Release
	if err = decoder.Decode(&rel); err != nil {
		return fmt.Errorf("Error while decoding json: %v", err)
	}

	streamMetadata := stream.Stream{
		Stream:        rel.Stream,
		Metadata:      stream.Metadata{LastModified: time.Now().UTC().Format(time.RFC3339)},
		Architectures: rel.ToStreamArchitectures(),
	}

	if overrideFilename != "" {
		overrideFile, err := os.Open(overrideFilename)
		if err != nil {
			return fmt.Errorf("Can't open file %s: %v", overrideFilename, err)
		}
		defer overrideFile.Close()

		streamMetadataJSON, err := json.Marshal(&streamMetadata)
		if err != nil {
			return fmt.Errorf("Error during Marshal: %v", err)
		}
		streamMetadataMap := make(map[string]interface{})
		if err = json.Unmarshal(streamMetadataJSON, &streamMetadataMap); err != nil {
			return fmt.Errorf("Error during Unmarshal: %v", err)
		}

		overrideMap := make(map[string]interface{})
		overrideDecoder := json.NewDecoder(overrideFile)
		if err = overrideDecoder.Decode(&overrideMap); err != nil {
			return fmt.Errorf("Error while decoding: %v", err)
		}

		streamMetadataInterface := overrideData(streamMetadataMap, overrideMap)
		streamMetadataMap = streamMetadataInterface.(map[string]interface{})

		// We are doing Marshal and Unmarshal of streamMetadataMap to keep json in ordered way
		streamMetadataJSON, err = json.Marshal(streamMetadataMap)
		if err != nil {
			return fmt.Errorf("Error during Marshal: %v", err)
		}
		if err = json.Unmarshal(streamMetadataJSON, &streamMetadata); err != nil {
			return fmt.Errorf("Error during Unmarshal: %v", err)
		}
	}

	var out io.Writer
	// If outputFile option not specified print on stdout
	if outputFile != "" {
		streamFile, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("Can't open file: %v", err)
		}

		defer streamFile.Close()
		out = streamFile
	} else {
		out = os.Stdout
	}

	encoder := json.NewEncoder(out)
	if prettyPrint {
		encoder.SetIndent("", "    ")
	}
	if err := encoder.Encode(&streamMetadata); err != nil {
		return fmt.Errorf("Error while encoding: %v", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)

		if err == errReleaseIndexMissing {
			flag.Usage()
		}

		os.Exit(1)
	}
}
