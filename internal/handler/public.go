package handler

// Public read-only handlers are split by surface:
//
//   - public_events.go   — events index (P1), event detail (P2), trial
//     leaderboard (P3), public entry detail (P4)
//   - public_profiles.go — competitor directory (P5), competitor profile
//     (P6), dog profile (P7)
//   - public_mapper.go   — shared presentation helpers (discipline labels,
//     date formatting, registration state)
//
// This file is intentionally a package marker; add new public surfaces in
// the file matching their area.
