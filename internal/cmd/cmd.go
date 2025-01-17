package cmd

import (
	"akapurgo/internal/cmd/run"
	"strings"

	"github.com/spf13/cobra"
)

const (
	descriptionShort = `Akapurgo is a webserver to purge Akamai paths.`
	descriptionLong  = `
	Akapurgo is a webserver to purge Akama paths. 
	`
)

func NewRootCommand(name string) *cobra.Command {
	c := &cobra.Command{
		Use:   name,
		Short: descriptionShort,
		Long:  strings.ReplaceAll(descriptionLong, "\t", ""),
	}

	c.AddCommand(
		run.NewCommand(),
	)

	return c
}
