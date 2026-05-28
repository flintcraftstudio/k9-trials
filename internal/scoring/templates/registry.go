package templates

import "github.com/flintcraftstudio/k9-trials/internal/scoring"

// Lookup returns the ScoresheetTemplate for the given (discipline, level,
// version) triple. Returns false when no template is registered.
//
// The version match is exact — historical scoresheets pin a specific
// rulebook revision via trials.template_version, so a 2026.1 entry must
// keep evaluating against the 2026.1 encoding even after a 2026.2 ships.
func Lookup(discipline scoring.Discipline, level scoring.Level, version string) (scoring.ScoresheetTemplate, bool) {
	switch {
	case discipline == scoring.DisciplineOB && level == scoring.LevelOne && version == "2026.1":
		return L1OB(), true
	}
	return scoring.ScoresheetTemplate{}, false
}
