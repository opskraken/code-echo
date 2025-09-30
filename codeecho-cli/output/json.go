package output

import (
	"encoding/json"

	"github.com/opskraken/codeecho-cli/scanner"
)

func GenerateJSONOutput(result *scanner.ScanResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
