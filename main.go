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

var errReleaseIndexMissing = errors.New("Please specify release index url or release override")

func releaseToStream(releaseArch *ReleaseArch, rel Release) StreamArch {
	artifacts := StreamArtifacts{}
	cloudImages := StreamImages{}
	if releaseArch.Media.Aws != nil {
		aws := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Aws.Artifacts,
		}
		artifacts.Aws = &aws
		awsAmis := StreamAwsImage{
			Regions: make(map[string]*StreamAwsAMI),
		}

		if releaseArch.Media.Aws != nil && releaseArch.Media.Aws.Images != nil {
			for region, ami := range *releaseArch.Media.Aws.Images {
				streamAwsAMI := StreamAwsAMI{}
				streamAwsAMI.Release = rel.Release
				streamAwsAMI.Image = *ami.Image
				awsAmis.Regions[region] = &streamAwsAMI

			}

			cloudImages.Aws = &awsAmis
		}
	}

	if releaseArch.Media.Azure != nil {
		azure := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Azure.Artifacts,
		}
		artifacts.Azure = &azure

		// Not enabled right now
		// if az := releaseArch.Media.Azure.Images; az != nil && az.Global != nil && az.Global.Image != nil {
		// 	azureImage := StreamCloudImage{}
		// 	azureImage.Image = fmt.Sprintf("Fedora:CoreOS:%s:latest", rel.Stream)
		// 	cloudImages.Azure = &azureImage
		// }
	}

	if releaseArch.Media.Aliyun != nil {
		aliyun := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Aliyun.Artifacts,
		}
		artifacts.Aliyun = &aliyun
	}

	if releaseArch.Media.Exoscale != nil {
		exoscale := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Exoscale.Artifacts,
		}
		artifacts.Exoscale = &exoscale
	}

	if releaseArch.Media.Vultr != nil {
		vultr := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Vultr.Artifacts,
		}
		artifacts.Vultr = &vultr
	}

	if releaseArch.Media.Gcp != nil {
		gcp := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Gcp.Artifacts,
		}
		artifacts.Gcp = &gcp

		if releaseArch.Media.Gcp != nil && releaseArch.Media.Gcp.Image != nil {
			gcpImage := StreamGcpImage{
				Name:    releaseArch.Media.Gcp.Image.Name,
				Family:  releaseArch.Media.Gcp.Image.Family,
				Project: releaseArch.Media.Gcp.Image.Project,
			}
			cloudImages.Gcp = &gcpImage

		}
	}

	if releaseArch.Media.Digitalocean != nil {
		digitalOcean := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Digitalocean.Artifacts,
		}
		artifacts.Digitalocean = &digitalOcean

		/* We're producing artifacts but they're not yet available
		   in DigitalOcean as distribution images.
		digitalOceanImage := StreamCloudImage{Image: fmt.Sprintf("fedora-coreos-%s", release.Stream)}
		cloudImages.Digitalocean = &digitalOceanImage
		*/
	}

	if releaseArch.Media.Ibmcloud != nil {
		ibmcloud := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Ibmcloud.Artifacts,
		}
		artifacts.Ibmcloud = &ibmcloud
	}

	if releaseArch.Media.Packet != nil {
		packet := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Packet.Artifacts,
		}
		artifacts.Packet = &packet

		packetImage := StreamCloudImage{Image: fmt.Sprintf("fedora_coreos_%s", rel.Stream)}
		cloudImages.Packet = &packetImage
	}

	if releaseArch.Media.Openstack != nil {
		openstack := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Openstack.Artifacts,
		}
		artifacts.Openstack = &openstack
	}

	if releaseArch.Media.Qemu != nil {
		qemu := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Qemu.Artifacts,
		}
		artifacts.Qemu = &qemu
	}

	if releaseArch.Media.Virtualbox != nil {
		virtualbox := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Virtualbox.Artifacts,
		}
		artifacts.Virtualbox = &virtualbox
	}

	if releaseArch.Media.Vmware != nil {
		vmware := StreamMediaDetails{
			Release: rel.Release,
			Formats: releaseArch.Media.Vmware.Artifacts,
		}
		artifacts.Vmware = &vmware
	}

	if releaseArch.Media.Metal != nil {
		metal := StreamMediaDetails{
			Release: rel.Release,
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
		releasePath, err = ReleaseURL(releasesURL)
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

	var rel Release
	if err = decoder.Decode(&rel); err != nil {
		return fmt.Errorf("Error while decoding json: %v", err)
	}

	streamArch := make(map[string]*StreamArch)
	for arch, releaseArch := range rel.Architectures {
		archContent := releaseToStream(releaseArch, rel)
		streamArch[arch] = &archContent
	}

	streamMetadata := StreamMetadata{
		Stream:        rel.Stream,
		Metadata:      Metadata{LastModified: time.Now().UTC().Format(time.RFC3339)},
		Architectures: streamArch,
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
