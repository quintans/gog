package plugins

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quintans/gog/generator"
)

func run(t *testing.T, in, want string) {
	t.Helper()

	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "src.go", in, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	code, err := generator.InspectGoFile(f).GenerateCode("src_gog.go")
	if err != nil {
		t.Fatal(err)
	}
	got := string(code)
	if got != want {
		t.Errorf("\ngot ----------\n%swant ++++++++++\n%sdiff =========\n%s", got, want, cmp.Diff(got, want))
	}
}
