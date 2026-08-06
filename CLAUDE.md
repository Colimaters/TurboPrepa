# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository state

This repository currently contains **specs only** — no application code has been written yet
(`specs/`, an empty `README.md`, and a Go-flavored `.gitignore` anticipating the stack below).
There are no build/lint/test commands to run because there is nothing to build yet. When
implementation starts, this file should be updated with real commands (`wails build`,
`go test ./...`, etc.).

Treat `specs/00-architecture-generale-et-design.md` through `specs/10-import-fichiers.md` as
the product/engineering spec for **TurboPrepa**, a French-language macOS desktop app to
organize study for four competitive law-enforcement exams (concours de la fonction publique) :
Gardien de la paix, Officier de police, Commissaire de police, Officier de gendarmerie.
`00-architecture-generale-et-design.md` is the umbrella doc — every other spec file
references it instead of repeating its rules, so read it first.

## Spec file map

| File | Covers |
|---|---|
| `00-architecture-generale-et-design.md` | Stack, packaging, design system, data model — common to all tabs |
| `01-accueil.md` | Home tab: quote of the day, progress summary, today's tasks |
| `02-matieres.md` | Matières tab: subjects, chapter status tracking, folders/files |
| `03-planning.md` | Planning tab: calendar, manual tasks, auto-scheduling |
| `04-annuaire.md` | Annuaire tab: index of case-law locations within Dalloz codes |
| `05-jurisprudence.md` | Jurisprudence tab: case-law summary sheets |
| `06-quiz-du-jour.md` | Daily quiz: 5 questions/day, 2-min timer, streaks |
| `07-veille-juridique.md` | Legal-watch tab: external links |
| `08-concours.md` | Exam structure per concours (écrit/oral/sport/admission) + oral prep |
| `09-textes-reference.md` | Reference regulatory texts per concours |
| `10-import-fichiers.md` | Cross-cutting file-import component used by Matières and Planning |

Each feature file lists its own "Critères d'acceptation" checklist — treat those as the
definition of done for that tab, not the prose above them.

## Target architecture (from `00-architecture-generale-et-design.md`)

- **Framework**: Wails — Go backend, HTML/CSS/JS frontend, rendered via macOS's native
  WebKit (no embedded Chromium). All sensitive logic (file access, DB access, auto-scheduling)
  must live in Go and be exposed to the frontend via Wails bindings, not implemented in JS.
- **Storage**: embedded SQLite, single file at
  `~/Library/Application Support/<AppName>/data.db`. Use `modernc.org/sqlite` (pure-Go, no
  CGO) — deliberately **not** `mattn/go-sqlite3`, to keep the build/distribution simple.
  `PRAGMA foreign_keys = ON`; use real foreign keys for hierarchical data.
- **Imported files** (Word/Pages/PDF/images/Excel/Numbers) are never stored as BLOBs in
  SQLite — they're copied to
  `~/Library/Application Support/<AppName>/fichiers/{uuid}-{original-name}`, with only the
  path/metadata stored in the DB.
- **Schema migrations**: versioned and idempotent (e.g. a `schema_migrations` table), run at
  startup. `data.db` must never be recreated from scratch on an app update — updates must
  never lose user data.
- **Packaging**: `wails build` → `.dmg` with a drag-to-Applications shortcut. The app is
  unsigned/unnotarized (no paid Apple Developer account), so first-launch requires
  right-click → Open; document this for the user. Fully offline except user-initiated clicks
  on Veille juridique links, which open in the system default browser.
- **Indicative table list** (names are the cross-spec reference — keep specs and schema in
  sync if they change): `matieres`, `chapitres`, `fichiers`, `taches_planning`,
  `annuaire_entrees`, `jurisprudence_fiches`, `quiz_questions`, `quiz_historique`,
  `citations_motivation`, `concours` / `concours_epreuves`, `oral_questions_suggerees`,
  `oral_questions_perso`, `textes_reference`, `veille_liens`.

## Cross-cutting rules that apply to every feature

- **No multi-user auth, no OCR/content extraction.** Imported files are opaque attachments
  (name, type, size, date) — never parse or extract their content automatically. This is
  called out explicitly and repeatedly in the specs; don't "improve" it unprompted.
- **No data duplication across tabs.** Accueil reads chapter statuses and Planning tasks
  rather than storing its own copies; Planning reads chapter/matière data from Matières;
  Annuaire and Jurisprudence can cross-link but remain distinct tables. When a spec says a
  tab "lit" (reads) data from another tab, implement it as a read, not a sync/copy.
- **Shared color panel**: a single list of 15 pastel colors is used for both matière default
  colors and individual task colors — must be stored once and reused, not duplicated between
  Planning and Matières.
- **Nothing is hardcoded to a fixed count.** The initial 18 matières and the 4 concours are
  seed data, not schema constraints — subjects must be addable/renameable/removable by the
  user without limit.
- **Verifiable content only.** Quotes on Accueil and regulatory/exam-structure info on
  Concours/Textes de référence must come from real, attributed sources — never invented or
  freely generated. Regulatory/exam content additionally needs a source/date and a
  vérifié / à confirmer status field the user can toggle, since exam rules change with
  reforms.
- **File import** is one reusable component (`10-import-fichiers.md`) used identically by
  Matières and Planning: single file, multi-file, or `.zip` (auto-extracted, unsupported
  formats inside skipped with an explicit message, not a hard failure). Supported formats:
  .doc/.docx, .pages, .pdf, .jpg/.jpeg/.png, .xlsx/.xls, .numbers.
- **Design system**: pastel palette, rounded/soft typography (Nunito/Quicksand/Poppins-like),
  globally adjustable font size (not a one-off fix), rounded cards with light shadows, visible
  hover/active states, and feedback on every save/add/delete action. Single nav (top bar OR
  sidebar, not both) across all 9 tabs, SPA-like transitions (no full page reloads).
