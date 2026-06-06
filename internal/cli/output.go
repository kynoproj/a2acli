package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// printJSON writes v as indented JSON to w with a trailing newline.
func printJSON(w io.Writer, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	_, err = w.Write([]byte{'\n'})
	return err
}
