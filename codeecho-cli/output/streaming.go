package output

import (
	"fmt"
	"io"

	"github.com/opskraken/codeecho-cli/config"
	"github.com/opskraken/codeecho-cli/scanner"
)


type StreamingWriter interface {
	WriteHeader(repoPath string, scanTime string) error
	WriteTree(paths []string) error
	WriteFile(file *scanner.FileInfo) error
	WriteFooter(stats *scanner.StreamingStats) error
	Close() error
}

// NewStreamingWriter creates the appropriate writer based on format
// Factory pattern - returns different implementations of same interface
func NewStreamingWriter(w io.Writer, format string, opts config.OutputOptions) (StreamingWriter, error) {
	switch format {
	case "xml":
		return NewStreamingXMLWriter(w, opts), nil
	case "json":
		return NewStreamingJSONWriter(w, opts), nil
	case "markdown", "md":
		return NewStreamingMarkdownWriter(w, opts), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
