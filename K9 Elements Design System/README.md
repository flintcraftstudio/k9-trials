# K9 Elements Design System

K9 Elements supports engagement and training in working dog sports by organizing **judged trials** and supplying **high-quality training materials**. The mission is to grow the sport through streamlined trial participation and guided development paths for the four primary events:

- **Obedience** — precision heelwork, recalls, retrieves, control
- **Protection** — bitework, courage tests, handler defense
- **Tracking** — scent-trailing and article indication on natural terrain
- **Detection** — odor recognition and source identification in target areas

This document indexes the brand's visual + content foundations and points to the working files in this folder.

---

## Audience & product surfaces

| Surface | Role | Tone |
|---|---|---|
| **Marketing site** (public) | Image- and video-first. Sells the sport and the trials. Minimal copy on the front. | Cinematic, confident, restrained. |
| **Training & trials app** (paid, gated) | Where the long-form text lives — guided development paths, trial entry/scheduling, judging tools, event-specific lesson tracks. | Practical, instructional, organized by event. |

This kit prioritises the **marketing site**; app conventions are sketched but not exhaustive.

---

## Source materials

| Source | Where | Status |
|---|---|---|
| Logo | `uploads/k9-elements-logo.png` → `assets/k9-elements-logo.png` | Imported. 600×600 PNG, three-stripe wordmark on white. |
| Mounted codebase | `oatmeal-mist-instrument/` (read-only, local mount) | Read. Tailwind Plus "Oatmeal" template (Instrument Serif display + `mist-*` neutral scale). Visual chassis adapted for K9 Elements — body face swapped to Instrument Sans to keep the type pairing inside one foundry. |
| Brand brief | Direct from team | Logo features three colored stripes (orange, blue, green); paired with grey/black these theme the four primary events. Site should be **clean** and let imagery do the work — long-form copy lives behind the paywall. |

> **Note on the chassis.** The Oatmeal template is a Tailwind Plus license. Component implementations in `ui_kits/marketing-site/` are **fresh, cosmetic recreations** of the template's *visual language* — not copies of the licensed source code. The neutral scale and type pairing are reused because they map well to the K9 brief (premium-but-restrained, photography-forward, dark-mode-ready).

---

## The event theming system

The single most important decision in this kit: **users should always know which event they're focused on by subtle themed highlighting** — never by loud color blocking.

| Event | Hue | Hex (600) | Use |
|---|---|---|---|
| Obedience | Blue | `#066d9e` | Logo middle stripe |
| Protection | Charcoal/Black | `#1a1a1a` | The K-letter colour |
| Tracking | Green | `#3aa548` | Logo right stripe |
| Detection | Orange | `#bf692e` | Logo left stripe |

> **⚠ Assumed mapping.** The brief defined the four hues and the four events but did **not** explicitly pair them. The mapping above is my proposal (Obedience→blue for precision; Protection→charcoal for seriousness and to match the wordmark; Tracking→green for outdoors; Detection→orange for high-visibility alert). **Please confirm or remap before locking the system.**

**Where event color shows up (subtle, always):**
- A 3px left rail on event-scoped cards / nav rows
- The eyebrow label above an event headline (e.g. `OBEDIENCE`)
- A tinted filter on imagery within an event section (`mix-blend-overlay` at low opacity)
- The progress ring on a development-path module
- Tab indicators in the app
- **Never** as a full-bleed background, never as button fill, never as body text color.

---

## File index

| File | What's in it |
|---|---|
| `README.md` | This file — manifest, tone & content rules, visual foundations, iconography. |
| `colors_and_type.css` | All CSS custom properties: brand + neutral + event scales, typography tokens, radii, shadows, spacing. Import this in any artifact. |
| `SKILL.md` | Agent-Skill entry point. Read first when this folder is loaded as a skill. |
| `assets/` | Logos and brand imagery. `k9-elements-logo.png` is the master; SVG wordmark TBD. |
| `fonts/` | Empty — fonts are loaded from Google Fonts CDN (see CONTENT FUNDAMENTALS below). |
| `preview/` | Design-system specimen cards (typography, colour swatches, components, spacing). Surfaced in the Design System tab. |
| `ui_kits/marketing-site/` | Marketing-site UI kit: navbar, hero, event-grid feature blocks, trial card, footer, etc. Interactive `index.html` demo. |

