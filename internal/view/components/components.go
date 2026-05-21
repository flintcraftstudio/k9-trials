// Package components is the templ port of the K9 Elements marketing-site
// UI kit foundations. It carries the shared visual language between the
// public marketing site (github.com/flintcraftstudio/k9-elements) and the
// trials/scoring app so the two surfaces feel like one product.
//
// Only the layout-level primitives live here: buttons, logo, nav, footer,
// container, icon, eyebrow, pill, and form controls. Marketing-only
// components (Hero, TrialCard, PricingTiers, Testimonials, EventGrid)
// stay in the marketing repo.
package components

import (
	"strings"
	"time"

	"github.com/a-h/templ"
)

// ButtonProps mirrors the JSX <Button variant size href className />.
// Zero values default to variant=primary, size=md, type=button.
//
// Type only applies when Href is empty (i.e. the component renders as
// a <button>). Defaults to "button" so callers don't accidentally
// submit forms on any in-page click.
type ButtonProps struct {
	Variant  string // "primary" | "soft" | "plain" | "event"
	Size     string // "md" | "lg"
	Href     string // when set, renders as <a> instead of <button>
	Class    string // extra classes appended after the variant
	Type     string // "button" (default) | "submit" | "reset"
	Disabled bool
	Attrs    templ.Attributes // pass-through for hx-*, data-*, aria-*
}

func buttonType(t string) string {
	if t == "" {
		return "button"
	}
	return t
}

func buttonClasses(p ButtonProps) string {
	size := p.Size
	if size == "" {
		size = "md"
	}
	variant := p.Variant
	if variant == "" {
		variant = "primary"
	}
	return joinClasses("btn", size, "btn-"+variant, p.Class)
}

func pillClasses(variant string) string {
	if variant == "" {
		variant = "open"
	}
	return joinClasses("pill", "pill-"+variant)
}

func eventPillClasses(event string) string {
	return joinClasses("pill", "pill-event-"+event)
}

// pillHasDot returns true for the status pills (open/wait/closed/info)
// that show the leading dot, false for event-themed pills.
func pillHasDot(variant string) bool {
	switch variant {
	case "open", "wait", "info":
		return true
	}
	return false
}

func containerClasses(extra string) string {
	return joinClasses("container", extra)
}

// joinClasses concatenates class names with single spaces, skipping empties.
func joinClasses(parts ...string) string {
	out := parts[:0:0]
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return strings.Join(out, " ")
}

// capitalize returns s with the first byte uppercased (ASCII only,
// sufficient for the four event labels).
func capitalize(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	if b[0] >= 'a' && b[0] <= 'z' {
		b[0] -= 32
	}
	return string(b)
}

// footerYear returns the current year for the Footer copyright line.
func footerYear() int { return time.Now().Year() }

// InputProps mirrors flint-ui's input.Props shape (Tailwind v4 there;
// our markup uses v3 utilities + mist-* tokens). Zero values default to
// Type="text". Attrs spreads htmx/data-*/aria-* onto the <input>.
type InputProps struct {
	Type         string
	Name         string
	Value        string
	Placeholder  string
	ID           string
	Autocomplete string

	Required bool
	Disabled bool
	Readonly bool
	Invalid  bool

	Class string
	Attrs templ.Attributes
}

func inputResolveType(t string) string {
	if t == "" {
		return "text"
	}
	return t
}

// inputClasses composition is done inline inside input.templ so the
// literal Tailwind class strings live in a .templ file (Tailwind's
// content scanner only reads .templ; classes that live only in .go
// helpers never get picked up). Don't move them back here.

// TextareaProps mirrors InputProps for multi-line input. Rows defaults
// to 4 when zero. Class strings live inline in textarea.templ — see the
// note on inputClasses in this file.
type TextareaProps struct {
	Name        string
	Value       string
	Placeholder string
	ID          string

	Rows int

	Required bool
	Disabled bool
	Readonly bool
	Invalid  bool

	Class string
	Attrs templ.Attributes
}

func textareaResolveRows(r int) int {
	if r <= 0 {
		return 4
	}
	return r
}

// FieldProps is the shared shape for Field, FieldGroup, ErrorMessage,
// Description — each wraps children in a styled element.
type FieldProps struct {
	Class string
	Attrs templ.Attributes
}

// LabelProps adds a For target so <label for=""> wires up to the control.
type LabelProps struct {
	For   string
	Class string
	Attrs templ.Attributes
}

func fieldClasses(p FieldProps) string {
	return joinClasses("flex flex-col gap-1.5", p.Class)
}

func fieldGroupClasses(p FieldProps) string {
	return joinClasses("flex flex-col gap-5", p.Class)
}

func labelClasses(p LabelProps) string {
	return joinClasses("text-sm font-medium text-mist-950", p.Class)
}

func errorMessageClasses(p FieldProps) string {
	return joinClasses("text-sm text-[var(--color-danger)]", p.Class)
}

func descriptionClasses(p FieldProps) string {
	return joinClasses("text-sm text-mist-600", p.Class)
}
