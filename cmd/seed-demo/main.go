// seed-demo builds a complete, realistic demo dataset that exercises every
// surface of the app: signup/account, the competitor directory, the public
// event + results pages, the judge scoring flow, the registration bridge,
// and the full admin surface (events, trials, registrations, judge
// assignments, challenges, users).
//
// It WIPES all application data first, then reseeds deterministically, so it
// is the one-command way to reset the demo to a known state.
//
//	mage seeddemo            # uses DB_PATH or ./data/app.db
//
// Logins (all demo passwords are "demo1234" except the admin):
//
//	admin@example.com   / admin1234     (admin)
//	judge@example.com   / demo1234      (judge)
//	jpereira@example.com/ demo1234      (judge)
//	ltanaka@example.com / demo1234      (competitor, @ltanaka, 3 dogs)
//	rokafor@example.com / demo1234      (competitor, @rokafor, 2 dogs)
//	khessel@example.com / demo1234      (competitor, @khessel, 1 dog)
//	syi@example.com     / demo1234      (competitor, @syi, 1 dog)
//	dfowler@example.com / demo1234      (competitor, @dfowler, no dogs)
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/scoring"
	"github.com/flintcraftstudio/k9-trials/internal/scoring/templates"
	"github.com/flintcraftstudio/k9-trials/internal/store"
)

const demoPass = "demo1234"

// --- World definition -------------------------------------------------------

type dogSpec struct {
	call, regName, breed, dob, regno string
}

type compSpec struct {
	email, name, handle, bio string
	dogs                     []dogSpec
}

// entrySpec is an accepted registration that produced an entry. target is
// the finalized point total (drives Q/NQ); chal, when set, files a challenge
// on the entry by its competitor.
type entrySpec struct {
	handle, dog string
	status      string // registered | scoring | finalized
	target      int    // finalized score total; ignored otherwise
	chal        string // "" | open | under_review | resolved
	chalReason  string
	chalNotes   string
}

// regSpec is a registration with no entry yet (the D5 review queue).
type regSpec struct {
	handle, dog, status string // status: pending | waitlisted
}

type trialSpec struct {
	disc    string
	level   int64
	date    string
	status  string // pending | in_progress | complete
	judge   string // judge email assigned to this trial entries, or ""
	entries []entrySpec
	pending []regSpec
}

type eventSpec struct {
	slug, name, loc, start, end, status string
	trials                              []trialSpec
}

var competitors = []compSpec{
	{
		email: "ltanaka@example.com", name: "L. Tanaka", handle: "ltanaka",
		bio: "Working IPO since 2014. Mali and Czech-line GSDs. Happy to talk training.",
		dogs: []dogSpec{
			{"Vex", "Forge vom Schwarzwald", "Czech German Shepherd", "2021-04-12", "K9-3187"},
			{"Kestrel", "Kestrel of Tanaka", "Belgian Malinois", "2019-06-01", "K9-2741"},
			{"Nyx", "", "Belgian Malinois", "2023-02-20", ""},
		},
	},
	{
		email: "rokafor@example.com", name: "R. Okafor", handle: "rokafor",
		bio: "Detection and obedience. Two dogs in active competition.",
		dogs: []dogSpec{
			{"Atlas", "Atlas vom Okafor", "Belgian Malinois", "2020-09-15", "K9-3041"},
			{"Echo", "Echo of the High Line", "German Shepherd", "2018-11-03", "K9-2419"},
		},
	},
	{
		email: "khessel@example.com", name: "K. Hessel", handle: "khessel",
		bio: "Dutch Shepherds. Tracking specialist.",
		dogs: []dogSpec{
			{"Saber", "Saber van Hessel", "Dutch Shepherd", "2020-03-30", "K9-3902"},
		},
	},
	{
		email: "syi@example.com", name: "S. Yi", handle: "syi",
		bio: "First season competing with Lumen.",
		dogs: []dogSpec{
			{"Lumen", "", "Belgian Malinois", "2021-08-08", "K9-2884"},
		},
	},
	{
		email: "dfowler@example.com", name: "D. Fowler", handle: "dfowler",
		bio: "Just signed up. Looking for my first trial.",
	},
}

