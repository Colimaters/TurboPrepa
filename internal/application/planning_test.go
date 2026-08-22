package application

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestPlanningTasksAndGenerationUseExistingChapters(t *testing.T) {
	app := newTestApp(t, time.Date(2026, time.August, 17, 9, 0, 0, 0, time.Local))

	var subjectID int64
	if err := app.db.QueryRow(`SELECT id FROM matieres WHERE nom = 'Droit pénal général'`).Scan(&subjectID); err != nil {
		t.Fatalf("find subject: %v", err)
	}
	chapter, err := app.CreateChapter(subjectID, tabProgramme, "Les infractions", "")
	if err != nil {
		t.Fatalf("CreateChapter() error = %v", err)
	}
	task, err := app.CreatePlanningTask(PlanningTaskInput{
		MatiereID: &subjectID, ChapitreID: &chapter.ID, Title: "Relire les infractions",
		Date: "2026-08-17", StartTime: "08:00", EndTime: "09:00",
	})
	if err != nil {
		t.Fatalf("CreatePlanningTask() error = %v", err)
	}
	if task.SubjectName != "Droit pénal général" || task.ChapterName != "Les infractions" {
		t.Errorf("task links = %#v, want resolved subject and chapter", task)
	}
	if task.Color != "#BFD7ED" {
		t.Errorf("task color = %q, want the subject color", task.Color)
	}
	if _, err := app.CreatePlanningTask(PlanningTaskInput{Title: "Sans matière", Date: "2026-08-17", StartTime: "10:00", EndTime: "11:00"}); err == nil {
		t.Error("CreatePlanningTask() error = nil, want a required subject")
	}
	if _, err := app.MovePlanningTask(task.ID, "2026-08-18", "09:00", "10:00"); err != nil {
		t.Fatalf("MovePlanningTask() error = %v", err)
	}
	tasks, err := app.ListPlanningTasks("2026-08-18", "2026-08-18")
	if err != nil {
		t.Fatalf("ListPlanningTasks() error = %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Errorf("tasks = %#v, want moved task only", tasks)
	}

	generated, err := app.GeneratePlanning(GeneratePlanningInput{
		StartDate: "2026-08-17",
		Selections: []PlanningSelection{{
			ChapterID: chapter.ID, StartDays: []int{1, 2}, RevisionCount: 2, DurationMinutes: 60, SpacingDays: 1,
		}},
	})
	if err != nil {
		t.Fatalf("GeneratePlanning() error = %v", err)
	}
	if len(generated) != 2 || generated[0].StartTime != "08:00" || generated[1].Date != "2026-08-18" {
		t.Errorf("generated = %#v, want two scheduled revisions", generated)
	}
	var status string
	if err := app.db.QueryRow(`SELECT statut FROM chapitres WHERE id = ?`, chapter.ID).Scan(&status); err != nil {
		t.Fatalf("read chapter status: %v", err)
	}
	if status != "planifie" {
		t.Errorf("status = %q, want planifie", status)
	}
}

func TestPlanningWorkdayPreferencesValidateSlots(t *testing.T) {
	app := newTestApp(t, time.Date(2026, time.August, 17, 9, 0, 0, 0, time.Local))

	preferences, err := app.GetWorkdayPreferences()
	if err != nil {
		t.Fatalf("GetWorkdayPreferences() error = %v", err)
	}

	if len(preferences.Slots) != 3 || preferences.Slots[0].Start != "08:00" {
		t.Errorf("preferences = %#v, want default morning slot", preferences)
	}
	preferences.Slots[0].End = "07:00"
	if err := app.SaveWorkdayPreferences(preferences); err == nil {
		t.Error("SaveWorkdayPreferences() error = nil, want invalid time range")
	}
}

func TestPlanningTemplateIncludesGuide(t *testing.T) {
	path := filepath.Join(t.TempDir(), "modele.xlsx")
	if err := writePlanningTemplate(path); err != nil {
		t.Fatalf("writePlanningTemplate() error = %v", err)
	}

	book, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open template: %v", err)
	}
	defer book.Close()
	sheets := book.GetSheetList()
	hasPlanning, hasGuide := false, false
	for _, sheet := range sheets {
		hasPlanning = hasPlanning || sheet == planningTemplateSheet
		hasGuide = hasGuide || sheet == "Guide"
	}
	if !hasPlanning || !hasGuide {
		t.Errorf("sheets = %v, want planning sheet and Guide", book.GetSheetList())
	}
	if value, err := book.GetCellValue("Guide", "C2"); err != nil || value != "Format AAAA-MM-JJ, par exemple 2026-08-22." {
		t.Errorf("Guide!C2 = %q, %v", value, err)
	}
	if value, err := book.GetCellValue("Guide", "C6"); err != nil || value != "Nom exact d'une matière déjà créée dans TurboPrepa. Sa couleur est automatiquement appliquée à la tâche." {
		t.Errorf("Guide!C6 = %q, %v", value, err)
	}
	if value, err := book.GetCellValue(planningTemplateSheet, "G1"); err != nil || value != "notes" {
		t.Errorf("Emploi du temps!G1 = %q, %v", value, err)
	}
	if value, err := book.GetCellValue(planningTemplateSheet, "H1"); err != nil || value != "" {
		t.Errorf("Emploi du temps!H1 = %q, %v, want no color column", value, err)
	}
}

