package template

// slotRemovedMessage is emitted when a template uses the removed slot surface.
// Composition now runs through snippets: declare a {snippet name()} inside the
// component tag and invoke it with {render name()} in the component template.
const slotRemovedMessage = "slots were removed; declare a {snippet name()} inside the component tag and {render name()} in the component template"