---

## CONTENT FUNDAMENTALS

The marketing front is deliberately copy-light. The paid back-end (training paths, trial briefs, judging guides) is where prose lives. Tone shifts accordingly.

### Voice
- **On the public site** — cinematic, confident, quiet. Headlines are short fragments. Sub-copy is one or two sentences max. We let video and photography carry weight.
- **In the training app** — practical, instructional, dog-handler peer-to-peer. We assume the reader is a serious amateur or a working professional. Never condescending.
- **In trial communications** — procedural and exact. Judging language. No marketing fluff in a trial brief.

### Pronouns & address
- **You** for the reader / handler. Direct, no warm-up.
- **We** for K9 Elements as an organization, sparingly. Most copy doesn't need a first person.
- Never "us" as a substitute for "you and your dog." The handler-dog team owns the work; we just organize the trial.

### Casing
- **Sentence case** for almost everything — headlines, buttons, nav. Long marketing headlines read like a sentence: *"Built for the four events that matter."*
- **ALL CAPS** in two places only: eyebrows above section headlines (`TRACKING`, `THIS WEEKEND`) and trial-status pills (`OPEN`, `WAITLIST`).
- **Title Case** for event names when standing alone in a list ("Obedience, Protection, Tracking, Detection") — these are proper nouns of the sport.

### Punctuation & rhythm
- Em dashes used freely as a rhythmic device — they suit the cinematic register.
- Oxford commas always.
- Numbers as digits past nine; written out below ("four events," "12 trials this season").
- Times use a dot separator on the marketing site (`08.30`); the app uses colon (`08:30`).

### Things we don't do
- No exclamation points. Anywhere. The work is the energy.
- No emoji in product or marketing. Emoji feels off-register for the working-dog world.
- No "Discover," "Unlock," "Elevate," "Empower," or other SaaS verbs. Use real verbs: *Enter a trial. Score a session. Track a path.*
- No question-mark headlines. Statements only.
- No "AI" framing — even where ML is involved, talk about the result, not the method.

### Example copy (do / don't)

| ✓ Use | ✗ Avoid |
|---|---|
| "Built for the four events." | "Discover your training journey!" |
| "Tracking opens at 06.00." | "Don't miss out — tracking starts soon 🐕" |
| "Enter the spring trial. 24 spots, 14 left." | "Unlock exclusive early-bird pricing today!" |
| "A guided path through 12 weeks of detection work." | "Empower your dog's full potential." |

---

## VISUAL FOUNDATIONS

### Type
- **Display** — *Instrument Serif* (Google Fonts). Used for h1, h2, hero headlines, large pull quotes. Italic supported and used sparingly for tonal lift. The serif is the brand's quiet authority — it pairs the seriousness of the working-dog world with a cinematic, almost editorial register.
- **Body / UI** — *Instrument Sans* (Google Fonts, sister face to Instrument Serif from the same foundry). Body copy, nav, buttons, labels, tables. Weights 400/500/600. Pairing with the display serif from a single foundry keeps type texture consistent.
- No third family. No monospace in marketing surfaces (the app uses Instrument Sans at smaller sizes for tabular data; numeric tabular figures `font-variant-numeric: tabular-nums` are enabled for trial schedules and scoring).

