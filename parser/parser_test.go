package parser

import (
	"go/token"
	"testing"
)

var validFiles = []string{
	"../sass-spec/spec/basic/00_empty/input.scss",
	"../sass-spec/spec/sass/imported/imported.sass",
	"../sass-spec/spec/libsass/bourbon/lib/_bourbon.scss",
}

func TestParse(t *testing.T) {
	for _, name := range validFiles {
		_, err := ParseFile(token.NewFileSet(), name, nil, DeclarationErrors)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", name, err)
		}
	}
}

func TestParseDir(t *testing.T) {
	// paths := "../sass-spec/spec/basic/00_empty"
}
