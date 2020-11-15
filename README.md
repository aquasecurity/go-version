# go-version

![Test](https://github.com/aquasecurity/go-version/workflows/Test/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/aquasecurity/go-version)](https://goreportcard.com/report/github.com/aquasecurity/go-version)
![GitHub](https://img.shields.io/github/license/aquasecurity/go-version)

go-version is a library for parsing versions and version constraints, and verifying versions against a set of constraints.
go-version can sort a collection of versions properly, handles prerelease versions, etc.

go-version provides two packages:
- [semver](./pkg/semver)
    - [Semantic Versioning](https://semver.org/)
    - MAJOR.MINOR.PATCH-PRERELEASE+BUILDMETADATA (e.g. 1.1.3-alpha+110)
- [version](./pkg/version)
    - Semantic Versioning like versioning
    - Accept more than 3 numbers (e.g. 2.3.1.4)
    
# Table of Contents
- [semver](#semver)
  * [Parsing and Comparison](#semver-parsing-and-comparison)
  * [Sorting](#semver-sorting)
  * [Constraints](#semver-constraints)
    + [Pre-release](#semver-pre-release)
    + [Missing major/minor/patch versions](#missing-majorminorpatch-versions)
- [version](#version)
  * [Parsing, Comparison, and Sorting](#version-parsing-comparison-and-sorting)
  * [Constraints](#version-constraints)
    + [Pre-release](#pre-release)
    + [Zero Padding](#zero-padding)
 

## semver
Versions used with `semver` package must follow [Semantic Versioning](https://semver.org/).

### SemVer Parsing and Comparison
When two versions are compared using functions such as Compare, LessThan, and others, it will follow the specification and always include pre-releases within the comparison.
It will provide an answer that is valid with the comparison section of [the spec](https://semver.org/#spec-item-11).

See [example](./examples/semver/main.go)

```
v1, _ := semver.Parse("1.2.0")
v2, _ := semver.Parse("1.2.1")

// Comparison example. There is also GreaterThan, Equal, and just
// a simple Compare that returns an int allowing easy >=, <=, etc.
if v1.LessThan(v2) {
	fmt.Printf("%s is less than %s", v1, v2)
}
```

### SemVer Sorting
It follows [the spec](https://semver.org/#spec-item-11).

See [example](./examples/semver/main.go)

```
versionsRaw := []string{"1.1.0", "0.7.1", "1.4.0", "1.4.0-alpha", "1.4.1-beta", "1.4.0-alpha.2+20130313144700"}
versions := make([]semver.Version, len(versionsRaw))
for i, raw := range versionsRaw {
	v, _ := semver.Parse(raw)
	versions[i] = v
}

// After this, the versions are properly sorted
sort.Sort(semver.Collection(versions))
```

### SemVer Constraints
Comma-separated version constraints are considered an `AND`.
For example, ">= 1.2.3, < 2.0.0" means the version needs to be greater than or equal to 1.2 and less than 3.0.0.
In addition, they can be separated by `|| (OR)`.
For example, ">= 1.2.3, < 2.0.0 || > 4.0.0" means the version needs to be greater than or equal to 1.2 and less than 3.0.0, or greater than 4.0.0.


See [example](./examples/semver/main.go)

```
v, _ := semver.Parse("2.1.0")
c, _ := semver.NewConstraints(">= 1.0, < 1.4 || > 2.0")

if c.Check(v) {
	fmt.Printf("%s satisfies constraints '%s'", v, c)
}
```

Supported operators
- `=` : you accept that exact version
- `!=` : not equal
- `>` : you accept any version higher than the one you specify
- `>=` : you accept any version equal to or higher than the one you specify
- `<` : you accept any version lower to the one you specify
- `<=` : you accept any version equal or lower to the one you specify
- `^` : it will only do updates that do not change the leftmost non-zero number.
    - e.g. `^1.2.3` := `>=1.2.3, <2.0.0`
- `~` : allows patch-level changes if a minor version is specified on the comparator. Allows minor-level changes if not.
    - e.g. `~1.2.3` := `>=1.2.3, <1.3.0`
    
`^` and `~` work like `npm`. See [here](https://docs.npmjs.com/cli/v6/using-npm/semver).

As for version constraints, there are a few caveats since the constraints are not part of [the specification](https://semver.org/).
- Pre-release
- Missing major/minor/patch versions


#### SemVer Pre-release
A pre-release version may be denoted by appending a hyphen and a series of dot separated identifiers immediately following the patch version.
Pre-release versions have a lower precedence than the associated normal version (e.g. 1.2.3-alpha < 1.2.3).
A pre-release version indicates that the version is unstable and might not satisfy the intended compatibility requirements as denoted by its associated normal version.

`semver` comparisons using constraints without a prerelease comparator will skip prerelease versions.
For example, >=1.2.3 will skip pre-releases when looking at a list of releases.

In the following example, `2.1.0-alpha` looks greater than `2.0.0`, but `c.Check(v)` returns false.

```
v, _ := semver.Parse("2.1.0-alpha")
c, _ := semver.NewConstraints(">= 2.0.0")

c.Check(v) // false
```

Constraints with a prerelease comparator will include prerelease versions.

```
v, _ := semver.Parse("2.1.0-alpha")
c, _ := semver.NewConstraints(">= 2.0.0-alpha")

c.Check(v) // true
```

Note that this is different from the behavior of npm.
`>= 2.0.0-alpha` allows pre-releases in the 2.0.0 version only, if they are greater than or equal to alpha.
So, 2.0.0-beta would be allowed, while 2.1.0-alpha would not.
You can use [go-npm-version](https://github.com/aquasecurity/go-npm-version) for npm version comparion.
It strictly follows the npm rules.

If you want to include pre-releases even with no pre-releases constraint, you can pass `semver.WithPreRelease(true)` as an argument of `semver.NewConstraints`

```
v, _ := semver.Parse("2.1.0-alpha")
c, _ := semver.NewConstraints(">= 2.0.0", semver.WithPreRelease(true))

c.Check(v) // true
```

#### Missing major/minor/patch versions 
If some of major/minor/patch versions are not specified, it is treated as `*` by default.
In short, `3.1.3` satisfies `= 3` because `= 3` is converted to `= 3.*.*`.

```
v, _ := semver.Parse("2.3.4")
c, _ := semver.NewConstraints("= 2")

c.Check(v) // true
```

Then, `2.2.3` doesn't satisfy `> 2` as `> 2` is treated as `> 2.*.*` = `>= 3.0.0`

```
v, _ := semver.Parse("2.2.3")
c, _ := semver.NewConstraints("> 2")

c.Check(v) // false
```

`3.3.9` satisifies `= 3.3`, and `5.1.2` doesn't satisfy `> 5.1` likewise.

If you want to treat them as 0, you can pass `semver.WithZeroPadding(true)` as an argument of `semver.NewConstraints`

```
v, _ := semver.Parse("2.3.4")
c, _ := semver.NewConstraints("= 2", semver.WithZeroPadding(true))

c.Check(v) // false
```

## version
Versions used with `version` package follows [Semantic Versioning](https://semver.org/) like versioning.
It accepts more than 3 numbers such as `2.2.4.3`.

It works almost as well as `semver` package. It also accepts pre-release and build metadata.

### Version Parsing, Comparison, and Sorting
When two versions are compared using functions such as Compare, LessThan, and others, it will follow the specification and always include pre-releases within the comparison.

See [example](./examples/version/main.go)

```
v1, _ := version.Parse("1.2.0.9-alpha")
v2, _ := version.Parse("1.2.1.0+11")

// Comparison example. There is also GreaterThan, Equal, and just
// a simple Compare that returns an int allowing easy >=, <=, etc.
if v1.LessThan(v2) {
	fmt.Printf("%s is less than %s\n", v1, v2)
}
```

It also supports version sorting.

### Version Constraints
It is almost the same as `semver` package, but there are some differences.

See [example](./examples/version/main.go)

```
v, _ := version.Parse("2.1.0.1-alpha")
c, _ := version.NewConstraints(">= 1.0, < 1.4 || > 2.1")

if c.Check(v) {
	fmt.Printf("%s satisfies constraints '%s'", v, c)
}
```

Supported operators:
- `=` : you accept that exact version
- `!=` : not equal
- `>` : you accept any version higher than the one you specify
- `>=` : you accept any version equal to or higher than the one you specify
- `<` : you accept any version lower to the one you specify
- `<=` : you accept any version equal or lower to the one you specify
- `^` : it will only do updates that do not change the leftmost non-zero number.
    - e.g. `^1.2.3` := `>=1.2.3, <2.0.0`
- `~` : allows patch-level changes if a minor version is specified on the comparator. Allows minor-level changes if not.
    - e.g. `~1.2.3` := `>=1.2.3, <1.3.0`
- `~>` : you accept any version equal to or greater than in the last digit
    - e.g. `~>3.0.3` := `>= 3.0.3, < 3.1`
    
**NOTE** : `version` package doesn't support wildcards such as `x`, `X`, and `*`.

#### Pre-release
Unlike the `semver` package, `version` package always includes pre-release versions even with no pre-releases constraint.

```
v, _ := version.Parse("2.1.0.1-alpha")
c, _ := version.NewConstraints("> 2.0.0")

c.Check(v) // true
```

#### Zero Padding
Unlike the `semver` package, `version` package fills in the missing versions with 0.
In short, `3.1.3` doesn't satisfy `= 3` because `= 3` is converted to `= 3.0.0`.

```
v, _ := version.Parse("3.1.3")
c, _ := version.NewConstraints("= 3")

c.Check(v) // false
```

## Constraints

### Wildcards
The x, X, and * characters can be used as a wildcard character.
This works for all comparison operators.
When used on the = operator it falls back to the patch level comparison (see tilde below). For example,

- `2.0.x` := `>= 2.0.0, < 2.1.0`
- `>= 1.2.x` := `>= 1.2.0`
- `<= 3.x` := `< 4.0.0`
- `*` := `>= 0.0.0`

### Tilde Range
Allows patch-level changes if a minor version is specified on the comparator. Allows minor-level changes if not.
If it consists of more than 3 numbers, it allows changes that increases the last revision number. 


- `~1.2.3` := `>=1.2.3 <1.(2+1).0` := `>=1.2.3 <1.3.0`
- `~1.2` := `>=1.2.0 <1.(2+1).0` := `>=1.2.0 <1.3.0` (Same as `1.2.x`)
- `~1` := `>=1.0.0 <(1+1).0.0` := `>=1.0.0 <2.0.0` (Same as `1.x`)
- `~0.2.3` := `>=0.2.3 <0.(2+1).0` := `>=0.2.3 <0.3.0`
- `~0.2` := `>=0.2.0 <0.(2+1).0` := `>=0.2.0 <0.3.0` (Same as `0.2.x`)
- `~0` := `>=0.0.0 <(0+1).0.0` := `>=0.0.0 <1.0.0` (Same as `0.x`)
- `~1.2.3-beta.2` := `>=1.2.3-beta.2 <1.3.0`
- `~0.0.0.4` := `>=0.0.0.4 <0.0.1`

### Caret Range
Allows changes that do not modify the left-most non-zero digit.

- `^1.2.3` := `>=1.2.3 <2.0.0`
- `^0.2.3` := `>=0.2.3 <0.3.0`
- `^0.0.3` := `>=0.0.3 <0.0.4`
- `^1.2.3-beta.2` := `>=1.2.3-beta.2 <2.0.0`
- `^0.0.0.4` := `>=0.0.0.4 <0.0.0.5`

### Pessimistic Range
It means the version must be equal to or greater than in the last digit.

- `~>1.2.3` := `>=1.2.3 <1.3.0`
- `~>1.2` := `>=1.2.0 <2.0` (*different from `~`*)
- `~>1` := `>=1.0.0 <2.0.0`
- `~>1.2.3-beta.2` := `>=1.2.3-beta.2 <1.3.0`
- `~>0.0.0.4` := `>=0.0.0.4 <0.0.1`