var world = []eventSpec{
	{
		slug: "cedar-creek-spring-2026", name: "Cedar Creek Spring Trial",
		loc: "Cedar Creek Field, MT", start: "2026-05-15", end: "2026-05-16", status: "published",
		trials: []trialSpec{
			{
				disc: "OB", level: 1, date: "2026-05-15", status: "in_progress", judge: "judge@example.com",
				entries: []entrySpec{
					{handle: "rokafor", dog: "Echo", status: "finalized", target: 104},
					{handle: "khessel", dog: "Saber", status: "finalized", target: 98},
					{handle: "rokafor", dog: "Atlas", status: "finalized", target: 74,
						chal: "open", chalReason: "Recall happened before the line — should be a deduction, not NQ. Asking for re-score under 4.3.2."},
					{handle: "ltanaka", dog: "Kestrel", status: "finalized", target: 95,
						chal: "under_review", chalReason: "Heel-free score looks low; dog was attentive the whole pattern."},
					{handle: "ltanaka", dog: "Vex", status: "scoring"},
					{handle: "syi", dog: "Lumen", status: "registered"},
				},
			},
			{
				disc: "OB", level: 2, date: "2026-05-16", status: "pending",
				pending: []regSpec{
					{handle: "ltanaka", dog: "Vex", status: "pending"},
					{handle: "rokafor", dog: "Atlas", status: "pending"},
					{handle: "syi", dog: "Lumen", status: "waitlisted"},
				},
			},
			{
				disc: "TR", level: 1, date: "2026-05-16", status: "pending",
				pending: []regSpec{
					{handle: "ltanaka", dog: "Nyx", status: "pending"},
				},
			},
		},
	},
	{
		slug: "hopkins-mill-tracking", name: "Hopkins Mill Tracking",
		loc: "Hopkins Mill, OR", start: "2026-06-20", end: "2026-06-21", status: "published",
		trials: []trialSpec{
			{
				disc: "TR", level: 1, date: "2026-06-20", status: "pending",
				entries: []entrySpec{
					{handle: "ltanaka", dog: "Kestrel", status: "registered"},
					{handle: "khessel", dog: "Saber", status: "registered"},
				},
				pending: []regSpec{
					{handle: "syi", dog: "Lumen", status: "pending"},
				},
			},
		},
	},
	{
		slug: "brindle-bay-2025", name: "Brindle Bay Autumn",
		loc: "Brindle Bay, CA", start: "2025-10-04", end: "2025-10-05", status: "closed",
		trials: []trialSpec{
			{
				disc: "OB", level: 1, date: "2025-10-04", status: "complete", judge: "jpereira@example.com",
				entries: []entrySpec{
					{handle: "ltanaka", dog: "Vex", status: "finalized", target: 110},
					{handle: "rokafor", dog: "Atlas", status: "finalized", target: 76},
					{handle: "khessel", dog: "Saber", status: "finalized", target: 92,
						chal: "resolved", chalReason: "Long-down break was the dog next to us, not Saber.",
						chalNotes: "Reviewed the ring video with the judge. Re-scored; the break stands. Dispute closed."},
				},
			},
		},
	},
	{
		slug: "cedar-creek-summer", name: "Cedar Creek Summer",
		loc: "Cedar Creek Field, MT", start: "2026-07-11", end: "2026-07-13", status: "draft",
		trials: []trialSpec{
			{disc: "OB", level: 1, date: "2026-07-11", status: "pending"},
		},
	},
}

