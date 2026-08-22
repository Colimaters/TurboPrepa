package application

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

const planningTemplateSheet = "Emploi du temps"

type PlanningTask struct {
	ID          int64  `json:"id"`
	MatiereID   *int64 `json:"matiereId,omitempty"`
	ChapitreID  *int64 `json:"chapitreId,omitempty"`
	Title       string `json:"title"`
	Date        string `json:"date"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Color       string `json:"color"`
	Notes       string `json:"notes"`
	Completed   bool   `json:"completed"`
	SubjectName string `json:"subjectName"`
	ChapterName string `json:"chapterName"`
}

type PlanningTaskInput struct {
	MatiereID  *int64 `json:"matiereId"`
	ChapitreID *int64 `json:"chapitreId"`
	Title      string `json:"title"`
	Date       string `json:"date"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	Notes      string `json:"notes"`
}

type PlanningChapter struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PlanningSubject struct {
	ID       int64             `json:"id"`
	Name     string            `json:"name"`
	Color    string            `json:"color"`
	Chapters []PlanningChapter `json:"chapters"`
}

type PlanningData struct {
	Subjects []PlanningSubject `json:"subjects"`
}

type WorkdaySlot struct {
	Period  string `json:"period"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Enabled bool   `json:"enabled"`
}

type WorkdayPreferences struct {
	Slots []WorkdaySlot `json:"slots"`
}

type PlanningSelection struct {
	ChapterID       int64 `json:"chapterId"`
	StartDays       []int `json:"startDays"`
	RevisionCount   int   `json:"revisionCount"`
	DurationMinutes int   `json:"durationMinutes"`
	SpacingDays     int   `json:"spacingDays"`
}

type GeneratePlanningInput struct {
	Selections []PlanningSelection `json:"selections"`
	StartDate  string              `json:"startDate"`
}

type PlanningImportStatus struct {
	State   string `json:"state"`
	Message string `json:"message"`
}

func migratePlanning(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE planning_workday_slots (
			period TEXT PRIMARY KEY CHECK(period IN ('morning', 'afternoon', 'evening')),
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1 CHECK(enabled IN (0, 1))
		);
		INSERT INTO planning_workday_slots(period, start_time, end_time, enabled) VALUES
			('morning', '08:00', '12:00', 1),
			('afternoon', '14:00', '19:00', 1),
			('evening', '20:00', '22:00', 1);
		CREATE INDEX idx_taches_planning_date_time ON taches_planning(date, heure_debut, heure_fin);
		CREATE INDEX idx_taches_planning_chapter ON taches_planning(chapitre_id);
		CREATE INDEX idx_fichiers_tache_ordre ON fichiers(tache_id, ordre);`)
	return err
}

func (a *App) ListPlanningData() (PlanningData, error) {
	rows, err := a.db.Query(`SELECT m.id, m.nom, m.couleur, c.id, c.nom, c.statut
		FROM matieres m LEFT JOIN chapitres c ON c.matiere_id = m.id AND c.sous_onglet = 'programme'
		ORDER BY m.ordre, m.id, CASE c.statut WHEN 'a_planifier' THEN 0 ELSE 1 END, c.ordre, c.id`)
	if err != nil {
		return PlanningData{}, fmt.Errorf("lecture des chapitres du planning : %w", err)
	}
	defer rows.Close()
	data := PlanningData{Subjects: []PlanningSubject{}}
	byID := map[int64]int{}
	for rows.Next() {
		var subjectID int64
		var subjectName, subjectColor string
		var chapterID sql.NullInt64
		var chapterName, chapterStatus sql.NullString
		if err := rows.Scan(&subjectID, &subjectName, &subjectColor, &chapterID, &chapterName, &chapterStatus); err != nil {
			return data, err
		}
		index, exists := byID[subjectID]
		if !exists {
			index = len(data.Subjects)
			byID[subjectID] = index
			data.Subjects = append(data.Subjects, PlanningSubject{ID: subjectID, Name: subjectName, Color: subjectColor, Chapters: []PlanningChapter{}})
		}
		if chapterID.Valid {
			data.Subjects[index].Chapters = append(data.Subjects[index].Chapters, PlanningChapter{ID: chapterID.Int64, Name: chapterName.String, Status: chapterStatus.String})
		}
	}
	return data, rows.Err()
}

