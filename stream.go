package main

// StreamMetadata contains artifacts available in a stream
type StreamMetadata struct {
	Stream        string                 `json:"stream"`
	Metadata      Metadata               `json:"metadata"`
	Architectures map[string]*StreamArch `json:"architectures"`
	// Updates       StreamUpdates          `json:"updates"`
}

// StreamArch release details for x86_64 architetcure
type StreamArch struct {
	Artifacts StreamArtifacts `json:"artifacts"`
	Images    *StreamImages   `json:"images,omitempty"`
}

// StreamArtifacts contains shipped artifacts list
type StreamArtifacts struct {
	Aliyun       *StreamMediaDetails `json:"aliyun,omitempty"`
	Aws          *StreamMediaDetails `json:"aws,omitempty"`
	Azure        *StreamMediaDetails `json:"azure,omitempty"`
	Digitalocean *StreamMediaDetails `json:"digitalocean,omitempty"`
	Exoscale     *StreamMediaDetails `json:"exoscale,omitempty"`
	Gcp          *StreamMediaDetails `json:"gcp,omitempty"`
	Metal        *StreamMediaDetails `json:"metal,omitempty"`
	Openstack    *StreamMediaDetails `json:"openstack,omitempty"`
	Packet       *StreamMediaDetails `json:"packet,omitempty"`
	Qemu         *StreamMediaDetails `json:"qemu,omitempty"`
	Virtualbox   *StreamMediaDetails `json:"virtualbox,omitempty"`
	Vmware       *StreamMediaDetails `json:"vmware,omitempty"`
	Vultr        *StreamMediaDetails `json:"vultr,omitempty"`
}

// StreamMediaDetails contains image artifact and release detail
type StreamMediaDetails struct {
	Release string                  `json:"release"`
	Formats map[string]*ImageFormat `json:"formats"`
}

// StreamImages contains images available in cloud providers
type StreamImages struct {
	Aws          *StreamAwsImage   `json:"aws,omitempty"`
	Azure        *StreamCloudImage `json:"azure,omitempty"`
	Gcp          *StreamGcpImage   `json:"gcp,omitempty"`
	Digitalocean *StreamCloudImage `json:"digitalocean,omitempty"`
	Packet       *StreamCloudImage `json:"packet,omitempty"`
}

// StreamCloudImage image for Cloud provider
type StreamCloudImage struct {
	Image string `json:"image,omitempty"`
}

// StreamAwsImage Aws images
type StreamAwsImage struct {
	Regions map[string]*StreamAwsAMI `json:"regions,omitempty"`
}

// StreamAwsAMI aws AMI detail
type StreamAwsAMI struct {
	Release string `json:"release"`
	Image   string `json:"image"`
}

// StreamGcpImage GCP cloud image information
type StreamGcpImage struct {
	Project string `json:"project,omitempty"`
	Family  string `json:"family,omitempty"`
	Name    string `json:"name,omitempty"`
}

// StreamUpdates contains release version
// type StreamUpdates struct {
// 	Release string `json:"release"`
// }
