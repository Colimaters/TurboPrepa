# TurboPrepa Copilot Instructions

## Repository state and workflow

This repository currently contains product specifications only; it has no application source, dependency manifest, or test suite. There are therefore no build, lint, test, or single-test commands yet. Update this section with the actual commands when implementation introduces them.

Read `specs/00-architecture-generale-et-design.md` before implementing a feature: it defines the shared architecture, storage, data model, navigation, and design system. Then read the feature specification(s) involved. Each feature file's **Critères d'acceptation** checklist is its definition of done.

The application and UI copy are French. Preserve the established French domain terms and labels, including `Matières`, `Planning`, `Annuaire`, and `Jurisprudence`.

## Target architecture

TurboPrepa is an offline, single-user macOS desktop study organizer for four public-service law-enforcement competitive exams.

- Use Wails: Go owns sensitive business logic, file-system access, SQLite access, and automatic scheduling; expose those operations through Wails bindings. The frontend is HTML/CSS/JS and may use a lightweight framework, but must not implement that backend logic itself.
- Persist user data in one SQLite database at `~/Library/Application Support/<AppName>/data.db`, using `modernc.org/sqlite` rather than a CGO-dependent driver. Enable `PRAGMA foreign_keys = ON`.
- Make schema migrations versioned and idempotent at app startup. Never recreate `data.db` during an update or discard user data.
- Copy imported attachments to `~/Library/Application Support/<AppName>/fichiers/{uuid}-{original-name}`. Store only the path and metadata in SQLite, never BLOB file contents.
- Build and distribute the final application as a Wails macOS `.app` inside a `.dmg`; it is intentionally unsigned and unnotarized.

The app has nine separately implemented sections sharing one navigation system and data layer: Accueil, Matières, Planning, Annuaire, Jurisprudence, Veille juridique, Quiz du jour, Concours, and Textes de référence.

## Cross-feature conventions

- Treat initial content as seed data, not fixed schema limits: the 18 matières and four concours must remain user-editable and extensible.
- Model relationships with foreign keys. In particular, `fichiers` attaches to either a matière chapter/folder or a planning task; Annuaire location entries and Jurisprudence content sheets remain separate entities with optional cross-links.
- Do not duplicate cross-tab data. Accueil reads chapter status and planning tasks; Planning reads Matières and chapters; generated planning blocks become normal editable planning tasks.
- Define the 15-color pastel palette once and reuse it for matière defaults and individually overridden planning-task colors.
- Implement file import as one shared component for Matières and Planning. Accept `.doc`, `.docx`, `.pages`, `.pdf`, `.jpg`, `.jpeg`, `.png`, `.xlsx`, `.xls`, and `.numbers`; support single selection, multi-selection, and `.zip` expansion. Unsupported archive entries must be reported and skipped without failing supported imports.
- Imported files are opaque, downloadable attachments. Do not add OCR, content parsing, or automatic timetable extraction; make this limitation clear in the UI. Users can add, rename, and delete attachments after the parent item exists.
- Keep the application offline. Veille juridique links are the only user-initiated external navigation and must open in the system browser.

## Product and UI rules

- Use one persistent navigation pattern only (top bar or sidebar), with SPA-like tab changes rather than full page reloads.
- Apply the shared pastel, rounded visual system consistently. The font-size preference is global, not a per-screen workaround; actions need visible feedback.
- When a Playwright MCP server is configured and the frontend exists, use it to validate visible workflows such as tab navigation, chapter-status propagation, planning task edits, and file-import feedback. Keep browser automation focused on the specification acceptance criteria.
- Annuaire is a searchable index of where an arrêt appears in a Dalloz code. Jurisprudence contains the separate substantive case-law sheets (facts, solution, scope). Do not conflate them.
- Use deterministic date-based rotation without premature repeats for Accueil quotes and Quiz questions. Quotes, exam information, and legal references must be real and attributed; concours and reference-text entries require a source/date plus a user-toggleable `vérifié` / `à confirmer` status.
- Keep suggested oral questions distinct from the user's personal question/answer workspace.

## Specification map

| Specification | Primary responsibility |
| --- | --- |
| `01-accueil.md` | Daily quote, aggregate chapter progress, and today’s planning tasks |
| `02-matieres.md` | Editable subjects, chapter status, folders, and course attachments |
| `03-planning.md` | Calendar, manual tasks, attachment reference, and constraint-based automatic scheduling |
| `04-annuaire.md` / `05-jurisprudence.md` | Code-location index and substantive case-law sheets with optional links |
| `06-quiz-du-jour.md` | Five-question daily quiz, two-minute timer, streak, and history |
| `07-veille-juridique.md` | User-editable external legal-watch links |
| `08-concours.md` / `09-textes-reference.md` | Exam structures, oral preparation, and linked regulatory texts |
| `10-import-fichiers.md` | Reusable attachment-import behavior |
