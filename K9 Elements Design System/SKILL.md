---
name: k9-elements-design
description: Use this skill to generate well-branded interfaces and assets for K9 Elements, either for production or throwaway prototypes/mocks/etc. Contains essential design guidelines, colors, type, fonts, assets, and UI kit components for prototyping the K9 Elements brand — the working-dog sports organization that runs trials and supplies training material for the four primary events (Obedience, Protection, Tracking, Detection).
user-invocable: true
---

# K9 Elements design skill

Read `README.md` first — it contains the full system: company context, content tone, visual foundations, iconography, and the all-important event-color theming rules.

## What's in this folder

- `README.md` — single source of truth for the design system. Read this first.
- `colors_and_type.css` — every CSS custom property (colors, type ramp, radii, shadows, spacing, motion). Import this in any artifact.
- `assets/` — the K9 Elements logo plus a stroke-icon set (paw, flag, cone, sleeve are custom; the rest match the chassis style).
- `preview/` — design-system specimen cards. Don't reference these in production work; they're for review only.
- `ui_kits/marketing-site/` — pixel-fidelity recreation of the marketing site, factored into reusable JSX components.

## How to design with this brand

If creating **visual artifacts** (slides, mocks, throwaway prototypes, marketing-site variations, etc.):

1. Read `README.md` to internalize tone and visual rules.
2. Copy the assets you need (logo, relevant icons, the `colors_and_type.css` file) into your artifact's folder.
3. Reach for the existing UI-kit components first (`ui_kits/marketing-site/components.jsx` etc.) — re-use rather than re-invent.
4. Apply event theming **subtly** — never as full-section background fills. The 3px rail, eyebrow color, and progress ring are the canonical applications. See "The event theming system" in the README.

If working on **production code**:

1. Read `README.md` to internalize tone and visual rules.
2. The codebase is Tailwind v4 (Tailwind Plus "Oatmeal" chassis). Use the `mist-*` neutral scale and the event-scoped variables documented in `colors_and_type.css`. Drop `data-event="<id>"` on subtrees to retint accent.

## If invoked without further guidance

Ask the user what they want to build or design. Confirm the surface (marketing site / training app / trial brief / slide deck / social post) and the event focus (if any). Then act as an expert designer, producing HTML artifacts or production-ready code as appropriate.

**Critical reminders for any output:**
- Imagery does the heavy lifting on the marketing front; copy is restrained.
- Long-form prose belongs in the paid app (training paths, trial briefs, judging tools), not the marketing surface.
- No emoji, no exclamation points, no SaaS verbs ("Discover," "Unlock," "Empower").
- Sentence case headlines. ALL CAPS only on eyebrows and trial-status pills.
- Event colors are an accent system, not a fill system.
