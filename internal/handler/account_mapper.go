package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// recentResultsLimit caps how many finalized rows the dashboard shows.
const recentResultsLimit = 3

// toDashboardVD assembles the A1 dashboard from the competitor's full
// entry list. It derives the up-next card (nearest non-finalized run) and
// the recent-results cluster (latest finalized runs, each scored).
func toDashboardVD(r *http.Request, st *store.Store, c db.Competitor, entries []db.ListEntriesByHandlerRow, dogCount, openCh int) account.DashboardViewData {
	finalizedCount := 0
	recent := make([]account.RecentRow, 0, recentResultsLimit)
	for _, e := range entries {
		if e.Status != "finalized" {
			continue
		}
		finalizedCount++
		if len(recent) < recentResultsLimit {
			pts, passed, ok := evalFinalizedScore(r, st, e.Discipline, e.Level, e.TemplateVersion, e.ID)
			recent = append(recent, account.RecentRow{
				EntryID:   e.ID,
				Title:     e.DogName + " · " + disciplineLevelLabel(e.Discipline, e.Level),
				Sub:       e.EventName + " · " + shortDate(e.TrialDate),
				Points:    pts,
				Qualified: passed,
				HasScore:  ok,
			})
		}
	}

	return account.DashboardViewData{
		DisplayName:    c.DisplayName,
		Handle:         c.Handle,
		DogCount:       dogCount,
		FinalizedCount: finalizedCount,
		OpenChallenges: openCh,
		UpNext:         pickUpNext(entries),
		Recent:         recent,
	}
}

// pickUpNext chooses the dashboard's prominent entry: the soonest
// non-finalized run dated today or later; failing that, the most recent
// non-finalized run. Returns nil when every entry is finalized.
func pickUpNext(entries []db.ListEntriesByHandlerRow) *account.UpNextCard {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	var upcoming, recent *db.ListEntriesByHandlerRow
	for i := range entries {
		e := &entries[i]
		if e.Status == "finalized" {
			continue
		}
		date := e.TrialDate.UTC()
		if !date.Before(today) {
			// Future/today: keep the earliest.
			if upcoming == nil || date.Before(upcoming.TrialDate.UTC()) {
				upcoming = e
			}
		}
		// Past non-finalized: keep the latest (list is newest-first, so the
		// first one we see wins).
		if recent == nil {
			recent = e
		}
	}
	chosen := upcoming
	if chosen == nil {
		chosen = recent
	}
	if chosen == nil {
		return nil
	}

	meta := disciplineLevelLabel(chosen.Discipline, chosen.Level) +
		" · " + chosen.TrialDate.UTC().Format("Mon 2 Jan") +
		" · " + entryNumberLabel(chosen.EntryNumber)
	label, kind := "Upcoming", "wait"
	if chosen.Status == "scoring" {
		label, kind = "Scoring", "scoring"
	}
	return &account.UpNextCard{
		EntryID:     chosen.ID,
		EventName:   chosen.EventName,
		Meta:        meta,
		DogName:     chosen.DogName,
		EventKey:    disciplineKey(chosen.Discipline),
		StatusLabel: label,
		StatusKind:  kind,
	}
}

// profileVD builds the profile editor view from a competitor row.
func profileVD(c db.Competitor, saved bool, errMsg string) account.ProfileViewData {
	return account.ProfileViewData{
		DisplayName: c.DisplayName,
		Handle:      c.Handle,
		Bio:         c.Bio,
		PublicURL:   "/competitors/" + c.Handle,
		Saved:       saved,
		Err:         errMsg,
	}
}

// profileVDFrom builds the profile view from submitted values, so a save
// re-render reflects exactly what the user typed.
func profileVDFrom(displayName, handle, bio string, saved bool, errMsg string) account.ProfileViewData {
	return account.ProfileViewData{
		DisplayName: displayName,
		Handle:      handle,
		Bio:         bio,
		PublicURL:   "/competitors/" + handle,
		Saved:       saved,
		Err:         errMsg,
	}
}

