# Marketing site — UI kit

A working visual recreation of the K9 Elements public marketing site. Imagery-forward, copy-light, with subtle event theming.

## Pages

| Route | Built from | Notes |
|---|---|---|
| **Home** | `Hero` + `EventGrid` + `TrialListSection (limit=3)` + `Testimonials` + `PricingTiers` | The default landing experience. |
| **Trials** | `TrialListSection` (full list) | The public schedule. |
| **Training** | `TrainingPage` (renders one `PathRow` per event) | Front door to the paid product. |
| **Pricing** | `PricingTiers` + simple contact section | Three-tier plan layout. |

Navigation is in-page state (no real routing). The shared `window.kit.navigate(route)` API is mounted on `App` startup and consumed by links and CTAs throughout the kit. The trial detail drawer (`TrialDetailDrawer`) opens on top of any page via `window.kit.openTrial(trial)`.

## Components factored

- `components.jsx` — `Container`, `Button`, `Pill`, `EventPill`, `Eyebrow`, `Icon`, `Logo`, `NavBar`, `Footer`, `PhotoPlaceholder`.
- `Hero.jsx` — `Hero` (left-aligned, large h1, lede, two CTAs, full-bleed photo).
- `EventGrid.jsx` — `EventGrid` (2×2 tile grid; each tile has the 3px event-color rail + themed eyebrow).
- `TrialList.jsx` — `TrialCard`, `TrialListSection` (used in both home & /trials).
- `Testimonials.jsx` — `Testimonials` (3-column quotes; each tinted by event).
- `PricingTiers.jsx` — `Tier`, `PricingTiers`.
- `TrainingPage.jsx` — `PathRow`, `TrainingPage` (full /training page).
- `TrialDetailDrawer.jsx` — slide-in drawer with trial schedule + entry CTA.

## How event theming actually shows up here

This kit is the canonical answer to "what does subtle event theming look like in practice?" — drop `data-event="<id>"` on any container and:

1. The **3px rail** on `.event-tile` reads `var(--event-600)`.
2. The **eyebrow** color (`.eyebrow`) reads `var(--event-600)`.
3. The **`btn-event`** variant fills with `var(--event-600)` (used for "Enter trial" and "Start path").
4. The **path-progress ring** in the training kit reads `var(--event-600)`.

Everywhere else, surfaces stay neutral mist. The brand never floods a viewport with event color.

## Imagery

`PhotoPlaceholder` is a CSS-gradient stand-in for real video stills. The marketing site brief calls out that imagery is the brand's main visual driver — once real photo/video lands, drop those into `<img>` or `<video>` and remove the gradient class.

## How to run

Open `index.html` directly. No build step. React + Babel via CDN.
