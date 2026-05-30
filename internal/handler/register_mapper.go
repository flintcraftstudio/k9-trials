package handler

import (
	"strings"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// selectableTrials returns the event trials that still accept
// registrations: those not yet started (status "pending"). In-progress and
// complete trials are closed to new entries.
func selectableTrials(ewt store.EventWithTrials) []db.Trial {
	out := make([]db.Trial, 0, len(ewt.Trials))
	for _, t := range ewt.Trials {
		if t.Status == "pending" {
			out = append(out, t)
		}
	}
	return out
}

// regDogMeta composes the "breed · age · regno" sub-line for a dog choice.
func regDogMeta(dog db.Dog) string {
	parts := []string{}
	if dog.Breed != "" {
		parts = append(parts, dog.Breed)
	}
	if age := ageLabel(dog.DateOfBirth); age != "" {
		parts = append(parts, age)
	}
	if dog.RegistrationNumber != "" {
		parts = append(parts, dog.RegistrationNumber)
	}
	return strings.Join(parts, " · ")
}

// regEventKey picks the page-level accent for the registration view from
// the first selectable trial, defaulting to obedience.
func regEventKey(trials []db.Trial) string {
	if len(trials) > 0 {
		return disciplineKey(trials[0].Discipline)
	}
	return "obedience"
}

// buildRegisterVD assembles the R1 form view: dog radios (selectedDogID
// marked), and the trial checklist for that dog with already-registered
// trials disabled. checked carries the trial ids to re-check after a
// validation error.
func buildRegisterVD(ewt store.EventWithTrials, dogs []db.Dog, selectedDogID int64, registered, checked map[int64]bool, notes, errMsg string) account.RegisterViewData {
	trials := selectableTrials(ewt)

	dogOpts := make([]account.RegDogOption, 0, len(dogs))
	for _, dog := range dogs {
		dogOpts = append(dogOpts, account.RegDogOption{
			ID:       dog.ID,
			CallName: dog.CallName,
			Meta:     regDogMeta(dog),
			Selected: dog.ID == selectedDogID,
		})
	}

	return account.RegisterViewData{
		EventName: ewt.Event.Name,
		EventSlug: ewt.Event.Slug,
		DateRange: dateRange(ewt.Event.StartDate, ewt.Event.EndDate),
		EventKey:  regEventKey(trials),
		Dogs:      dogOpts,
		Trials:    regTrialOptions(trials, registered, checked),
		Notes:     notes,
		Err:       errMsg,
	}
}

// regTrialOptions builds the trial checklist, disabling trials the selected
// dog already holds a registration in.
func regTrialOptions(trials []db.Trial, registered, checked map[int64]bool) []account.RegTrialOption {
	out := make([]account.RegTrialOption, 0, len(trials))
	for _, t := range trials {
		out = append(out, account.RegTrialOption{
			ID:       t.ID,
			Label:    disciplineLevelLabel(t.Discipline, t.Level),
			Date:     t.TrialDate.UTC().Format("Mon 2 Jan"),
			EventKey: disciplineKey(t.Discipline),
			Disabled: registered[t.ID],
			Checked:  checked[t.ID] && !registered[t.ID],
		})
	}
	return out
}
