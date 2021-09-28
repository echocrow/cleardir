package cleardir

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"

	os "github.com/echocrow/osa"
)

func ParseClearables(custPath string) (
	clearables []string,
	path string,
	err error,
) {
	path = custPath
	useDefault := path == ""
	if path == "" {
		path, err = defaultCfgPath()
		if err != nil {
			return
		}
	}

	clearables, err = readCfgLines(path)

	if useDefault && os.IsNotExist(err) {
		clearables = []string{}
		err = nil
	}

	return
}

func defaultCfgPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	cfgPath := filepath.Join(cfgDir, "cleardir", "clearignore")
	return cfgPath, nil
}

func readCfgLines(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(b)
	scanner := bufio.NewScanner(r)
	lines := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}