**Hierarchy (display sizes are responsive and run large):**
- Hero h1: `text-5xl/12 → sm:text-[5rem]/20`, Instrument Serif, tracking-tight, balanced wrap.
- Section h2: `text-[2rem]/10 → sm:text-5xl/14`, Instrument Serif.
- Body lg: `text-lg/8`, Instrument Sans, `neutral-700` (mist-700).
- Body md: `text-base/7`, Instrument Sans.
- Eyebrow: `text-sm/7 font-semibold uppercase tracking-wide` — this is where the event color lives.
- Caption / fine print: `text-sm/7`, mist-600.

### Colour
- **Neutral spine** — the `mist-*` scale from the chassis, a cool desaturated near-grey. Pages live on `mist-100` light / `mist-950` dark. We never use pure white or pure black for surfaces; mist tones read more cinematic against photography.
- **Event hues** — the four event colors (orange `#bf692e`, blue `#066d9e`, green `#3aa548`, charcoal `#1a1a1a`) each have an 11-step scale (50→950) in `colors_and_type.css`. Use the 600 step as the canonical accent.
- **Semantic** — success, warning, danger. We co-opt the event hues conservatively (green=success, orange=warning) so the palette stays tight; danger uses a dedicated red (`#b3261e`).

### Imagery
- **Photography and video are the brand.** Pages are designed to frame imagery, not compete with it.
- Vibe: warm-cool tonal range, natural light, field-realistic. No studio-glamour, no oversaturation. Dogs working, not posing.
- **Subtle event tint:** where a photo sits inside an event section, we use a `mix-blend-overlay` at ~12–18% opacity with that event's 600-step color. The effect is barely perceptible but unifies the section.
- Aspect ratios: 16:9 for hero video; 4:5 for portrait dog cards; 3:2 for trial location photos. Avoid square unless asset is square-native.
- Crops favour mid-body and head-on-action. Avoid full-body silhouettes against busy backgrounds.

### Backgrounds & textures
- Default surface: solid `mist-100` (light) or `mist-950` (dark).
- **Wallpaper blocks** (adapted from the Oatmeal chassis) — soft linear gradients with a 4× SVG turbulence noise overlay at ~30% opacity, mix-blend-overlay. Used as backdrops behind hero screenshots and event tiles. Four wallpaper colorways match the four events. Documented in `preview/wallpapers.html`.
- No hand-drawn illustration. No gradient meshes. No pattern fills beyond the noise overlay.

### Layout
- Page max-width: `7xl` (80rem / 1280px) for content, with `px-6 md:px-10` gutters.
- Vertical rhythm: sections are `py-16` (64px). Hero/CTA sections may go `py-24`.
- 12-column grid implied; we usually compose with flex + gap, not explicit grid lines.
- Sticky navbar (`5.25rem` tall) with `bg-mist-100/95 backdrop-blur`.

### Borders, radii, shadows
- Radii are restrained: `--radius-sm: 4px`, `--radius-md: 6px` for buttons/inputs, `--radius-lg: 12px` for cards, `--radius-xl: 24px` for hero screenshot frames, `--radius-full: 9999px` for buttons and pills. No 32px+ "blob" rounding.
- Borders are 1px, `mist-950/10` light, `white/10` dark. Cards on light mode use `bg-mist-950/2.5` (a ~2.5% tinted surface) and no border; dark mode uses `bg-white/5` with `outline-1 outline-white/10` for definition.
- **Shadow system is minimal.** We don't lean on drop shadows. Where elevation is needed (modals, dropdowns), use `--shadow-sm` for resting and `--shadow-md` for popovers. No coloured shadows, no neon glow.

### Buttons & interactive
- Primary button — pill (`rounded-full`), `bg-mist-950` text-white, sizes `md` (px-3 py-1) and `lg` (px-4 py-2). Hover: lifts to `mist-800`. No transform.
- Soft button — `bg-mist-950/10`, used for secondary CTAs. Hover: `bg-mist-950/15`.
- Plain button — text-only, hover `bg-mist-950/10`. Used for tertiary actions.
- **Event-themed buttons exist but are rare** — `bg-event-600` is used at most once per page (the primary "Enter trial" CTA on a trial detail page); everywhere else, the mist-950 button is themed by context, not color.
- All buttons have `text-sm/7 font-medium`. No uppercase button labels.

