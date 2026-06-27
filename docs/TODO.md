# TODO — deferred wiring

Running list of UI surfaces that are built but intentionally not wired to a
backend yet. Each entry lists the exact touchpoints to finish later.

## Profile photo upload (handlers + dogs)

The management UI is live: handlers can manage a photo for themselves (A2
profile editor) and for each dog (A4 dog form). Selecting a file shows a
live in-page preview; a Remove button clears it. **Nothing is uploaded,
stored, or recalled** — the file is discarded on submit and `PhotoURL` is
always empty, so the initials placeholder always renders.

### What exists

- **Component** — `internal/view/components/avatar_upload.templ` +
  `AvatarUploadProps` / helpers in `internal/view/components/components.go`.
  Renders current photo (or initials placeholder), an "Upload photo" button
  with live preview, and a "Remove" toggle.
- **Behaviour** — `web/static/js/avatar.js` (plain DOM, no Alpine — survives
  htmx form swaps). Loaded once in `internal/view/layout.templ`.
- **Styling** — `.avatar-edit` / `.avatar-img` in `tailwind/input.css`
  (reuses the design system's `.avatar xl circ|sq` classes).

### Form contract (already posted, nothing consumes it)

Both forms are `enctype="multipart/form-data"` and post:

- `photo` — the chosen file (multipart file part).
- `photo_remove` — hidden field, `"true"` when the user cleared the photo,
  else `"false"`. Lets the backend distinguish "no change" from "remove".

### Touchpoints to finish the wiring

1. **Storage layer** — decide where bytes live (local dir under a served
   path, or object storage). Add a column to persist the reference:
   - `migrations/` — add `photo_url` (or `photo_key`) to `competitors`
     (`009_create_competitors.sql`) and `dogs` (`010_create_dogs.sql`).
   - `queries/` — extend the competitor + dog update queries to set it;
     `mage generate` to regenerate `internal/db/`. (Keep query files ASCII —
     non-ASCII breaks `sqlc generate`.)
   - `internal/store/account.go` — extend `UpdateCompetitorProfile`,
     `CreateDog`, `UpdateDog` (and `store.DogInput`) to carry the photo ref.

2. **Handlers** — `internal/handler/account.go`:
   - `AccountProfileSave` — after `r.ParseForm`, read the file via
     `r.FormFile("photo")` and the `photo_remove` flag; validate type/size
     (server-side — the `accept` attr is advisory only); persist or clear.
   - `AccountDogsCreate` / `AccountDogsUpdate` (via `parseDogForm` in
     `account_mapper.go`) — same: read `photo` + `photo_remove`, validate,
     persist or clear.
   - Note: `parseDogForm`/`AccountProfileSave` currently call `r.ParseForm()`
     then `r.FormValue(...)`. `FormValue` auto-parses multipart for text
     fields, but for the file part call `r.ParseMultipartForm(maxBytes)` and
     `r.FormFile("photo")` explicitly, with a real size cap.

3. **Mappers** — `internal/handler/account_mapper.go`: set `PhotoURL` from
   the stored reference in `profileVD`, `profileVDFrom`, and `dogFormVD`
   (currently left empty). Once populated, the avatar control shows the real
   photo instead of initials automatically.

4. **Public display** — surface the photo on the public pages too:
   - `internal/view/competitors/profile.templ` (P6) — `<div class="avatar-lg">`
     currently renders `d.Initials`; swap to an `<img>` when a photo exists.
   - `internal/view/dogs/profile.templ` (P7) — add an avatar where relevant.
   - Add the photo field to those view-data structs + their mappers.

### Validation reminders

- Enforce content-type and max size **server-side**; never trust the client
  `accept` attribute or the extension.
- Strip/ignore EXIF and re-encode if you serve user images directly.
- Scope reads/writes to the owning competitor (dogs are owner-scoped via
  `GetOwnerDog`) so one user can't overwrite another's photo.
