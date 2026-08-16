package main

import (
	"database/sql"
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
	if subjectCount != len(defaultSubjects) {
		t.Errorf("subject count = %d, want %d", subjectCount, len(defaultSubjects))
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
	if err := app.db.QueryRow(`SELECT id FROM matieres WHERE nom = ?`, defaultSubjects[0]).Scan(&subjectID); err != nil {
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
