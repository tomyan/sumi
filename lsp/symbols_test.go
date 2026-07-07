package lsp

import "testing"

// symbolsFixture exercises functions, state decls, id elements and a
// component usage.
const symbolsFixture = "<script>\n" +
	"count := sumi.New(0)\n" +
	"func increment() { count.Set(count.Get() + 1) }\n" +
	"</script>\n" +
	"<div id=\"root\">\n<Widget />\n</div>\n"

func symbolByName(syms []SymbolInformation, name string) (SymbolInformation, bool) {
	for _, s := range syms {
		if s.Name == name {
			return s, true
		}
	}
	return SymbolInformation{}, false
}

func TestDocumentSymbolsFunction(t *testing.T) {
	// When
	syms := DocumentSymbols(symbolsFixture, "file:///c.sumi")

	// Then
	s, ok := symbolByName(syms, "increment")
	if !ok {
		t.Fatalf("no symbol for func increment (%+v)", syms)
	}
	if s.Kind != SymbolFunction {
		t.Errorf("increment kind = %d, want %d", s.Kind, SymbolFunction)
	}
	if s.Location.URI != "file:///c.sumi" {
		t.Errorf("location URI = %q", s.Location.URI)
	}
}

func TestDocumentSymbolsState(t *testing.T) {
	// When
	syms := DocumentSymbols(symbolsFixture, "file:///c.sumi")

	// Then
	s, ok := symbolByName(syms, "count")
	if !ok {
		t.Fatalf("no symbol for state count (%+v)", syms)
	}
	if s.Kind != SymbolVariable {
		t.Errorf("count kind = %d, want %d", s.Kind, SymbolVariable)
	}
}

func TestDocumentSymbolsIDAndComponent(t *testing.T) {
	// When
	syms := DocumentSymbols(symbolsFixture, "file:///c.sumi")

	// Then
	if _, ok := symbolByName(syms, "root"); !ok {
		t.Errorf("no symbol for id=root (%+v)", syms)
	}
	w, ok := symbolByName(syms, "Widget")
	if !ok {
		t.Fatalf("no symbol for <Widget/> (%+v)", syms)
	}
	if w.Kind != SymbolObject {
		t.Errorf("Widget kind = %d, want %d", w.Kind, SymbolObject)
	}
}