// --- Seeding ----------------------------------------------------------------

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/app.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return fmt.Errorf("ensure db dir: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()
	if _, err := conn.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		return fmt.Errorf("pragma: %w", err)
	}
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.Up(conn, "migrations"); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	if err := wipe(conn); err != nil {
		return fmt.Errorf("wipe: %w", err)
	}

	st := store.New(conn)
	ctx := context.Background()
	s := &seeder{
		st:    st,
		conn:  conn,
		ctx:   ctx,
		tpl:   templates.L1OB(),
		comp:  map[string]int64{},
		user:  map[string]int64{},
		dog:   map[string]int64{},
		entry: map[string]createdEntry{},
	}

	if err := s.seedUsersAndCompetitors(); err != nil {
		return err
	}
	if err := s.seedWorld(); err != nil {
		return err
	}

	fmt.Println(s.summary())
	return nil
}

// wipe clears every application table so the seed is a clean reset. Foreign
// keys are disabled for the duration so delete order does not matter, and
// the autoincrement counters are reset for stable ids.
func wipe(conn *sql.DB) error {
	tables := []string{
		"modifier_applications", "auto_trigger_firings", "penalty_occurrences",
		"criterion_scores", "challenges", "registrations", "entries",
		"dogs", "competitors", "trials", "events", "sessions", "users",
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=OFF;"); err != nil {
		return err
	}
	for _, t := range tables {
		if _, err := conn.Exec("DELETE FROM " + t); err != nil {
			return fmt.Errorf("delete %s: %w", t, err)
		}
	}
	// Reset autoincrement so ids start at 1 on every reseed.
	_, _ = conn.Exec("DELETE FROM sqlite_sequence")
	if _, err := conn.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		return err
	}
	return nil
}

type createdEntry struct {
	id     int64
	number int64
	compID int64
}

type seeder struct {
	st   *store.Store
	conn *sql.DB
	ctx  context.Context
	tpl  scoring.ScoresheetTemplate

	adminID  int64
	user     map[string]int64 // email -> user id
	comp     map[string]int64 // handle -> competitor id
	dog      map[string]int64 // handle/call -> dog id
	entry    map[string]createdEntry
	nEntries int
	nRegs    int
	nChals   int
}

func (s *seeder) seedUsersAndCompetitors() error {
	var err error
	if s.adminID, err = s.st.CreateUser(s.ctx, "admin@example.com", "admin1234", "admin"); err != nil {
		return fmt.Errorf("admin: %w", err)
	}
	for _, email := range []string{"judge@example.com", "jpereira@example.com"} {
		id, err := s.st.CreateUser(s.ctx, email, demoPass, "judge")
		if err != nil {
			return fmt.Errorf("judge %s: %w", email, err)
		}
		s.user[email] = id
	}

	for _, c := range competitors {
		uid, err := s.st.CreateUser(s.ctx, c.email, demoPass, "competitor")
		if err != nil {
			return fmt.Errorf("user %s: %w", c.email, err)
		}
		s.user[c.email] = uid
		comp, err := s.st.Q().CreateCompetitor(s.ctx, db.CreateCompetitorParams{
			UserID:      sql.NullInt64{Int64: uid, Valid: true},
			Handle:      c.handle,
			DisplayName: c.name,
			Bio:         c.bio,
		})
		if err != nil {
			return fmt.Errorf("competitor %s: %w", c.handle, err)
		}
		s.comp[c.handle] = comp.ID
		for _, d := range c.dogs {
			dog, err := s.st.Q().CreateDog(s.ctx, db.CreateDogParams{
				OwnerID:            comp.ID,
				CallName:           d.call,
				RegisteredName:     d.regName,
				Breed:              d.breed,
				DateOfBirth:        parseNullDate(d.dob),
				RegistrationNumber: d.regno,
			})
			if err != nil {
				return fmt.Errorf("dog %s: %w", d.call, err)
			}
			s.dog[c.handle+"/"+d.call] = dog.ID
		}
	}
	return nil
}

