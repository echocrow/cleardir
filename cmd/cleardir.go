package cmd

import (
	"errors"
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/echocrow/cleardir/pkg/cleardir"
	os "github.com/echocrow/osa"
	"github.com/spf13/cobra"
)

type cleardirCmd struct {
	cmd  *cobra.Command
	opts cleardirOpts
}

type cleardirOpts struct {
	cfg      string
	maxDepth int
	trivials []string
	dry      bool
	silent   bool
	yes      bool
}

const (
	getCfgFlag = "?"
)

// NewCmd creates a new cleardir command.
func NewCmd(version string) *cobra.Command {
	cmd := newCleardirCmd().cmd
	cmd.Version = version
	cmd.SetIn(os.Stdin)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	return cmd
}

// Execute executes the root command.
func Execute(version string) {
	cmd := NewCmd(version)
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

	cmd.Flags().StringVarP(&opts.cfg, "config", "c", "", "specify the configuration file path")
	cmd.Flags().StringSliceVarP(&opts.trivials, "files", "f", nil, "list files that can be deleted safely")
	cmd.Flags().IntVarP(&opts.maxDepth, "max-depth", "d", -1, flushHeredoc(`
		limit how many sub-directories to descend to at most;
		use "-1" for no limit
	`))
	cmd.Flags().BoolVarP(&opts.dry, "dry", "", false, "only list clearable files and directories")
	cmd.Flags().BoolVarP(&opts.yes, "yes", "y", false, "skip and confirm prompts")
	cmd.Flags().BoolVarP(&opts.silent, "silent", "s", false, "silence standard output; implies \"-y\"")

	root.cmd = cmd
	return root
}

func runCleardir(cmd *cobra.Command, opts *cleardirOpts, args []string) error {
	rawDir := ""
	if len(args) >= 1 {
		rawDir = args[0]
	}

	dir, err := filepath.Abs(rawDir)
	if err != nil {
		return err
	}

	getCfgPath := false
	if opts.cfg == getCfgFlag {
		getCfgPath = true
		opts.cfg = ""
	}
	trivials, cfgPath, err := cleardir.ParseClearables(opts.cfg)
	if err != nil {
		return err
	}
	if getCfgPath {
		cmd.Println(cfgPath)
		return nil
	}

	trivials = append(trivials, opts.trivials...)

	matches := make(chan string)
	go func() {
		err = cleardir.FindClearables(
			matches,
			dir,
			trivials,
			opts.maxDepth,
		)
		close(matches)
	}()
	dels := []string{}
	for m := range matches {
		if !opts.silent {
			cmd.Printf("- %s\n", m)
		}
		dels = append(dels, m)
	}
	if err != nil {
		return err
	}

	if len(dels) == 0 {
		if !opts.silent {
			cmd.Println("All clear!")
		}
		return nil
	} else {
		if !opts.silent {
			cmd.Printf("Can clear %d files.\n", len(dels))
		}
	}

	if opts.dry {
		return nil
	}

	ok := opts.silent || confirm(cmd, "Continue?", 1, opts.yes)
	if !ok {
		cmd.SilenceUsage = true
		return errors.New("Aborted")
	}

	err = cleardir.Remove(dels...)
	if err != nil {
		return err
	}

	return nil
}
