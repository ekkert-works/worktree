package worktree_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDependenciesPointInward(t *testing.T) {
	const core = "github.com/ekkert-works/worktree/internal/worktree"
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, declaration := range file.Imports {
			dependency, err := strconv.Unquote(declaration.Path.Value)
			if err != nil {
				return err
			}
			if filepath.Dir(path) == "." && dependency != "context" && dependency != "errors" && dependency != "sort" && dependency != "strings" {
				t.Errorf("core file %s imports %s; review core dependency allowlist", path, dependency)
			}
			if strings.HasPrefix(dependency, core+"/") {
				t.Errorf("%s imports adapter %s", path, dependency)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