func (a *App) ListPlanningTasks(startDate, endDate string) ([]PlanningTask, error) {
	if _, err := parseDate(startDate); err != nil {
		return nil, fmt.Errorf("date de début : %w", err)
	}
	if _, err := parseDate(endDate); err != nil {
		return nil, fmt.Errorf("date de fin : %w", err)
	}
	rows, err := a.db.Query(`SELECT t.id, t.matiere_id, t.chapitre_id, t.titre, t.date, t.heure_debut, t.heure_fin,
		COALESCE(m.couleur, '#D4D4D4'), t.notes, t.terminee, COALESCE(m.nom, ''), COALESCE(c.nom, '')
		FROM taches_planning t
		LEFT JOIN matieres m ON m.id = t.matiere_id
		LEFT JOIN chapitres c ON c.id = t.chapitre_id
		WHERE t.date BETWEEN ? AND ? ORDER BY t.date, t.heure_debut, t.heure_fin, t.id`, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("lecture des tâches : %w", err)
	}
	defer rows.Close()
	tasks := []PlanningTask{}
	for rows.Next() {
		var task PlanningTask
		var subjectID, chapterID sql.NullInt64
		var completed int
		if err := rows.Scan(&task.ID, &subjectID, &chapterID, &task.Title, &task.Date, &task.StartTime, &task.EndTime, &task.Color, &task.Notes, &completed, &task.SubjectName, &task.ChapterName); err != nil {
			return nil, err
		}
		if subjectID.Valid {
			value := subjectID.Int64
			task.MatiereID = &value
		}
		if chapterID.Valid {
			value := chapterID.Int64
			task.ChapitreID = &value
		}
		task.Completed = completed == 1
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (a *App) CreatePlanningTask(input PlanningTaskInput) (PlanningTask, error) {
	if err := a.validatePlanningTaskInput(&input); err != nil {
		return PlanningTask{}, err
	}
	var id int64
	err := a.db.QueryRow(`INSERT INTO taches_planning(matiere_id, chapitre_id, titre, date, heure_debut, heure_fin, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		input.MatiereID, input.ChapitreID, input.Title, input.Date, input.StartTime, input.EndTime, input.Notes).Scan(&id)
	if err != nil {
		return PlanningTask{}, fmt.Errorf("création de la tâche : %w", err)
	}
	return a.getPlanningTask(id)
}

func (a *App) UpdatePlanningTask(taskID int64, input PlanningTaskInput) (PlanningTask, error) {
	if err := a.validatePlanningTaskInput(&input); err != nil {
		return PlanningTask{}, err
	}
	result, err := a.db.Exec(`UPDATE taches_planning SET matiere_id = ?, chapitre_id = ?, titre = ?, date = ?, heure_debut = ?, heure_fin = ?, notes = ? WHERE id = ?`,
		input.MatiereID, input.ChapitreID, input.Title, input.Date, input.StartTime, input.EndTime, input.Notes, taskID)
	if err := a.requireAffected(result, err, "tâche"); err != nil {
		return PlanningTask{}, err
	}
	return a.getPlanningTask(taskID)
}

func (a *App) DeletePlanningTask(taskID int64) error {
	return a.deleteWithAttachments(`SELECT chemin_disque FROM fichiers WHERE tache_id = ?`, `DELETE FROM taches_planning WHERE id = ?`, "tâche", taskID)
}

func (a *App) MovePlanningTask(taskID int64, date, startTime, endTime string) (PlanningTask, error) {
	if _, err := parseDate(date); err != nil {
		return PlanningTask{}, fmt.Errorf("date : %w", err)
	}
	if err := validateTimes(startTime, endTime); err != nil {
		return PlanningTask{}, err
	}
	result, err := a.db.Exec(`UPDATE taches_planning SET date = ?, heure_debut = ?, heure_fin = ? WHERE id = ?`, date, startTime, endTime, taskID)
	if err := a.requireAffected(result, err, "tâche"); err != nil {
		return PlanningTask{}, err
	}
	return a.getPlanningTask(taskID)
}

func (a *App) GetWorkdayPreferences() (WorkdayPreferences, error) {
	rows, err := a.db.Query(`SELECT period, start_time, end_time, enabled FROM planning_workday_slots
		ORDER BY CASE period WHEN 'morning' THEN 0 WHEN 'afternoon' THEN 1 ELSE 2 END`)
	if err != nil {
		return WorkdayPreferences{}, err
	}
	defer rows.Close()
	preferences := WorkdayPreferences{Slots: []WorkdaySlot{}}
	for rows.Next() {
		var slot WorkdaySlot
		var enabled int
		if err := rows.Scan(&slot.Period, &slot.Start, &slot.End, &enabled); err != nil {
			return preferences, err
		}
		slot.Enabled = enabled == 1
		preferences.Slots = append(preferences.Slots, slot)
	}
	return preferences, rows.Err()
}

func (a *App) SaveWorkdayPreferences(preferences WorkdayPreferences) error {
	if len(preferences.Slots) != 3 {
		return fmt.Errorf("les trois plages de la journée doivent être définies")
	}
	periods := map[string]bool{}
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	for _, slot := range preferences.Slots {
		if periods[slot.Period] || (slot.Period != "morning" && slot.Period != "afternoon" && slot.Period != "evening") {
			tx.Rollback()
			return fmt.Errorf("plage de travail invalide")
		}
		if err := validateTimes(slot.Start, slot.End); err != nil {
			tx.Rollback()
			return err
		}
		periods[slot.Period] = true
		if _, err := tx.Exec(`UPDATE planning_workday_slots SET start_time = ?, end_time = ?, enabled = ? WHERE period = ?`, slot.Start, slot.End, boolToInt(slot.Enabled), slot.Period); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (a *App) GeneratePlanning(input GeneratePlanningInput) ([]PlanningTask, error) {
	if len(input.Selections) == 0 {
		return nil, fmt.Errorf("sélectionnez au moins un chapitre")
	}
	start, err := a.generationStartDate(input.StartDate)
	if err != nil {
		return nil, err
	}
	preferences, err := a.GetWorkdayPreferences()
	if err != nil {
		return nil, err
	}
	slots := enabledSlots(preferences.Slots)
	if len(slots) == 0 {
		return nil, fmt.Errorf("aucune plage de travail n'est activée")
	}
	tx, err := a.db.Begin()
	if err != nil {
		return nil, err
	}
	created := []PlanningTask{}
	for _, selection := range input.Selections {
		if err := validateSelection(selection); err != nil {
			tx.Rollback()
			return nil, err
		}
		var subjectID int64
		var title, color, subjectName string
		if err := tx.QueryRow(`SELECT c.matiere_id, c.nom, m.couleur, m.nom FROM chapitres c JOIN matieres m ON m.id = c.matiere_id WHERE c.id = ? AND c.sous_onglet = 'programme'`, selection.ChapterID).Scan(&subjectID, &title, &color, &subjectName); err != nil {
			tx.Rollback()
			return nil, notFound("chapitre du programme", err)
		}
		cursor := firstPreferredDate(start, selection.StartDays)
		for revision := 0; revision < selection.RevisionCount; revision++ {
			if revision > 0 {
				cursor = firstPreferredDate(cursor.AddDate(0, 0, selection.SpacingDays), selection.StartDays)
			}
			date, startTime, endTime, err := a.findAvailableSlot(tx, cursor, selection.DurationMinutes, slots)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("%s, révision %d : %w", title, revision+1, err)
			}
			var taskID int64
			taskTitle := title
			if selection.RevisionCount > 1 {
				taskTitle = fmt.Sprintf("%s — révision %d/%d", title, revision+1, selection.RevisionCount)
			}
			if err := tx.QueryRow(`INSERT INTO taches_planning(matiere_id, chapitre_id, titre, date, heure_debut, heure_fin, notes)
				VALUES (?, ?, ?, ?, ?, ?, '') RETURNING id`, subjectID, selection.ChapterID, taskTitle, date, startTime, endTime).Scan(&taskID); err != nil {
				tx.Rollback()
				return nil, err
			}
			created = append(created, PlanningTask{ID: taskID, MatiereID: &subjectID, ChapitreID: &selection.ChapterID, Title: taskTitle, Date: date, StartTime: startTime, EndTime: endTime, Color: color, SubjectName: subjectName, ChapterName: title})
			cursor = mustDate(date)
		}
		if _, err := tx.Exec(`UPDATE chapitres SET statut = 'planifie' WHERE id = ? AND statut = 'a_planifier'`, selection.ChapterID); err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return created, nil
}

func (a *App) DownloadPlanningTemplate() error {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{Title: "Enregistrer le modèle d'emploi du temps", DefaultFilename: "modele-emploi-du-temps.xlsx", Filters: []runtime.FileFilter{{DisplayName: "Fichier Excel", Pattern: "*.xlsx"}}})
	if err != nil {
		return fmt.Errorf("choix du fichier : %w", err)
	}
	if path == "" {
		return nil
	}
	return writePlanningTemplate(path)
}

func (a *App) ImportPlanningTemplate() (ImportResult, error) {
	a.reportPlanningImport("selecting", "Sélectionnez votre fichier Excel dans la fenêtre qui vient de s’ouvrir.")
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Importer un emploi du temps", Filters: []runtime.FileFilter{{DisplayName: "Fichier Excel", Pattern: "*.xlsx"}}})
	if err != nil {
		return ImportResult{}, fmt.Errorf("sélection du fichier : %w", err)
	}
	if path == "" {
		return ImportResult{Imported: []Attachment{}, Skipped: []string{}}, nil
	}
	a.reportPlanningImport("processing", fmt.Sprintf("Fichier « %s » sélectionné. Analyse des créneaux en cours…", filepath.Base(path)))
	return a.importPlanningTemplate(path)
}

func (a *App) reportPlanningImport(state, message string) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "planning-import-status", PlanningImportStatus{State: state, Message: message})
	}
}

func writePlanningTemplate(path string) error {
	book := excelize.NewFile()
	defer book.Close()
	book.SetSheetName("Sheet1", planningTemplateSheet)
	headers := []string{"date", "heure_debut", "heure_fin", "intitule", "matiere", "chapitre", "notes"}
	if err := book.SetSheetRow(planningTemplateSheet, "A1", &headers); err != nil {
		return err
	}
	if err := book.SetColWidth(planningTemplateSheet, "A", "G", 20); err != nil {
		return err
	}
	const guideSheet = "Guide"
	if _, err := book.NewSheet(guideSheet); err != nil {
		return err
	}
	if err := book.SetSheetRow(guideSheet, "A1", &[]string{"Colonne", "Obligatoire", "Format et valeurs attendus"}); err != nil {
		return err
	}
	guideRows := [][]string{
		{"date", "Oui", "Format AAAA-MM-JJ, par exemple 2026-08-22."},
		{"heure_debut", "Oui", "Format HH:MM sur 24 heures, par exemple 08:00."},
		{"heure_fin", "Oui", "Format HH:MM sur 24 heures, après heure_debut, par exemple 09:30."},
		{"intitule", "Oui", "Titre de la tâche, par exemple « Relire le chapitre 1 »."},
		{"matiere", "Oui", "Nom exact d'une matière déjà créée dans TurboPrepa. Sa couleur est automatiquement appliquée à la tâche."},
		{"chapitre", "Non", "Nom exact d'un chapitre Programme de la matière indiquée. Laisser vide si aucun chapitre n'est associé."},
		{"notes", "Non", "Texte libre affiché dans le détail de la tâche."},
	}
	for index, row := range guideRows {
		if err := book.SetSheetRow(guideSheet, fmt.Sprintf("A%d", index+2), &row); err != nil {
			return err
		}
	}
	headerStyle, err := book.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"DDE4F5"}, Pattern: 1},
	})
	if err != nil {
		return err
	}
	if err := book.SetCellStyle(guideSheet, "A1", "C1", headerStyle); err != nil {
		return err
	}
	if err := book.SetColWidth(guideSheet, "A", "A", 22); err != nil {
		return err
	}
	if err := book.SetColWidth(guideSheet, "B", "B", 14); err != nil {
		return err
	}
	if err := book.SetColWidth(guideSheet, "C", "C", 100); err != nil {
		return err
	}
	return book.SaveAs(path)
}

func (a *App) importPlanningTemplate(path string) (ImportResult, error) {
	book, err := excelize.OpenFile(path)
	if err != nil {
		return ImportResult{}, fmt.Errorf("lecture du fichier Excel : %w", err)
	}
	defer book.Close()
	rows, err := book.GetRows(planningTemplateSheet)
	if err != nil {
		return ImportResult{}, fmt.Errorf("lecture de l'onglet %q : %w", planningTemplateSheet, err)
	}
	expected := []string{"date", "heure_debut", "heure_fin", "intitule", "matiere", "chapitre", "notes"}
	if len(rows) == 0 || !sameHeaders(rows[0], expected) {
		return ImportResult{}, fmt.Errorf("le fichier doit utiliser le modèle TurboPrepa et ses colonnes attendues")
	}
	result := ImportResult{Imported: []Attachment{}, Skipped: []string{}}
	entries := make([]planningImportEntry, 0, len(rows)-1)
	for index, row := range rows[1:] {
		if emptyRow(row) {
			continue
		}
		values := paddedRow(row, len(expected))
		entries = append(entries, planningImportEntry{
			line:    index + 2,
			input:   spreadsheetTaskInput(values),
			subject: strings.TrimSpace(values[4]),
			chapter: strings.TrimSpace(values[5]),
		})
	}

	references, err := a.loadPlanningImportReferences()
	if err != nil {
		return result, err
	}
	validInputs := make([]PlanningTaskInput, 0, len(entries))
	for _, entry := range entries {
		input, err := references.validateSpreadsheetTask(entry)
		if err != nil {
			result.Skipped = append(result.Skipped, fmt.Sprintf("ligne %d : %v", entry.line, err))
			continue
		}
		validInputs = append(validInputs, input)
		result.Imported = append(result.Imported, Attachment{DisplayName: fmt.Sprintf("ligne %d", entry.line)})
	}
	if len(validInputs) == 0 {
		return result, nil
	}

	tx, err := a.db.Begin()
	if err != nil {
		return ImportResult{}, err
	}
	if err := insertPlanningTasks(tx, validInputs); err != nil {
		tx.Rollback()
		return ImportResult{}, fmt.Errorf("import des créneaux : %w", err)
	}
	if err := tx.Commit(); err != nil {
		return ImportResult{}, err
	}
	return result, nil
}

type planningImportEntry struct {
	line    int
	input   PlanningTaskInput
	subject string
	chapter string
}

type planningImportReferences struct {
	subjects map[string]int64
	chapters map[int64]map[string]int64
}

func spreadsheetTaskInput(values []string) PlanningTaskInput {
	return PlanningTaskInput{
		Date:      strings.TrimSpace(values[0]),
		StartTime: strings.TrimSpace(values[1]),
		EndTime:   strings.TrimSpace(values[2]),
		Title:     strings.TrimSpace(values[3]),
		Notes:     strings.TrimSpace(values[6]),
	}
}

func (a *App) loadPlanningImportReferences() (planningImportReferences, error) {
	references := planningImportReferences{
		subjects: map[string]int64{},
		chapters: map[int64]map[string]int64{},
	}
	subjectRows, err := a.db.Query(`SELECT id, nom FROM matieres`)
	if err != nil {
		return references, fmt.Errorf("lecture des matières : %w", err)
	}
	for subjectRows.Next() {
		var id int64
		var name string
		if err := subjectRows.Scan(&id, &name); err != nil {
			subjectRows.Close()
			return references, fmt.Errorf("lecture d'une matière : %w", err)
		}
		references.subjects[name] = id
	}
	if err := subjectRows.Err(); err != nil {
		subjectRows.Close()
		return references, fmt.Errorf("parcours des matières : %w", err)
	}
	if err := subjectRows.Close(); err != nil {
		return references, fmt.Errorf("fermeture des matières : %w", err)
	}

	chapterRows, err := a.db.Query(`SELECT id, matiere_id, nom FROM chapitres WHERE sous_onglet = 'programme'`)
	if err != nil {
		return references, fmt.Errorf("lecture des chapitres : %w", err)
	}
	for chapterRows.Next() {
		var id, subjectID int64
		var name string
		if err := chapterRows.Scan(&id, &subjectID, &name); err != nil {
			chapterRows.Close()
			return references, fmt.Errorf("lecture d'un chapitre : %w", err)
		}
		if references.chapters[subjectID] == nil {
			references.chapters[subjectID] = map[string]int64{}
		}
		references.chapters[subjectID][name] = id
	}
	if err := chapterRows.Err(); err != nil {
		chapterRows.Close()
		return references, fmt.Errorf("parcours des chapitres : %w", err)
	}
	if err := chapterRows.Close(); err != nil {
		return references, fmt.Errorf("fermeture des chapitres : %w", err)
	}
	return references, nil
}

func (references planningImportReferences) validateSpreadsheetTask(entry planningImportEntry) (PlanningTaskInput, error) {
	input := entry.input
	if input.Title == "" {
		return input, fmt.Errorf("l'intitulé est obligatoire")
	}
	if _, err := parseDate(input.Date); err != nil {
		return input, fmt.Errorf("date : %w", err)
	}
	if err := validateTimes(input.StartTime, input.EndTime); err != nil {
		return input, err
	}
	if entry.subject == "" {
		return input, fmt.Errorf("la matière est obligatoire")
	}
	subjectID, exists := references.subjects[entry.subject]
	if !exists {
		return input, fmt.Errorf("matière %q introuvable", entry.subject)
	}
	input.MatiereID = &subjectID
	if entry.chapter == "" {
		return input, nil
	}
	chapterID, exists := references.chapters[subjectID][entry.chapter]
	if !exists {
		return input, fmt.Errorf("chapitre %q introuvable", entry.chapter)
	}
	input.ChapitreID = &chapterID
	return input, nil
}

func insertPlanningTasks(tx *sql.Tx, inputs []PlanningTaskInput) error {
	var query strings.Builder
	query.WriteString(`INSERT INTO taches_planning(matiere_id, chapitre_id, titre, date, heure_debut, heure_fin, notes) VALUES `)
	arguments := make([]any, 0, len(inputs)*7)
	for index, input := range inputs {
		if index > 0 {
			query.WriteString(", ")
		}
		query.WriteString("(?, ?, ?, ?, ?, ?, ?)")
		arguments = append(arguments, input.MatiereID, input.ChapitreID, input.Title, input.Date, input.StartTime, input.EndTime, input.Notes)
	}
	_, err := tx.Exec(query.String(), arguments...)
	return err
}

func (a *App) validatePlanningTaskInput(input *PlanningTaskInput) error {
	input.Title = strings.TrimSpace(input.Title)
	input.Date = strings.TrimSpace(input.Date)
	input.StartTime = strings.TrimSpace(input.StartTime)
	input.EndTime = strings.TrimSpace(input.EndTime)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Title == "" {
		return fmt.Errorf("l'intitulé est obligatoire")
	}
	if _, err := parseDate(input.Date); err != nil {
		return fmt.Errorf("date : %w", err)
	}
	if err := validateTimes(input.StartTime, input.EndTime); err != nil {
		return err
	}
	if input.ChapitreID != nil {
		var chapterSubjectID int64
		if err := a.db.QueryRow(`SELECT matiere_id FROM chapitres WHERE id = ? AND sous_onglet = 'programme'`, *input.ChapitreID).Scan(&chapterSubjectID); err != nil {
			return notFound("chapitre du programme", err)
		}
		if input.MatiereID == nil {
			input.MatiereID = &chapterSubjectID
		} else if *input.MatiereID != chapterSubjectID {
			return fmt.Errorf("le chapitre ne correspond pas à la matière")
		}
	}
	if input.MatiereID == nil {
		return fmt.Errorf("la matière est obligatoire")
	}
	var exists bool
	if err := a.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM matieres WHERE id = ?)`, *input.MatiereID).Scan(&exists); err != nil || !exists {
		if err != nil {
			return err
		}
		return fmt.Errorf("matière introuvable")
	}
	return nil
}

func (a *App) getPlanningTask(taskID int64) (PlanningTask, error) {
	tasks, err := a.ListPlanningTasks("0001-01-01", "9999-12-31")
	if err != nil {
		return PlanningTask{}, err
	}
	for _, task := range tasks {
		if task.ID == taskID {
			return task, nil
		}
	}
	return PlanningTask{}, fmt.Errorf("tâche introuvable")
}

func (a *App) generationStartDate(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		now := a.now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}
	return parseDate(value)
}

func (a *App) findAvailableSlot(tx *sql.Tx, cursor time.Time, duration int, slots []WorkdaySlot) (string, string, string, error) {
	for day := 0; day < 366; day++ {
		date := cursor.AddDate(0, 0, day).Format("2006-01-02")
		for _, slot := range slots {
			candidate := slot.Start
			for timeBefore(candidate, slot.End) {
				end := addMinutes(candidate, duration)
				if !timeBeforeOrEqual(end, slot.End) {
					break
				}
				var conflicts int
				if err := tx.QueryRow(`SELECT COUNT(*) FROM taches_planning WHERE date = ? AND heure_debut < ? AND heure_fin > ?`, date, end, candidate).Scan(&conflicts); err != nil {
					return "", "", "", err
				}
				if conflicts == 0 {
					return date, candidate, end, nil
				}
				candidate = addMinutes(candidate, 15)
			}
		}
	}
	return "", "", "", fmt.Errorf("aucun créneau libre dans les 12 prochains mois")
}

func validateSelection(selection PlanningSelection) error {
	if selection.ChapterID <= 0 || selection.RevisionCount < 1 || selection.DurationMinutes < 15 || selection.DurationMinutes > 720 || selection.SpacingDays < 0 {
		return fmt.Errorf("contraintes de planification invalides")
	}
	if len(selection.StartDays) == 0 {
		return fmt.Errorf("choisissez au moins un jour de début")
	}
	for _, day := range selection.StartDays {
		if day < 0 || day > 6 {
			return fmt.Errorf("jour de début invalide")
		}
	}
	return nil
}

func enabledSlots(slots []WorkdaySlot) []WorkdaySlot {
	active := []WorkdaySlot{}
	for _, slot := range slots {
		if slot.Enabled {
			active = append(active, slot)
		}
	}
	sort.Slice(active, func(i, j int) bool { return active[i].Start < active[j].Start })
	return active
}

func firstPreferredDate(start time.Time, days []int) time.Time {
	allowed := map[int]bool{}
	for _, day := range days {
		allowed[day] = true
	}
	for offset := 0; offset < 7; offset++ {
		candidate := start.AddDate(0, 0, offset)
		if allowed[int(candidate.Weekday())] {
			return candidate
		}
	}
	return start
}

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func mustDate(value string) time.Time {
	date, _ := parseDate(value)
	return date
}

func validateTimes(start, end string) error {
	if _, err := time.Parse("15:04", start); err != nil {
		return fmt.Errorf("heure de début invalide")
	}
	if _, err := time.Parse("15:04", end); err != nil {
		return fmt.Errorf("heure de fin invalide")
	}
	if !timeBefore(start, end) {
		return fmt.Errorf("l'heure de fin doit être après l'heure de début")
	}
	return nil
}

func addMinutes(value string, minutes int) string {
	parsed, _ := time.Parse("15:04", value)
	return parsed.Add(time.Duration(minutes) * time.Minute).Format("15:04")
}

func timeBefore(left, right string) bool        { return left < right }
func timeBeforeOrEqual(left, right string) bool { return left <= right }

func sameHeaders(actual, expected []string) bool {
	if len(actual) < len(expected) {
		return false
	}
	for index, header := range expected {
		if strings.ToLower(strings.TrimSpace(actual[index])) != header {
			return false
		}
	}
	return true
}

func paddedRow(row []string, length int) []string {
	values := make([]string, length)
	copy(values, row)
	return values
}

func emptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
