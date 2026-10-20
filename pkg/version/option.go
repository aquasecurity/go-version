package version

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

// Option configures CompareWithOptions and NewConstraints. Other comparison
// methods and Collection sorting do not accept options.
type Option interface {
	apply(*options)
}

type options struct {
	buildMetadataOrder BuildMetadataOrder
}

type buildMetadataOrderOption BuildMetadataOrder

func (option buildMetadataOrderOption) apply(config *options) {
	config.buildMetadataOrder = BuildMetadataOrder(option)
}

// WithBuildMetadataOrder selects how CompareWithOptions orders build metadata.
// When passed to NewConstraints, the selected order applies to every constraint
// operator.
func WithBuildMetadataOrder(order BuildMetadataOrder) Option {
	return buildMetadataOrderOption(order)
}

func newOptions(opts ...Option) options {
	var config options
	for _, opt := range opts {
		if opt != nil {
			opt.apply(&config)
		}
	}
	return config
}
