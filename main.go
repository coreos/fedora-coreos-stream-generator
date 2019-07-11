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
)

var errReleaseIndexMissing = errors.New("Please specify release index url")

func releaseToStream(releaseArch *ReleaseArch, release Release) StreamArch {
	artifacts := StreamArtifacts{}
	cloudImages := StreamImages{}
	if releaseArch.Media.Aws != nil {
		aws := StreamMediaDetails{
			Release: release.Release,
			Formats: releaseArch.Media.Aws.Artifacts,
		}
		artifacts.Aws = &aws
		awsAmis := StreamAwsImage{
			Regions: make(map[string]*StreamAwsAMI),
		}

		if releaseArch.Media.Aws != nil && releaseArch.Media.Aws.Images != nil {
			for region, ami := range *releaseArch.Media.Aws.Images {
				streamAwsAMI := StreamAwsAMI{}
				streamAwsAMI.Release = release.Release
				streamAwsAMI.Image = *ami.Image
				awsAmis.Regions[region] = &streamAwsAMI

			}

			cloudImages.Aws = &awsAmis
		}
	}

	if releaseArch.Media.Azure != nil {
		azure := StreamMediaDetails{
			Release: release.Release,
			Formats: releaseArch.Media.Azure.Artifacts,
		}
		artifacts.Azure = &azure

		if az := releaseArch.Media.Azure.Images; az != nil && az.Global != nil && az.Global.Image != nil {
			azureImage := StreamCloudImage{}
			azureImage.Image = "Fedora:CoreOS:stable:latest"
			cloudImages.Azure = &azureImage
		}

	}

	if releaseArch.Media.Gcp != nil {
		gcp := StreamMediaDetails{
			Release: release.Release,
			Formats: releaseArch.Media.Gcp.Artifacts,
		}
		artifacts.Gcp = &gcp

		if releaseArch.Media.Gcp != nil && releaseArch.Media.Gcp.Image != nil {
			gcpImage := StreamCloudImage{}
			gcpImage.Image = "projects/fedora-cloud/global/images/family/fedora-coreos-stable"
			cloudImages.Gcp = &gcpImage

		}
	}

	if releaseArch.Media.Digitalocean != nil {
		digitalOcean := StreamMediaDetails{
			Release: release.Release,
			Formats: releaseArch.Media.Digitalocean.Artifacts,
		}
		artifacts.Digitalocean = &digitalOcean

		artifacts.Digitalocean = &digitalOcean
		digitalOceanImage := StreamCloudImage{Image: "fedora-coreos-stable"}
		cloudImages.Digitalocean = &digitalOceanImage
	}

	if releaseArch.Media.Packet != nil {
		packet := StreamMediaDetails{
			Release: release.Release,
			Formats: releaseArch.Media.Packet.Artifacts,
		}
		artifacts.Packet = &packet

		packetImage := StreamCloudImage{Image: "fedora_coreos_stable"}
		cloudImages.Packet = &packetImage
	}

	if releaseArch.Media.Openstack != nil {
		openstack := StreamMediaDetails{
			Release: release.Release,
			Formats: releaseArch.Media.Openstack.Artifacts,
		}
		artifacts.Openstack = &openstack
	}

	if releaseArch.Media.Qemu != nil {
		qemu := StreamMediaDetails{
			Release: release.Release,
			Formats: releaseArch.Media.Qemu.Artifacts,
		}
		artifacts.Qemu = &qemu
	}

	if releaseArch.Media.Virtualbox != nil {
		virtualbox := StreamMediaDetails{
			Release: release.Release,
			Formats: releaseArch.Media.Virtualbox.Artifacts,
		}
		artifacts.Virtualbox = &virtualbox
	}

	if releaseArch.Media.Vmware != nil {
		vmware := StreamMediaDetails{
			Release: release.Release,
			Formats: releaseArch.Media.Vmware.Artifacts,
		}
		artifacts.Vmware = &vmware
	}

	if releaseArch.Media.Metal != nil {
		metal := StreamMediaDetails{
			Release: release.Release,
			Formats: releaseArch.Media.Metal.Artifacts,
		}
		artifacts.Metal = &metal
	}

	streamArch := StreamArch{
		Artifacts: artifacts,
	}

	if cloudImages != (StreamImages{}) {
		streamArch.Images = &cloudImages
	}

	return streamArch
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
	var overrideFilename string
	flag.StringVar(&overrideFilename, "override", "", "Override file location for the required stream")
	var outputFile string
	flag.StringVar(&outputFile, "output-file", "", "Save output into a file")
	var prettyPrint bool
	flag.BoolVar(&prettyPrint, "pretty-print", false, "Pretty-print output")

	flag.Parse()

	if releasesURL == "" {
		return errReleaseIndexMissing
	}

	releasePath, err := ReleaseURL(releasesURL)
	if err != nil {
		return fmt.Errorf("Error with Release Index: %v", err)
	}

	parsedURL, err := url.Parse(releasePath)
	if err != nil {
		return fmt.Errorf("Error while parsing release path: %v", err)
	}

	var decoder *json.Decoder
	if parsedURL.Scheme == "" {
		// It is most likely a local file.
		releasesMetadataFile, err := os.Open(releasePath)
		if err != nil {
			return fmt.Errorf("Error opening file: %v", err)
		}

		defer releasesMetadataFile.Close()
		decoder = json.NewDecoder(releasesMetadataFile)
	} else {
		resp, err := http.Get(releasePath)
		if err != nil {
			return fmt.Errorf("Error while fetching: %v", err)
		}

		defer resp.Body.Close()
		decoder = json.NewDecoder(resp.Body)
	}

	var release Release
	if err = decoder.Decode(&release); err != nil {
		return fmt.Errorf("Error while decoding json: %v", err)
	}

	streamArch := make(map[string]*StreamArch)
	for arch, releaseArch := range release.Architectures {
		archContent := releaseToStream(releaseArch, release)
		streamArch[arch] = &archContent
	}

	streamMetadata := StreamMetadata{
		Stream:        release.Stream,
		Metadata:      Metadata{LastModified: time.Now().UTC().Format(time.RFC3339)},
		Architectures: streamArch,
		Updates:       StreamUpdates{Release: release.Release},
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
