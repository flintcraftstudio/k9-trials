package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// currentCompetitor resolves the competitor identity for the logged-in
// user. When the account has no competitor row (an admin, seeded
// server-side), it renders the neutral no-profile page and returns
// ok=false so the caller stops. RequireRole has already guaranteed a
// logged-in competitor or admin by the time this runs.
func currentCompetitor(w http.ResponseWriter, r *http.Request, st *store.Store) (db.Competitor, bool) {
	u := session.FromContext(r.Context())
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return db.Competitor{}, false
	}
	c, err := st.CurrentCompetitor(r.Context(), u.ID)
	if errors.Is(err, sql.ErrNoRows) {
		renderPublic(w, r, account.NoProfilePage())
		return db.Competitor{}, false
	}
	if err != nil {
		slog.Error("resolve competitor", "user", u.ID, "err", err)
		http.Error(w, "account unavailable", http.StatusInternalServerError)
		return db.Competitor{}, false
	}
	return c, true
}

// AccountDashboard serves GET /account — the competitor landing page (A1):
// up-next entry, recent results, open-challenge banner, quick actions.
func AccountDashboard(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		entries, err := st.ListHandlerEntries(r.Context(), c.ID)
		if err != nil {
			slog.Error("dashboard entries", "competitor", c.ID, "err", err)
			http.Error(w, "account unavailable", http.StatusInternalServerError)
			return
		}
		dogCount, err := st.CountOwnerDogs(r.Context(), c.ID)
		if err != nil {
			slog.Error("dashboard dog count", "competitor", c.ID, "err", err)
			http.Error(w, "account unavailable", http.StatusInternalServerError)
			return
		}
		openCh, err := st.CountOpenChallenges(r.Context(), c.ID)
		if err != nil {
			slog.Error("dashboard challenge count", "competitor", c.ID, "err", err)
			http.Error(w, "account unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, account.DashboardPage(toDashboardVD(r, st, c, entries, int(dogCount), int(openCh))))
	}
}

// AccountProfile serves GET /account/profile — the profile editor (A2),
// prefilled from the competitor row.
func AccountProfile(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		renderPublic(w, r, account.ProfilePage(profileVD(c, false, "")))
	}
}

// AccountProfileSave serves POST /account/profile — validates and saves
// the editable profile fields, then re-renders the form fragment with a
// saved confirmation (or the validation error in place).
func AccountProfileSave(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		displayName := strings.TrimSpace(r.FormValue("display_name"))
		handle := strings.ToLower(strings.TrimSpace(r.FormValue("handle")))
		bio := strings.TrimSpace(r.FormValue("bio"))

		render := func(vd account.ProfileViewData) {
			renderPublic(w, r, account.ProfileForm(vd))
		}

		switch {
		case displayName == "":
			render(profileVDFrom(displayName, handle, bio, false, "Display name is required."))
			return
		case handle == "" || !handlePattern.MatchString(handle):
			render(profileVDFrom(displayName, handle, bio, false, "Handle can use only lowercase letters, digits, and hyphens."))
			return
		}

		available, err := st.HandleAvailable(r.Context(), handle, c.ID)
		if err != nil {
			slog.Error("profile handle check", "err", err)
			render(profileVDFrom(displayName, handle, bio, false, "Something went wrong. Please try again."))
			return
		}
		if !available {
			render(profileVDFrom(displayName, handle, bio, false, "That handle is already taken. Pick another."))
			return
		}

		if err := st.UpdateCompetitorProfile(r.Context(), c.ID, displayName, handle, bio); err != nil {
			if isUniqueViolation(err) {
				render(profileVDFrom(displayName, handle, bio, false, "That handle is already taken. Pick another."))
				return
			}
			slog.Error("update profile", "competitor", c.ID, "err", err)
			render(profileVDFrom(displayName, handle, bio, false, "Something went wrong. Please try again."))
			return
		}
		render(profileVDFrom(displayName, handle, bio, true, ""))
	}
}

// AccountDogs serves GET /account/dogs — the dog roster (A3).
func AccountDogs(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		items, err := st.ListOwnerDogs(r.Context(), c.ID)
		if err != nil {
			slog.Error("dogs list", "competitor", c.ID, "err", err)
			http.Error(w, "account unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, account.DogsListPage(toDogsListVD(items)))
	}
}

// AccountDogsNew serves GET /account/dogs/new — the empty add form (A4).
func AccountDogsNew(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := currentCompetitor(w, r, st); !ok {
			return
		}
		renderPublic(w, r, account.DogsFormPage(account.DogFormViewData{}))
	}
}

// AccountDogsEdit serves GET /account/dogs/{id}/edit — the edit form for a
// dog the competitor owns. 404 when the id misses or belongs to another
// owner.
func AccountDogsEdit(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		dogID, ok := parseEntryID(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		dog, err := st.GetOwnerDog(r.Context(), dogID, c.ID)
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			slog.Error("dog edit load", "dog", dogID, "err", err)
			http.Error(w, "account unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, account.DogsFormPage(dogFormVD(dog)))
	}
}

// AccountDogsCreate serves POST /account/dogs — inserts a new dog under
// the competitor and redirects to the roster. Validation errors re-render
// the form fragment.
func AccountDogsCreate(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		in, vd, ok := parseDogForm(r, account.DogFormViewData{})
		if !ok {
			renderPublic(w, r, account.DogForm(vd))
			return
		}
		if _, err := st.CreateDog(r.Context(), c.ID, in); err != nil {
			slog.Error("create dog", "competitor", c.ID, "err", err)
			vd.Err = "Something went wrong. Please try again."
			renderPublic(w, r, account.DogForm(vd))
			return
		}
		hxRedirect(w, r, "/account/dogs")
	}
}

// AccountDogsUpdate serves POST /account/dogs/{id} — saves edits to a dog
// the competitor owns and redirects to the roster.
func AccountDogsUpdate(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		dogID, ok := parseEntryID(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		// Confirm ownership up front so an edit on someone else's dog 404s
		// rather than silently no-op'ing through the owner-scoped UPDATE.
		if _, err := st.GetOwnerDog(r.Context(), dogID, c.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			slog.Error("dog update load", "dog", dogID, "err", err)
			http.Error(w, "account unavailable", http.StatusInternalServerError)
			return
		}
		base := account.DogFormViewData{IsEdit: true, DogID: dogID, PublicURL: dogPublicURL(dogID)}
		in, vd, ok := parseDogForm(r, base)
		if !ok {
			renderPublic(w, r, account.DogForm(vd))
			return
		}
		if err := st.UpdateDog(r.Context(), dogID, c.ID, in); err != nil {
			slog.Error("update dog", "dog", dogID, "err", err)
			vd.Err = "Something went wrong. Please try again."
			renderPublic(w, r, account.DogForm(vd))
			return
		}
		hxRedirect(w, r, "/account/dogs")
	}
}

// AccountDogsDelete serves POST /account/dogs/{id}/delete — removes a dog
// the competitor owns. The owner-scoped delete makes a guessed id a no-op
// rather than a leak.
func AccountDogsDelete(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		dogID, ok := parseEntryID(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if err := st.DeleteDog(r.Context(), dogID, c.ID); err != nil {
			slog.Error("delete dog", "dog", dogID, "err", err)
			http.Error(w, "account unavailable", http.StatusInternalServerError)
			return
		}
		hxRedirect(w, r, "/account/dogs")
	}
}
