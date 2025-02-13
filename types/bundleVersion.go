package types

// BundleVersion The version of the build that identifies an iteration of the bundle.
type BundleVersion string

func (b BundleVersion) String() string {
	return string(b)
}