### States
- **Hover (desktop)** — backgrounds shift one step darker on dark surfaces, one step lighter on light surfaces. Text-only buttons gain a 10% tinted background. Imagery scales `1.02` over `300ms ease-out` when in a clickable card.
- **Press** — no shrink transform. We rely on the colour shift only.
- **Focus** — `outline-2 outline-offset-2 outline-mist-950` (or current event color in scoped contexts). Keyboard-visible only via `:focus-visible`.
- **Disabled** — `opacity-50`, `pointer-events-none`.

### Motion
- Easing: default `cubic-bezier(0.32, 0.72, 0, 1)` (ease-out-quart-ish). For UI movement, 200–300ms is the sweet spot. Hero video crossfades at 600ms.
- No bounces. No springs. No long page-load animations.
- Page-entry: a 24px upward translate + opacity fade on the hero unit only, staggered children at 60ms. Other sections fade in on scroll at 150ms — once.
- Hover micro-motion: subtle (chevron nudge 3px right; image scale 1.02). Never wholesale element jumps.

### Transparency & blur
- The navbar uses `backdrop-blur-md` over a 95% mist-100 base. We use blur sparingly — it's noise on top of imagery — and never as a stylistic flourish over solid backgrounds.
- Card surfaces use *tinted opacity* (`mist-950/2.5`) rather than `rgba(0,0,0,0.025)` directly — the difference is academic but it keeps cards in the colour system.

### Iconography
See the ICONOGRAPHY section below.

---

## ICONOGRAPHY

The Oatmeal chassis ships a custom **stroke icon set** — ~120 small SVG glyphs, all 13–24px native, `strokeWidth=1`, hairline weight, `round` joins. Tone matches Instrument Serif: precise, light, editorial.

**What we use:**
- `assets/icons/` — a hand-picked subset of the chassis icons (arrows, chevrons, checkmark, plus/minus, calendar, clock, map-pin, target, paperclip, magnifying-glass, user-circle, plus the four social glyphs). Total ~30 SVGs. Each is a tiny self-contained file; we inline them at use or `<img src>` them, sized via CSS.
- No icon font. No SVG sprite sheet (the icons are small and rarely repeat 10+ times on a page).
- Strokes are `currentColor`, so an icon inside event-colored text picks up the event hue automatically.

**Substitutions / extensions:**
- For event-specific icons that the chassis lacks — a paw, a leash, a tracking-flag, a bite-sleeve — we will add custom 24×24 stroke-1 SVGs in the chassis style. **None drawn yet.** This is a known gap; flagged as a follow-up.
- Where neither the chassis nor a planned custom icon fits, fall back to **Lucide** (`https://unpkg.com/lucide-static@latest/icons/<name>.svg`) and re-stroke to weight 1 to match. Document any Lucide substitution in the file that uses it.

**Emoji:** never used. (See content rules.)
**Unicode glyphs:** the curly quotes `“ ”` in pull-quote pseudo-elements, the en-dash `–` for ranges, the em-dash `—` for prose. Nothing else.

---

## What's still open

These are the things the team needs to weigh in on before this system is locked:

1. **Event → color mapping** (above). My defaults are a guess.
2. **Custom event icons** (paw / track / sleeve / cone) — not yet drawn.
3. **Logo SVG** — we only have the PNG. A vector wordmark + monogram lockup would let us put the logo at any size without ringing artifacts. Please drop a SVG when you have one.
4. **Photography library** — placeholder gray blocks are used in `ui_kits/marketing-site/` until real video stills land.
5. **Tracking and Protection sub-brand voice** — Detection and Obedience read fairly clearly; Tracking ("quiet, patient, weatherproof") and Protection ("controlled aggression, serious") may want their own micro-tone notes when we have more sample copy from the team.

See `SKILL.md` for the agent-skill entry point.
