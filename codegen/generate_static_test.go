package codegen

import (
	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// generateStatic drives the unified component codegen with an empty script,
// standing in for the retired static Generate entrypoint. Tests that exercise
// layout and style emission use it; the parsed-script argument is ignored (the
// legacy path never emitted handlers) and the component is always named App.
func generateStatic(doc *template.Document, _ *script.Script, ss *style.Stylesheet, pkg string) ([]byte, error) {
	return GenerateComponent(doc, "", ss, ComponentOptions{
		PackageName:   pkg,
		ComponentName: "App",
	})
}
