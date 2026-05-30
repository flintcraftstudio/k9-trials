package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// RegisterPage serves GET /events/{slug}/register — the competitor
// registration form (R1). It resolves the event, the competitor's dogs,
// and (for the selected dog) which trials are already entered, then renders
// the form or the no-dogs / not-open edge state.
func RegisterPage(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		ewt, ok := loadRegistrableEvent(w, r, st)
		if !ok {
			return
		}

		// Unpublished events are not accepting registrations.
		if !registrationOpen(ewt.Event.Status) {
			renderPublic(w, r, account.RegisterPage(registerNotOpenVD(ewt)))
			return
		}

		dogs, err := st.OwnerDogs(r.Context(), c.ID)
		if err != nil {
			slog.Error("register dogs", "competitor", c.ID, "err", err)
			http.Error(w, "registration unavailable", http.StatusInternalServerError)
			return
		}
		if len(dogs) == 0 {
			renderPublic(w, r, account.RegisterPage(registerNoDogsVD(ewt)))
			return
		}

		selected := selectedDog(r, dogs)
		registered, err := st.RegisteredTrialIDsForDog(r.Context(), selected.ID)
		if err != nil {
			slog.Error("register dog trials", "dog", selected.ID, "err", err)
			http.Error(w, "registration unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, account.RegisterPage(buildRegisterVD(ewt, dogs, selected.ID, registered, nil, "", "")))
	}
}

// RegisterTrials serves GET /events/{slug}/register/trials?dog={id} — the
// htmx fragment that re-renders the trial checklist when the competitor
// picks a different dog, so trials that dog already holds show as disabled.
func RegisterTrials(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		ewt, ok := loadRegistrableEvent(w, r, st)
		if !ok {
			return
		}
		dogID, err := strconv.ParseInt(r.URL.Query().Get("dog"), 10, 64)
		if err != nil || dogID <= 0 {
			http.NotFound(w, r)
			return
		}
		// Confirm the dog belongs to the competitor before reflecting its
		// registrations back.
		if _, err := st.GetOwnerDog(r.Context(), dogID, c.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			slog.Error("register trials dog", "dog", dogID, "err", err)
			http.Error(w, "registration unavailable", http.StatusInternalServerError)
			return
		}
		registered, err := st.RegisteredTrialIDsForDog(r.Context(), dogID)
		if err != nil {
			slog.Error("register trials lookup", "dog", dogID, "err", err)
			http.Error(w, "registration unavailable", http.StatusInternalServerError)
			return
		}
		trials := selectableTrials(ewt)
		vd := account.RegisterViewData{
			EventKey: regEventKey(trials),
			Trials:   regTrialOptions(trials, registered, nil),
		}
		renderPublic(w, r, account.TrialOptions(vd))
	}
}

// RegisterSubmit serves POST /events/{slug}/register — files a pending
// registration for each selected trial, then renders the confirmation.
// Validation failures re-render the form with the offending message.
func RegisterSubmit(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		u := session.FromContext(r.Context())
		ewt, ok := loadRegistrableEvent(w, r, st)
		if !ok {
			return
		}
		if !registrationOpen(ewt.Event.Status) {
			renderPublic(w, r, account.RegisterPage(registerNotOpenVD(ewt)))
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		dogs, err := st.OwnerDogs(r.Context(), c.ID)
		if err != nil {
			slog.Error("register submit dogs", "competitor", c.ID, "err", err)
			http.Error(w, "registration unavailable", http.StatusInternalServerError)
			return
		}
		if len(dogs) == 0 {
			renderPublic(w, r, account.RegisterPage(registerNoDogsVD(ewt)))
			return
		}

		dogID, _ := strconv.ParseInt(r.FormValue("dog"), 10, 64)
		selected, found := dogByID(dogs, dogID)
		notes := r.FormValue("notes")

		fail := func(sel db.Dog, msg string) {
			checked := idSet(r.Form["trial"])
			registered, _ := st.RegisteredTrialIDsForDog(r.Context(), sel.ID)
			renderPublic(w, r, account.RegisterPage(buildRegisterVD(ewt, dogs, sel.ID, registered, checked, notes, msg)))
		}

		if !found {
			// No valid dog selected: default to the first so the re-render
			// has a coherent trial list.
			fail(dogs[0], "Choose which dog you are registering.")
			return
		}

		// Only trials that belong to this event and still accept entries
		// are valid targets.
		valid := make(map[int64]bool)
		for _, t := range selectableTrials(ewt) {
			valid[t.ID] = true
		}
		var wanted []int64
		for _, raw := range r.Form["trial"] {
			id, err := strconv.ParseInt(raw, 10, 64)
			if err == nil && valid[id] {
				wanted = append(wanted, id)
			}
		}
		if len(wanted) == 0 {
			fail(selected, "Select at least one trial to enter.")
			return
		}

		registered, err := st.RegisteredTrialIDsForDog(r.Context(), selected.ID)
		if err != nil {
			slog.Error("register submit lookup", "dog", selected.ID, "err", err)
			http.Error(w, "registration unavailable", http.StatusInternalServerError)
			return
		}

		created := 0
		for _, trialID := range wanted {
			if registered[trialID] {
				continue
			}
			if _, err := st.CreateRegistration(r.Context(), trialID, c.ID, selected.ID, u.ID, notes); err != nil {
				// A racing duplicate trips the (trial_id, dog_id) UNIQUE
				// constraint; skip it rather than failing the batch.
				if isUniqueViolation(err) {
					continue
				}
				slog.Error("create registration", "trial", trialID, "dog", selected.ID, "err", err)
				fail(selected, "Something went wrong. Please try again.")
				return
			}
			created++
		}

		if created == 0 {
			fail(selected, selected.CallName+" is already registered for the selected trials.")
			return
		}

		renderPublic(w, r, account.RegisterDonePage(account.RegisterDoneViewData{
			EventName: ewt.Event.Name,
			EventSlug: ewt.Event.Slug,
			DogName:   selected.CallName,
			Count:     created,
		}))
	}
}

