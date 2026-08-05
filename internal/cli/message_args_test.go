package cli

import "testing"

func TestMessageArgs(t *testing.T) {
	tests := []struct {
		name    string
		src     messageSources
		args    []string
		wantErr bool
	}{
		{"text-only", messageSources{}, []string{"hello"}, false},
		{"file-only", messageSources{file: "m.json"}, nil, false},
		{"json-only", messageSources{jsonBody: `{}`}, nil, false},
		{"parts-only", messageSources{partsJSON: `[]`}, nil, false},

		{"no-source-no-text", messageSources{}, nil, true},
		{"text-plus-file", messageSources{file: "m.json"}, []string{"hello"}, true},
		{"text-plus-json", messageSources{jsonBody: `{}`}, []string{"hello"}, true},
		{"text-plus-parts", messageSources{partsJSON: `[]`}, []string{"hello"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// src is captured by pointer, mirroring how cobra invokes the
			// validator after flags are parsed.
			src := tt.src
			validate := messageArgs(&src)
			err := validate(nil, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("messageArgs err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestHasNonTextSource(t *testing.T) {
	tests := []struct {
		name string
		src  messageSources
		want bool
	}{
		{"empty", messageSources{}, false},
		{"text-only", messageSources{text: "hi"}, false},
		{"file", messageSources{file: "m.json"}, true},
		{"json", messageSources{jsonBody: `{}`}, true},
		{"parts", messageSources{partsJSON: `[]`}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.src.hasNonTextSource(); got != tt.want {
				t.Errorf("hasNonTextSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
