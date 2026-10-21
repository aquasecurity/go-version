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

// WithBuildMetadataOrder selects how build metadata is ordered when checking
// constraints. The selected order applies to every constraint operator.
// Version comparison (Compare, Equal, GreaterThan, LessThan, ...) is not
// affected and always ignores build metadata.
type WithBuildMetadataOrder BuildMetadataOrder

func (o WithBuildMetadataOrder) apply(c *conf) {
	c.buildMetadataOrder = BuildMetadataOrder(o)
}
