package cmd

import (
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
)

func indentHeredoc(raw string) string {
	d := heredoc.Doc(raw + ".")
	return d[:len(d)-2]
}

func flushHeredoc(raw string) string {
	return strings.TrimSuffix(heredoc.Doc(raw), "\n")
}
