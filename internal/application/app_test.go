package application

import (
	"archive/zip"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestApp(t *testing.T, now time.Time) *App {
	t.Helper()

	app := NewApp()
	app.dataDir = t.TempDir()
	app.now = func() time.Time { return now }
	if err := app.loadQuotes(); err != nil {
		t.Fatalf("loadQuotes() error = %v", err)
	}

	if err := app.openDatabase(); err != nil {
		t.Fatalf("openDatabase() error = %v", err)
	}
	t.Cleanup(func() { app.db.Close() })
	return app
}

func TestMatieresPersistContentAndAttachments(t *testing.T) {
	app := newTestApp(t, time.Date(2026, time.August, 16, 9, 0, 0, 0, time.Local))

	colors, err := app.ListPastelColors()
	if err != nil {
		t.Fatalf("ListPastelColors() error = %v", err)
	}
	if len(colors) != 15 {
		t.Fatalf("palette has %d colors, want 15", len(colors))
	}
	subject, err := app.CreateMatiere("Matière de test", colors[0])
	if err != nil {
		t.Fatalf("CreateMatiere() error = %v", err)
	}
	program, err := app.CreateChapter(subject.ID, tabProgramme, "Chapitre 1", "")
	if err != nil {
		t.Fatalf("CreateChapter(programme) error = %v", err)
	}
	folder, err := app.CreateChapter(subject.ID, tabFiches, "Fiches à revoir", "Texte de fiche")
	if err != nil {
		t.Fatalf("CreateChapter(fiches) error = %v", err)
	}
	if err := app.SetChapterStatus(program.ID, "maitrise"); err != nil {
		t.Fatalf("SetChapterStatus() error = %v", err)
	}
	work, err := app.CreateSubjectWork(subject.ID, "Relire le chapitre", "2026-08-20")
	if err != nil {
		t.Fatalf("CreateSubjectWork() error = %v", err)
	}
	if err := app.SetSubjectWorkCompleted(work.ID, true); err != nil {
		t.Fatalf("SetSubjectWorkCompleted() error = %v", err)
	}

	sourcePath := filepath.Join(t.TempDir(), "support.xlsx")
	if err := os.WriteFile(sourcePath, []byte("contenu"), 0o600); err != nil {
		t.Fatalf("write source attachment: %v", err)
	}
	result, err := app.ImportChapterFiles(folder.ID, []string{sourcePath})
	if err != nil {
		t.Fatalf("ImportChapterFiles() error = %v", err)
	}
	if len(result.Imported) != 1 || result.Imported[0].MimeType == "" {
		t.Fatalf("import result = %#v, want one attachment with a MIME type", result)
	}
	archivePath := filepath.Join(t.TempDir(), "documents.zip")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create ZIP: %v", err)
	}
	archive := zip.NewWriter(archiveFile)
	for name := range map[string]bool{"fiche.docx": true, "ignore.txt": true} {
		entry, err := archive.Create(name)
		if err != nil {
			t.Fatalf("create ZIP entry: %v", err)
		}
		if _, err := entry.Write([]byte("contenu")); err != nil {
			t.Fatalf("write ZIP entry: %v", err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close ZIP: %v", err)
	}
	if err := archiveFile.Close(); err != nil {
		t.Fatalf("close ZIP file: %v", err)
	}
	zipResult, err := app.ImportChapterFiles(folder.ID, []string{archivePath})
	if err != nil {
		t.Fatalf("ImportChapterFiles(ZIP) error = %v", err)
	}
	if len(zipResult.Imported) != 1 || len(zipResult.Skipped) != 1 || zipResult.Skipped[0] != "ignore.txt" {
		t.Fatalf("ZIP result = %#v, want one imported file and ignore.txt skipped", zipResult)
	}
	var storedPath string
	if err := app.db.QueryRow(`SELECT chemin_disque FROM fichiers WHERE id = ?`, result.Imported[0].ID).Scan(&storedPath); err != nil {
		t.Fatalf("read attachment path: %v", err)
	}
	if _, err := os.Stat(storedPath); err != nil {
		t.Fatalf("stored attachment does not exist: %v", err)
	}
	if err := app.MoveAttachment(result.Imported[0].ID, folder.ID); err != nil {
		t.Fatalf("MoveAttachment() error = %v", err)
	}

	detail, err := app.GetMatiereDetail(subject.ID)
	if err != nil {
		t.Fatalf("GetMatiereDetail() error = %v", err)
	}
	if detail.Subject.Mastered != 1 || detail.Subject.Chapters != 1 {
		t.Errorf("progress = %#v, want one mastered program chapter", detail.Subject)
	}
	if len(detail.Works) != 1 || !detail.Works[0].Completed {
		t.Errorf("works = %#v, want one completed work", detail.Works)
	}
	dashboard, err := app.GetDashboard()
	if err != nil {
		t.Fatalf("GetDashboard() error = %v", err)
	}
	if dashboard.Progress != (Progress{Mastered: 1}) {
		t.Errorf("dashboard progress = %#v, want only the program chapter", dashboard.Progress)
	}
	var persistedFolder *Chapter
	for index := range detail.Chapters {
		if detail.Chapters[index].ID == folder.ID {
			persistedFolder = &detail.Chapters[index]
		}
	}
	if len(detail.Chapters) != 2 || persistedFolder == nil || persistedFolder.Content != "Texte de fiche" || len(persistedFolder.Files) != 2 {
		t.Errorf("chapters = %#v, want persisted folders, content, and attachment", detail.Chapters)
	}
	if err := app.DeleteMatiere(subject.ID); err != nil {
		t.Fatalf("DeleteMatiere() error = %v", err)
	}
	if _, err := os.Stat(storedPath); !os.IsNotExist(err) {
		t.Errorf("stored attachment still exists after parent deletion, stat error = %v", err)
	}
}

func TestMigrationsAreIdempotentAndSeedSubjects(t *testing.T) {
	app := newTestApp(t, time.Date(2026, time.August, 16, 9, 0, 0, 0, time.Local))

	if err := app.migrate(); err != nil {
		t.Fatalf("second migrate() error = %v", err)
	}

	var subjectCount, migrationCount, foreignKeys int
	if err := app.db.QueryRow(`SELECT COUNT(*) FROM matieres`).Scan(&subjectCount); err != nil {
		t.Fatalf("count subjects: %v", err)
	}
	if err := app.db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&migrationCount); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if err := app.db.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		t.Fatalf("read foreign_keys: %v", err)
	}
	if subjectCount != 18 {
		t.Errorf("subject count = %d, want 18", subjectCount)
	}
	if migrationCount != len(migrations) {
		t.Errorf("migration count = %d, want %d", migrationCount, len(migrations))
	}
	if foreignKeys != 1 {
		t.Errorf("foreign_keys = %d, want 1", foreignKeys)
	}

	var chapterIDColumn, notesColumn bool
	rows, err := app.db.Query(`PRAGMA table_info(taches_planning)`)
	if err != nil {
		t.Fatalf("read task columns: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var columnID int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue sql.NullString
		if err := rows.Scan(&columnID, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan task column: %v", err)
		}
		chapterIDColumn = chapterIDColumn || name == "chapitre_id"
		notesColumn = notesColumn || name == "notes"
	}
	if !chapterIDColumn || !notesColumn {
		t.Errorf("planning task columns chapitre_id=%t notes=%t, want both", chapterIDColumn, notesColumn)
	}
}

func TestDashboardUsesSharedDataAndRotatesQuotes(t *testing.T) {
	now := time.Date(2026, time.August, 16, 9, 0, 0, 0, time.Local)
	app := newTestApp(t, now)

	var subjectID int64
	if err := app.db.QueryRow(`SELECT id FROM matieres WHERE nom = ?`, "Droit pénal général").Scan(&subjectID); err != nil {
		t.Fatalf("find seeded subject: %v", err)
	}
	for _, status := range []string{"a_planifier", "planifie", "en_cours", "maitrise"} {
		if _, err := app.db.Exec(`INSERT INTO chapitres(matiere_id, nom, statut, ordre) VALUES (?, ?, ?, 0)`, subjectID, status, status); err != nil {
			t.Fatalf("insert chapter %s: %v", status, err)
		}
	}
	if _, err := app.db.Exec(`
		INSERT INTO taches_planning(matiere_id, titre, date, heure_debut, heure_fin, terminee)
		VALUES (?, 'Tâche du jour', ?, '08:00', '09:00', 0),
		       (?, 'Tâche demain', '2026-08-17', '08:00', '09:00', 0)`,
		subjectID, now.Format("2006-01-02"), subjectID); err != nil {
		t.Fatalf("insert tasks: %v", err)
	}

	dashboard, err := app.GetDashboard()
	if err != nil {
		t.Fatalf("GetDashboard() error = %v", err)
	}
	if dashboard.Progress != (Progress{ToPlan: 1, Planned: 1, InProgress: 1, Mastered: 1}) {
		t.Errorf("progress = %#v, want one chapter per status", dashboard.Progress)
	}
	if len(dashboard.Tasks) != 1 || dashboard.Tasks[0].Title != "Tâche du jour" {
		t.Errorf("tasks = %#v, want only today's task", dashboard.Tasks)
	}
	if dashboard.Tasks[0].Color == "" || dashboard.Tasks[0].Subject == "" {
		t.Errorf("task did not resolve subject color/name: %#v", dashboard.Tasks[0])
	}

	firstQuote := dashboard.Quote
	app.now = func() time.Time { return now.AddDate(0, 0, 1) }
	nextDashboard, err := app.GetDashboard()
	if err != nil {
		t.Fatalf("next GetDashboard() error = %v", err)
	}
	if firstQuote == nextDashboard.Quote {
		t.Error("quote repeated on consecutive days")
	}
	app.now = func() time.Time { return now.AddDate(0, 0, len(app.quotes)) }
	cycledDashboard, err := app.GetDashboard()
	if err != nil {
		t.Fatalf("cycled GetDashboard() error = %v", err)
	}
	if firstQuote != cycledDashboard.Quote {
		t.Error("quote did not repeat after the complete rotation")
	}
}

func TestToggleTodayTaskOnlyUpdatesToday(t *testing.T) {
	now := time.Date(2026, time.August, 16, 9, 0, 0, 0, time.Local)
	app := newTestApp(t, now)

	result, err := app.db.Exec(`
		INSERT INTO taches_planning(titre, date, heure_debut, heure_fin)
		VALUES ('Aujourd''hui', ?, '08:00', '09:00')`, now.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("insert today task: %v", err)
	}
	todayID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("today task ID: %v", err)
	}
	if err := app.ToggleTodayTask(todayID, true); err != nil {
		t.Fatalf("ToggleTodayTask(today) error = %v", err)
	}
	var completed int
	if err := app.db.QueryRow(`SELECT terminee FROM taches_planning WHERE id = ?`, todayID).Scan(&completed); err != nil {
		t.Fatalf("read completed task: %v", err)
	}
	if completed != 1 {
		t.Errorf("completed = %d, want 1", completed)
	}

	result, err = app.db.Exec(`
		INSERT INTO taches_planning(titre, date, heure_debut, heure_fin)
		VALUES ('Demain', '2026-08-17', '08:00', '09:00')`)
	if err != nil {
		t.Fatalf("insert tomorrow task: %v", err)
	}
	tomorrowID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("tomorrow task ID: %v", err)
	}
	if err := app.ToggleTodayTask(tomorrowID, true); err == nil {
		t.Error("ToggleTodayTask(tomorrow) error = nil, want an error")
	}
	if err := app.ToggleTodayTask(99999, true); err == nil {
		t.Error("ToggleTodayTask(missing) error = nil, want an error")
	}
}