func TestImportPlanningTemplateValidatesThenInsertsTasksInBulk(t *testing.T) {
	app := newTestApp(t, time.Date(2026, time.August, 17, 9, 0, 0, 0, time.Local))
	var subjectID int64
	if err := app.db.QueryRow(`SELECT id FROM matieres WHERE nom = 'Droit pénal général'`).Scan(&subjectID); err != nil {
		t.Fatalf("find subject: %v", err)
	}
	chapter, err := app.CreateChapter(subjectID, tabProgramme, "Les infractions", "")
	if err != nil {
		t.Fatalf("CreateChapter() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "planning.xlsx")
	book := excelize.NewFile()
	book.SetSheetName("Sheet1", planningTemplateSheet)
	headers := []string{"date", "heure_debut", "heure_fin", "intitule", "matiere", "chapitre", "notes"}
	if err := book.SetSheetRow(planningTemplateSheet, "A1", &headers); err != nil {
		t.Fatalf("write headers: %v", err)
	}
	rows := [][]string{
		{"2026-08-18", "08:00", "09:00", "Réviser les infractions", "Droit pénal général", chapter.Name, ""},
		{"2026-08-18", "09:00", "10:00", "Relire le cours", "Droit pénal général", "", "Priorité"},
		{"2026-08-18", "10:00", "11:00", "Matière inconnue", "Inconnue", "", ""},
	}
	for index, row := range rows {
		if err := book.SetSheetRow(planningTemplateSheet, fmt.Sprintf("A%d", index+2), &row); err != nil {
			t.Fatalf("write row %d: %v", index+2, err)
		}
	}
	if err := book.SaveAs(path); err != nil {
		t.Fatalf("save workbook: %v", err)
	}
	if err := book.Close(); err != nil {
		t.Fatalf("close workbook: %v", err)
	}

	result, err := app.importPlanningTemplate(path)
	if err != nil {
		t.Fatalf("importPlanningTemplate() error = %v", err)
	}
	if len(result.Imported) != 2 || len(result.Skipped) != 1 {
		t.Fatalf("import result = %#v, want two imported rows and one skipped row", result)
	}
	tasks, err := app.ListPlanningTasks("2026-08-18", "2026-08-18")
	if err != nil {
		t.Fatalf("ListPlanningTasks() error = %v", err)
	}
	if len(tasks) != 2 || tasks[0].ChapitreID == nil || tasks[0].Notes != "" || tasks[1].Notes != "Priorité" {
		t.Errorf("tasks = %#v, want both valid rows inserted with their chapter and notes", tasks)
	}
}

func TestPlanningMigrationRemovesTaskColorColumn(t *testing.T) {
	app := newTestApp(t, time.Date(2026, time.August, 17, 9, 0, 0, 0, time.Local))
	rows, err := app.db.Query(`PRAGMA table_info(taches_planning)`)
	if err != nil {
		t.Fatalf("read task columns: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var columnID, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&columnID, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan task column: %v", err)
		}
		if name == "couleur" {
			t.Error("taches_planning still contains the removed couleur column")
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate task columns: %v", err)
	}
}