func (s *seeder) seedWorld() error {
	for _, ev := range world {
		event, err := s.st.Q().CreateEvent(s.ctx, db.CreateEventParams{
			Slug:      ev.slug,
			Name:      ev.name,
			Location:  ev.loc,
			StartDate: parseDate(ev.start),
			EndDate:   parseDate(ev.end),
			Status:    ev.status,
			CreatedBy: s.adminID,
		})
		if err != nil {
			return fmt.Errorf("event %s: %w", ev.slug, err)
		}
		for _, tr := range ev.trials {
			if err := s.seedTrial(ev, event.ID, tr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *seeder) seedTrial(ev eventSpec, eventID int64, tr trialSpec) error {
	trial, err := s.st.Q().CreateTrial(s.ctx, db.CreateTrialParams{
		EventID:         eventID,
		Discipline:      tr.disc,
		Level:           tr.level,
		TrialDate:       parseDate(tr.date),
		TemplateVersion: "2026.1",
		Status:          tr.status,
	})
	if err != nil {
		return fmt.Errorf("trial %s/%s%d: %w", ev.slug, tr.disc, tr.level, err)
	}

	var judgeID sql.NullInt64
	if tr.judge != "" {
		judgeID = sql.NullInt64{Int64: s.user[tr.judge], Valid: true}
	}

	number := int64(0)
	for _, es := range tr.entries {
		number++
		if err := s.seedEntry(ev, trial.ID, number, judgeID, es); err != nil {
			return err
		}
	}
	for _, rg := range tr.pending {
		if err := s.seedPendingReg(trial.ID, rg); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedEntry(ev eventSpec, trialID, number int64, judgeID sql.NullInt64, es entrySpec) error {
	compID := s.comp[es.handle]
	dogID := s.dog[es.handle+"/"+es.dog]
	comp := competitorByHandle(es.handle)

	entry, err := s.st.Q().CreateEntryForRegistration(s.ctx, db.CreateEntryForRegistrationParams{
		TrialID:     trialID,
		JudgeID:     judgeID,
		EntryNumber: number,
		HandlerName: comp.name,
		DogName:     es.dog,
		DogBreed:    dogBreed(es.handle, es.dog),
		DogID:       sql.NullInt64{Int64: dogID, Valid: true},
		HandlerID:   sql.NullInt64{Int64: compID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("entry %s/%s: %w", es.handle, es.dog, err)
	}
	s.nEntries++
	s.entry[ev.slug+"/"+es.dog] = createdEntry{id: entry.ID, number: number, compID: compID}

	if es.status != "registered" {
		if _, err := s.st.Q().UpdateEntryStatus(s.ctx, db.UpdateEntryStatusParams{Status: es.status, ID: entry.ID}); err != nil {
			return fmt.Errorf("entry status %d: %w", entry.ID, err)
		}
	}

	// Every entry produced by an accepted registration. The registration is
	// dated a few days before the trial.
	if err := s.insertRegistration(trialID, compID, dogID, s.user[comp.email], "accepted", entry.ID); err != nil {
		return err
	}

	if es.status == "finalized" {
		if err := s.scoreEntry(entry.ID, judgeID.Int64, es.target); err != nil {
			return fmt.Errorf("score %d: %w", entry.ID, err)
		}
	}

	if es.chal != "" {
		if err := s.insertChallenge(entry.ID, compID, es.chal, es.chalReason, es.chalNotes); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedPendingReg(trialID int64, rg regSpec) error {
	compID := s.comp[rg.handle]
	dogID := s.dog[rg.handle+"/"+rg.dog]
	email := competitorByHandle(rg.handle).email
	return s.insertRegistration(trialID, compID, dogID, s.user[email], rg.status, 0)
}

// insertRegistration writes a registration row directly so the demo can set
// statuses (accepted/pending/waitlisted) and link an entry where one exists.
func (s *seeder) insertRegistration(trialID, compID, dogID, submittedBy int64, status string, entryID int64) error {
	var entry sql.NullInt64
	var reviewedBy sql.NullInt64
	var reviewedAt sql.NullTime
	if status != "pending" {
		reviewedBy = sql.NullInt64{Int64: s.adminID, Valid: true}
		reviewedAt = sql.NullTime{Time: time.Now().Add(-48 * time.Hour), Valid: true}
	}
	if entryID != 0 {
		entry = sql.NullInt64{Int64: entryID, Valid: true}
	}
	_, err := s.conn.ExecContext(s.ctx,
		`INSERT INTO registrations (trial_id, competitor_id, dog_id, status, submitted_by, submitted_at, reviewed_by, reviewed_at, entry_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		trialID, compID, dogID, status, submittedBy, time.Now().Add(-72*time.Hour), reviewedBy, reviewedAt, entry,
	)
	if err != nil {
		return fmt.Errorf("registration: %w", err)
	}
	s.nRegs++
	return nil
}

// insertChallenge writes a challenge row directly so the demo can set its
// status and resolution.
func (s *seeder) insertChallenge(entryID, compID int64, status, reason, notes string) error {
	var resolvedBy sql.NullInt64
	var resolvedAt sql.NullTime
	if status == "resolved" || status == "dismissed" {
		resolvedBy = sql.NullInt64{Int64: s.adminID, Valid: true}
		resolvedAt = sql.NullTime{Time: time.Now().Add(-24 * time.Hour), Valid: true}
	}
	_, err := s.conn.ExecContext(s.ctx,
		`INSERT INTO challenges (entry_id, filed_by, reason, status, resolution_notes, resolved_by, resolved_at, filed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		entryID, compID, reason, status, notes, resolvedBy, resolvedAt, time.Now().Add(-120*time.Hour),
	)
	if err != nil {
		return fmt.Errorf("challenge: %w", err)
	}
	s.nChals++
	return nil
}

// scoreEntry records criterion scores scaled to roughly target/maxTotal of
// full marks, applying the SAME fraction to every criterion. A uniform
// fraction keeps every exercise at the same percentage, so the qualifying
// decision turns on the point total rather than any single exercise dipping
// into the Insufficient tier: a high fraction qualifies (every exercise
// passes), a fraction below the 70% threshold does not. round records the
// half-up share for each criterion.
func (s *seeder) scoreEntry(entryID, judgeID int64, target int) error {
	maxTotal := 0
	for _, ph := range s.tpl.Phases {
		for _, ex := range ph.Exercises {
			for _, c := range ex.Criteria {
				maxTotal += int(c.MaxPoints)
			}
		}
	}
	if maxTotal == 0 {
		return nil
	}
	if target > maxTotal {
		target = maxTotal
	}

	for _, ph := range s.tpl.Phases {
		for _, ex := range ph.Exercises {
			for _, c := range ex.Criteria {
				max := int(c.MaxPoints)
				// Round-half-up share of the target for this criterion.
				p := (max*target*2 + maxTotal) / (maxTotal * 2)
				if p > max {
					p = max
				}
				if p < 1 {
					p = 1
				}
				if _, err := s.st.Q().RecordCriterionScore(s.ctx, db.RecordCriterionScoreParams{
					EntryID:       entryID,
					ExerciseCode:  ex.Code,
					CriterionCode: c.Code,
					Points:        int64(p),
					JudgedBy:      judgeID,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *seeder) summary() string {
	return fmt.Sprintf(
		"Demo seeded (data reset).\n"+
			"  %d users · %d competitors · %d events · %d entries · %d registrations · %d challenges\n"+
			"  admin@example.com / admin1234   (admin)\n"+
			"  judge@example.com / demo1234    (judge)\n"+
			"  ltanaka@example.com / demo1234  (competitor @ltanaka)\n"+
			"  Open /admin, /events, /competitors, or log in as a competitor.",
		3+len(competitors), len(competitors), len(world), s.nEntries, s.nRegs, s.nChals,
	)
}

// --- helpers ----------------------------------------------------------------

func competitorByHandle(handle string) compSpec {
	for _, c := range competitors {
		if c.handle == handle {
			return c
		}
	}
	return compSpec{}
}

func dogBreed(handle, call string) string {
	for _, c := range competitors {
		if c.handle != handle {
			continue
		}
		for _, d := range c.dogs {
			if d.call == call {
				return d.breed
			}
		}
	}
	return ""
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func parseNullDate(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: parseDate(s), Valid: true}
}