func TestExistingVersionOneDatabaseMigratesWithoutReplacingData(t *testing.T) {
	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "data.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY);
		INSERT INTO schema_migrations(version) VALUES (1);
		CREATE TABLE matieres (id INTEGER PRIMARY KEY, nom TEXT NOT NULL, couleur TEXT NOT NULL, ordre INTEGER NOT NULL);
		CREATE TABLE chapitres (
			id INTEGER PRIMARY KEY, matiere_id INTEGER NOT NULL, nom TEXT NOT NULL,
			statut TEXT NOT NULL, ordre INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE taches_planning (
			id INTEGER PRIMARY KEY, matiere_id INTEGER, titre TEXT NOT NULL, date TEXT NOT NULL,
			heure_debut TEXT NOT NULL, heure_fin TEXT NOT NULL, couleur TEXT, terminee INTEGER NOT NULL DEFAULT 0
		);
		INSERT INTO matieres(nom, couleur, ordre) VALUES ('Matière conservée', '#ffffff', 99);`); err != nil {
		db.Close()
		t.Fatalf("prepare legacy database: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	app := NewApp()
	app.dataDir = dataDir
	app.now = time.Now
	if err := app.loadQuotes(); err != nil {
		t.Fatalf("loadQuotes() error = %v", err)
	}
	if err := app.openDatabase(); err != nil {
		t.Fatalf("open migrated database: %v", err)
	}
	t.Cleanup(func() { app.db.Close() })

	var subjectCount int
	if err := app.db.QueryRow(`SELECT COUNT(*) FROM matieres WHERE nom = 'Matière conservée'`).Scan(&subjectCount); err != nil {
		t.Fatalf("find preserved subject: %v", err)
	}
	if subjectCount != 1 {
		t.Errorf("preserved subject count = %d, want 1", subjectCount)
	}
}
