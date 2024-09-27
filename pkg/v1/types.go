package v1

type Metadata struct {
	Name string
	Namespace string
}

type WebsiteSpec struct {
	GitRepo string
        Puerto int
}

type Website struct {
	Metadata Metadata
	Spec WebsiteSpec
}

type WebsiteWatchEvent struct {
	Type string
	Object Website
}
