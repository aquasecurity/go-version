package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/aquasecurity/go-version/pkg/version"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "compare",
				Aliases: []string{"c"},
				Usage:   "compare two versions",
				Action: func(c *cli.Context) error {
					s1 := c.Args().Get(0)
					v1, err := version.Parse(s1)
					if err != nil {
						log.Fatalf("failed to parse version (%s): %s", s1, err)
					}

					s2 := c.Args().Get(1)
					v2, err := version.Parse(s2)
					if err != nil {
						log.Fatalf("failed to parse version (%s): %s", s2, err)
					}

					fmt.Println(v1.Compare(v2))
					return nil
				},
			},
			{
				Name:    "satisfy",
				Aliases: []string{"s"},
				Usage:   "check if the version satisfies the constraint",
				Action: func(ctx *cli.Context) error {
					s1 := ctx.Args().Get(0)
					v, err := version.Parse(s1)
					if err != nil {
						log.Fatalf("failed to parse version (%s): %s", s1, err)
					}

					s2 := ctx.Args().Get(1)
					c, err := version.NewConstraints(s2)
					if err != nil {
						log.Fatalf("failed to parse version (%s): %s", s2, err)
					}

					fmt.Println(c.Check(v))
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
