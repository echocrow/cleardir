package cmd

import (
	"bufio"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

func confirm(cmd *cobra.Command, msg string, attempts int, resolve bool) bool {
	var res string
	var err error
	r := bufio.NewReader(cmd.InOrStdin())
	for ; attempts > 0; attempts-- {
		cmd.Printf("%s [y/N]: ", msg)
		if resolve {
			res = "y"
			cmd.Println(res)
		} else {
			res, err = r.ReadString('\n')
			res = strings.ToLower(strings.TrimSpace(res))
		}
		if err != nil {
			return false
		}
		if len(res) > 0 && len(res) <= 4 {
			yes := res[0] == 'y'
			no := !yes && res[0] == 'n'
			if yes || no {
				return yes
			}
		}
	}
	return false
}

func indentHeredoc(raw string) string {
	d := heredoc.Doc(raw + ".")
	return d[:len(d)-2]
}

func flushHeredoc(raw string) string {
	return strings.TrimSuffix(heredoc.Doc(raw), "\n")
}
