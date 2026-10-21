package version

type conf struct {
	buildMetadataOrder BuildMetadataOrder
}

type ConstraintOption interface {
	apply(*conf)
}

// BuildMetadataOrder identifies a strategy for ordering build metadata.
//
// Build metadata is ignored by default, preserving the package's existing
// behavior.
type BuildMetadataOrder int

const (
	// BuildMetadataOrderNodeSemver orders build metadata like node-semver's
	// compareBuild: a version without metadata is lower than one with metadata,
	// and dot-separated identifiers are compared like prerelease identifiers.
	BuildMetadataOrderNodeSemver BuildMetadataOrder = iota + 1
)

// WithBuildMetadataOrder selects how CompareWithOptions orders build metadata.
// When passed to NewConstraints, the selected order applies to every constraint
// operator.
type WithBuildMetadataOrder BuildMetadataOrder

func (o WithBuildMetadataOrder) apply(c *conf) {
	c.buildMetadataOrder = BuildMetadataOrder(o)
}
