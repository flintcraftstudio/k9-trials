package view

import (
	"strings"
	"unicode"

	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/view/components"
)

// sectionAnchorsFor returns the role-aware chips shown in the top bar's
// actions area. Anonymous visitors get none. Each role currently maps to
// a single chip pointing at its primary section; if multi-role users
// land later, this returns a slice naturally — the NavBar already loops.
func sectionAnchorsFor(u *session.User) []components.NavBarLink {
	if u == nil {
		return nil
	}
	initials := initialsFromEmail(u.Email)
	switch u.Role {
	case "admin":
		return []components.NavBarLink{
			{ID: "admin", Href: "/admin", Label: "Admin", Initials: initials},
		}
	case "judge":
		return []components.NavBarLink{
			{ID: "judge", Href: "/judge", Label: "Judge", Initials: initials},
		}
	case "competitor":
		return []components.NavBarLink{
			{ID: "account", Href: "/account", Label: "My account", Initials: initials},
		}
	}
	return nil
}

// initialsFromEmail picks up to two letters from the local part of an
// email address to fill the role-chip avatar bubble. We split on common
// separators ("." "_" "-" "+") and take the first letter of the first
// two tokens; falls back to the first two letters of the local part if
// there's no separator. Session.User has no display_name today — once
// the competitor profile join is wired in, prefer that.
func initialsFromEmail(email string) string {
	local, _, _ := strings.Cut(email, "@")
	if local == "" {
		return "?"
	}
	parts := strings.FieldsFunc(local, func(r rune) bool {
		return r == '.' || r == '_' || r == '-' || r == '+'
	})
	var b strings.Builder
	for _, p := range parts {
		if b.Len() >= 2 {
			break
		}
		for _, r := range p {
			if unicode.IsLetter(r) {
				b.WriteRune(unicode.ToUpper(r))
				break
			}
		}
	}
	if b.Len() < 2 {
		for _, r := range local {
			if b.Len() >= 2 {
				break
			}
			if unicode.IsLetter(r) {
				b.WriteRune(unicode.ToUpper(r))
			}
		}
	}
	if b.Len() == 0 {
		return "?"
	}
	return b.String()
}
