package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/aquasecurity/go-version/pkg/version"
)

func main() {
	compareVersions()
	sortVersions()
	satisfiedByVersion()
}

func compareVersions() {
	fmt.Println("=== compare ===")
	v1, err := version.Parse("1.2.0.9-alpha")
	if err != nil {
		log.Fatal(err)
	}

	v2, err := version.Parse("1.2.1.0+11")
	if err != nil {
		log.Fatal(err)
	}

	// Comparison example. There is also GreaterThan, Equal, and just
	// a simple Compare that returns an int allowing easy >=, <=, etc.
	if v1.LessThan(v2) {
		fmt.Printf("%s is less than %s\n", v1, v2)
	}
}

func sortVersions() {
	fmt.Println("\n=== sort ===")
	versionsRaw := []string{"1.1.0", "0.7.1", "1.4.0", "1.4.0-alpha", "1.4.1-beta", "1.4.0-alpha.2+20130313144700"}
	versions := make([]version.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		v, err := version.Parse(raw)
		if err != nil {
			log.Fatal(err)
		}
		versions[i] = v
	}

	// After this, the versions are properly sorted
	sort.Sort(version.Collection(versions))

	for _, v := range versions {
		fmt.Println(v)
	}
}

func satisfiedByVersion() {
	fmt.Println("\n=== constraint ===")
	v, err := version.Parse("2.1.0.1-alpha")
	if err != nil {
		log.Fatal(err)
	}

	c, err := version.NewConstraints(">= 1.0, < 1.4 || > 2.1")
	if err != nil {
		log.Fatal(err)
	}

	if c.Check(v) {
		fmt.Printf("%s satisfies constraints '%s'", v, c)
	}
}
