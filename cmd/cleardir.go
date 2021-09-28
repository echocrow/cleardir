package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	os "github.com/echocrow/osa"
	"github.com/spf13/cobra"
)

type cleardirCmd struct {
	cmd  *cobra.Command
	opts cleardirOpts
}

type cleardirOpts struct {
}

// Execute executes the root command.
func Execute(version string) {
	cmd := newCleardirCmd().cmd
	cmd.Version = version
	cmd.SetIn(os.Stdin)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newCleardirCmd() *cleardirCmd {
	root := &cleardirCmd{}
	opts := &root.opts

	cmd := &cobra.Command{
		Use:   "cleardir [PATH]",
		Short: "Clear empty directories",
		Long: heredoc.Doc(`
			Cleardir finds and deletes empty folders. Folders are considered empty
			when there are either no files located inside a given folder (including
			subfolders), or only white-listed files that are safe for deletion.
		`),
		Example: indentHeredoc(`
		  cleardir
		  cleardir some/other/path
		  cleardir -y -s
		  cleardir --dry
		`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleardir(cmd, opts, args)
		},
	}

	cmd.SetOut(os.Stdout)

	root.cmd = cmd
	return root
}

func runCleardir(cmd *cobra.Command, opts *cleardirOpts, args []string) error {
	// @todo.
	return nil
}