// toDogsListVD maps the store's dog list items into the A3 view, building
// each dog's "breed · age · activity" sub-line.
func toDogsListVD(items []store.DogListItem) account.DogsListViewData {
	cards := make([]account.DogCard, 0, len(items))
	for _, it := range items {
		cards = append(cards, account.DogCard{
			ID:        it.Dog.ID,
			CallName:  it.Dog.CallName,
			RegNo:     it.Dog.RegistrationNumber,
			Meta:      dogListMeta(it),
			PublicURL: dogPublicURL(it.Dog.ID),
		})
	}
	return account.DogsListViewData{Count: len(cards), Dogs: cards}
}

// dogListMeta composes the dog card sub-line: breed, age, and an activity
// hint (last-ran phrase, or "no entries yet" for a dog that has never been
// entered).
func dogListMeta(it store.DogListItem) string {
	parts := []string{}
	if it.Dog.Breed != "" {
		parts = append(parts, it.Dog.Breed)
	}
	if age := ageLabel(it.Dog.DateOfBirth); age != "" {
		parts = append(parts, age)
	}
	switch {
	case it.LastCompeted != nil:
		parts = append(parts, "last ran "+relativeTime(*it.LastCompeted))
	case it.EntryCount == 0:
		parts = append(parts, "no entries yet")
	}
	return strings.Join(parts, " · ")
}

// dogFormVD prefills the edit form from a dog row.
func dogFormVD(dog db.Dog) account.DogFormViewData {
	dob := ""
	if dog.DateOfBirth.Valid {
		dob = dog.DateOfBirth.Time.UTC().Format("2006-01-02")
	}
	return account.DogFormViewData{
		IsEdit:         true,
		DogID:          dog.ID,
		CallName:       dog.CallName,
		RegisteredName: dog.RegisteredName,
		Breed:          dog.Breed,
		DOB:            dob,
		RegNo:          dog.RegistrationNumber,
		PublicURL:      dogPublicURL(dog.ID),
	}
}

// parseDogForm reads and validates the dog form fields. base carries the
// IsEdit / DogID / PublicURL context so the returned view re-renders under
// the right URL. Returns ok=false with a populated view (values + error)
// when validation fails.
func parseDogForm(r *http.Request, base account.DogFormViewData) (store.DogInput, account.DogFormViewData, bool) {
	vd := base
	if err := r.ParseForm(); err != nil {
		vd.Err = "Could not read the form. Please try again."
		return store.DogInput{}, vd, false
	}
	callName := strings.TrimSpace(r.FormValue("call_name"))
	registered := strings.TrimSpace(r.FormValue("registered_name"))
	breed := strings.TrimSpace(r.FormValue("breed"))
	regNo := strings.TrimSpace(r.FormValue("registration_number"))
	dobStr := strings.TrimSpace(r.FormValue("date_of_birth"))

	vd.CallName = callName
	vd.RegisteredName = registered
	vd.Breed = breed
	vd.RegNo = regNo
	vd.DOB = dobStr

	if callName == "" {
		vd.Err = "Call name is required."
		return store.DogInput{}, vd, false
	}

	var dob *time.Time
	if dobStr != "" {
		t, err := time.Parse("2006-01-02", dobStr)
		if err != nil {
			vd.Err = "Date of birth must be a valid date."
			return store.DogInput{}, vd, false
		}
		dob = &t
	}

	return store.DogInput{
		CallName:           callName,
		RegisteredName:     registered,
		Breed:              breed,
		DateOfBirth:        dob,
		RegistrationNumber: regNo,
	}, vd, true
}

// dogPublicURL is the public profile path for a dog id (P7).
func dogPublicURL(id int64) string {
	return fmt.Sprintf("/dogs/%d", id)
}