// loadRegistrableEvent loads the event by {slug} and rejects drafts and
// archived events as not-found (they are not public). Returns ok=false
// after writing the response on any miss.
func loadRegistrableEvent(w http.ResponseWriter, r *http.Request, st *store.Store) (store.EventWithTrials, bool) {
	slug := r.PathValue("slug")
	ewt, err := st.LoadPublicEvent(r.Context(), slug)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return store.EventWithTrials{}, false
	}
	if err != nil {
		slog.Error("register event load", "slug", slug, "err", err)
		http.Error(w, "registration unavailable", http.StatusInternalServerError)
		return store.EventWithTrials{}, false
	}
	if ewt.Event.Status == "draft" || ewt.Event.Status == "archived" {
		http.NotFound(w, r)
		return store.EventWithTrials{}, false
	}
	return ewt, true
}

// selectedDog resolves the dog the form should show as selected: the ?dog=
// query value when it names an owned dog, otherwise the first dog.
func selectedDog(r *http.Request, dogs []db.Dog) db.Dog {
	if id, err := strconv.ParseInt(r.URL.Query().Get("dog"), 10, 64); err == nil {
		if dog, ok := dogByID(dogs, id); ok {
			return dog
		}
	}
	return dogs[0]
}

// dogByID finds a dog in the slice by id.
func dogByID(dogs []db.Dog, id int64) (db.Dog, bool) {
	for _, d := range dogs {
		if d.ID == id {
			return d, true
		}
	}
	return db.Dog{}, false
}

// idSet parses a slice of decimal id strings into a set, dropping invalid
// entries.
func idSet(raw []string) map[int64]bool {
	set := make(map[int64]bool, len(raw))
	for _, s := range raw {
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			set[id] = true
		}
	}
	return set
}

// registerNoDogsVD builds the no-dogs edge-state view.
func registerNoDogsVD(ewt store.EventWithTrials) account.RegisterViewData {
	return account.RegisterViewData{
		EventName: ewt.Event.Name,
		EventSlug: ewt.Event.Slug,
		DateRange: dateRange(ewt.Event.StartDate, ewt.Event.EndDate),
		EventKey:  regEventKey(selectableTrials(ewt)),
		NoDogs:    true,
	}
}

// registerNotOpenVD builds the not-open edge-state view for an event that
// is not accepting registrations.
func registerNotOpenVD(ewt store.EventWithTrials) account.RegisterViewData {
	return account.RegisterViewData{
		EventName:  ewt.Event.Name,
		EventSlug:  ewt.Event.Slug,
		DateRange:  dateRange(ewt.Event.StartDate, ewt.Event.EndDate),
		EventKey:   regEventKey(selectableTrials(ewt)),
		NotOpen:    true,
		NotOpenMsg: "The organizers have not opened registration for " + ewt.Event.Name + " yet. Check back soon.",
	}
}
