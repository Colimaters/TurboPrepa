package application

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	tabProgramme       = "programme"
	tabFiches          = "fiches"
	tabCoursEntier     = "cours_entier"
	tabAnnales         = "annales"
	maxAttachmentBytes = 100 * 1024 * 1024
	maxArchiveBytes    = 500 * 1024 * 1024
	maxArchiveFiles    = 100
)

var (
	validTabs               = map[string]bool{tabProgramme: true, tabFiches: true, tabCoursEntier: true, tabAnnales: true}
	validStatuses           = map[string]bool{"a_planifier": true, "planifie": true, "en_cours": true, "maitrise": true}
	hexColor                = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	supportedFileExtensions = map[string]bool{
		".doc": true, ".docx": true, ".pages": true, ".pdf": true, ".jpg": true, ".jpeg": true,
		".png": true, ".xlsx": true, ".xls": true, ".numbers": true,
	}
)

type MatiereSummary struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Order    int    `json:"order"`
	Chapters int    `json:"chapters"`
	Mastered int    `json:"mastered"`
}

type Chapter struct {
	ID        int64        `json:"id"`
	MatiereID int64        `json:"matiereId"`
	Tab       string       `json:"tab"`
	Name      string       `json:"name"`
	Status    string       `json:"status"`
	Content   string       `json:"content"`
	Order     int          `json:"order"`
	Files     []Attachment `json:"files"`
}

type SubjectWork struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	DueDate   string `json:"dueDate"`
	Completed bool   `json:"completed"`
	Order     int    `json:"order"`
}

type Attachment struct {
	ID           int64  `json:"id"`
	ChapterID    int64  `json:"chapterId"`
	DisplayName  string `json:"displayName"`
	OriginalName string `json:"originalName"`
	MimeType     string `json:"mimeType"`
	Size         int64  `json:"size"`
	AddedAt      string `json:"addedAt"`
}

type MatiereDetail struct {
	Subject  MatiereSummary `json:"subject"`
	Chapters []Chapter      `json:"chapters"`
	Works    []SubjectWork  `json:"works"`
}

type ImportResult struct {
	Imported []Attachment `json:"imported"`
	Skipped  []string     `json:"skipped"`
}

func migrateMatieres(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		CREATE TABLE couleurs_pastel (
			id INTEGER PRIMARY KEY,
			valeur TEXT NOT NULL UNIQUE CHECK(valeur GLOB '#[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]'),
			ordre INTEGER NOT NULL UNIQUE
		);
		INSERT INTO couleurs_pastel(valeur, ordre) VALUES
			('#BFD7ED', 0), ('#C9C5EB', 1), ('#C8E6D1', 2), ('#F8D7B5', 3), ('#F5C6D6', 4),
			('#D7E5C6', 5), ('#F4D6A5', 6), ('#BFE3E0', 7), ('#D8C3E8', 8), ('#F6CBC1', 9),
			('#CCE0F4', 10), ('#D4D4D4', 11), ('#E7D7B8', 12), ('#C6D9D1', 13), ('#E2C9D4', 14);
		ALTER TABLE chapitres ADD COLUMN sous_onglet TEXT NOT NULL DEFAULT 'programme'
			CHECK(sous_onglet IN ('programme', 'fiches', 'cours_entier', 'annales'));
		ALTER TABLE chapitres ADD COLUMN contenu TEXT NOT NULL DEFAULT '';
		CREATE TABLE travaux_matiere (
			id INTEGER PRIMARY KEY,
			matiere_id INTEGER NOT NULL,
			titre TEXT NOT NULL CHECK(length(trim(titre)) > 0),
			date_echeance TEXT,
			termine INTEGER NOT NULL DEFAULT 0 CHECK(termine IN (0, 1)),
			ordre INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (matiere_id) REFERENCES matieres(id) ON DELETE CASCADE
		);
		CREATE TABLE fichiers (
			id INTEGER PRIMARY KEY,
			chapitre_id INTEGER,
			tache_id INTEGER,
			nom_affiche TEXT NOT NULL CHECK(length(trim(nom_affiche)) > 0),
			nom_original TEXT NOT NULL,
			chemin_disque TEXT NOT NULL UNIQUE,
			type_mime TEXT NOT NULL,
			taille INTEGER NOT NULL CHECK(taille >= 0),
			date_ajout TEXT NOT NULL,
			ordre INTEGER NOT NULL DEFAULT 0,
			CHECK((chapitre_id IS NOT NULL AND tache_id IS NULL) OR (chapitre_id IS NULL AND tache_id IS NOT NULL)),
			FOREIGN KEY (chapitre_id) REFERENCES chapitres(id) ON DELETE CASCADE,
			FOREIGN KEY (tache_id) REFERENCES taches_planning(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_chapitres_matiere_onglet_ordre ON chapitres(matiere_id, sous_onglet, ordre, id);
		CREATE INDEX idx_travaux_matiere_ordre ON travaux_matiere(matiere_id, ordre, id);
		CREATE INDEX idx_fichiers_chapitre_ordre ON fichiers(chapitre_id, ordre, id);`); err != nil {
		return err
	}
	return nil
}

func (a *App) ListMatieres() ([]MatiereSummary, error) {
	rows, err := a.db.Query(`
		SELECT m.id, m.nom, m.couleur, m.ordre, COUNT(c.id),
		       COALESCE(SUM(CASE WHEN c.statut = 'maitrise' THEN 1 ELSE 0 END), 0)
		FROM matieres m LEFT JOIN chapitres c ON c.matiere_id = m.id AND c.sous_onglet = 'programme'
		GROUP BY m.id ORDER BY m.ordre, m.id`)
	if err != nil {
		return nil, fmt.Errorf("lecture des matières : %w", err)
	}
	defer rows.Close()
	subjects := []MatiereSummary{}
	for rows.Next() {
		var subject MatiereSummary
		if err := rows.Scan(&subject.ID, &subject.Name, &subject.Color, &subject.Order, &subject.Chapters, &subject.Mastered); err != nil {
			return nil, fmt.Errorf("lecture d'une matière : %w", err)
		}
		subjects = append(subjects, subject)
	}
	return subjects, rows.Err()
}

func (a *App) ListPastelColors() ([]string, error) {
	rows, err := a.db.Query(`SELECT valeur FROM couleurs_pastel ORDER BY ordre`)
	if err != nil {
		return nil, fmt.Errorf("lecture des couleurs : %w", err)
	}
	defer rows.Close()
	colors := []string{}
	for rows.Next() {
		var color string
		if err := rows.Scan(&color); err != nil {
			return nil, err
		}
		colors = append(colors, color)
	}
	return colors, rows.Err()
}

func (a *App) GetMatiereDetail(subjectID int64) (MatiereDetail, error) {
	var detail MatiereDetail
	if err := a.db.QueryRow(`SELECT id, nom, couleur, ordre FROM matieres WHERE id = ?`, subjectID).
		Scan(&detail.Subject.ID, &detail.Subject.Name, &detail.Subject.Color, &detail.Subject.Order); err != nil {
		return detail, notFound("matière", err)
	}
	rows, err := a.db.Query(`SELECT id, matiere_id, sous_onglet, nom, statut, contenu, ordre FROM chapitres WHERE matiere_id = ? ORDER BY sous_onglet, ordre, id`, subjectID)
	if err != nil {
		return detail, fmt.Errorf("lecture des chapitres : %w", err)
	}
	detail.Chapters = []Chapter{}
	for rows.Next() {
		var chapter Chapter
		if err := rows.Scan(&chapter.ID, &chapter.MatiereID, &chapter.Tab, &chapter.Name, &chapter.Status, &chapter.Content, &chapter.Order); err != nil {
			rows.Close()
			return detail, err
		}
		detail.Chapters = append(detail.Chapters, chapter)
		if chapter.Tab == tabProgramme {
			detail.Subject.Chapters++
			if chapter.Status == "maitrise" {
				detail.Subject.Mastered++
			}
		}
	}
	if err := rows.Err(); err != nil {
		return detail, err
	}
	if err := rows.Close(); err != nil {
		return detail, err
	}
	for index := range detail.Chapters {
		files, err := a.listAttachments(detail.Chapters[index].ID)
		if err != nil {
			return detail, err
		}
		detail.Chapters[index].Files = files
	}
	workRows, err := a.db.Query(`SELECT id, titre, COALESCE(date_echeance, ''), termine, ordre FROM travaux_matiere WHERE matiere_id = ? ORDER BY ordre, id`, subjectID)
	if err != nil {
		return detail, fmt.Errorf("lecture des travaux : %w", err)
	}
	defer workRows.Close()
	detail.Works = []SubjectWork{}
	for workRows.Next() {
		var work SubjectWork
		var completed int
		if err := workRows.Scan(&work.ID, &work.Title, &work.DueDate, &completed, &work.Order); err != nil {
			return detail, err
		}
		work.Completed = completed == 1
		detail.Works = append(detail.Works, work)
	}
	return detail, workRows.Err()
}

func (a *App) CreateMatiere(name, color string) (MatiereSummary, error) {
	var subject MatiereSummary
	name = strings.TrimSpace(name)
	if name == "" {
		return subject, fmt.Errorf("le nom de la matière est obligatoire")
	}
	if !hexColor.MatchString(color) {
		return subject, fmt.Errorf("couleur invalide")
	}
	var colorExists bool
	if err := a.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM couleurs_pastel WHERE valeur = ?)`, color).Scan(&colorExists); err != nil || !colorExists {
		if err != nil {
			return subject, err
		}
		return subject, fmt.Errorf("la couleur doit appartenir à la palette")
	}
	err := a.db.QueryRow(`INSERT INTO matieres(nom, couleur, ordre) VALUES (?, ?, COALESCE((SELECT MAX(ordre) + 1 FROM matieres), 0)) RETURNING id, nom, couleur, ordre`, name, color).
		Scan(&subject.ID, &subject.Name, &subject.Color, &subject.Order)
	if err != nil {
		return subject, fmt.Errorf("création de la matière : %w", err)
	}
	return subject, nil
}

func (a *App) RenameMatiere(subjectID int64, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("le nom de la matière est obligatoire")
	}
	return a.execAffected("matière", `UPDATE matieres SET nom = ? WHERE id = ?`, name, subjectID)
}

func (a *App) DeleteMatiere(subjectID int64) error {
	return a.deleteWithAttachments(
		`SELECT f.chemin_disque FROM fichiers f JOIN chapitres c ON c.id = f.chapitre_id WHERE c.matiere_id = ?`,
		`DELETE FROM matieres WHERE id = ?`, "matière", subjectID,
	)
}

func (a *App) CreateChapter(subjectID int64, tab, name, content string) (Chapter, error) {
	var chapter Chapter
	name = strings.TrimSpace(name)
	if !validTabs[tab] {
		return chapter, fmt.Errorf("sous-onglet invalide")
	}
	if name == "" {
		return chapter, fmt.Errorf("le nom du dossier ou chapitre est obligatoire")
	}
	err := a.db.QueryRow(`INSERT INTO chapitres(matiere_id, sous_onglet, nom, statut, contenu, ordre) VALUES (?, ?, ?, 'a_planifier', ?, COALESCE((SELECT MAX(ordre) + 1 FROM chapitres WHERE matiere_id = ? AND sous_onglet = ?), 0)) RETURNING id, matiere_id, sous_onglet, nom, statut, contenu, ordre`, subjectID, tab, name, content, subjectID, tab).
		Scan(&chapter.ID, &chapter.MatiereID, &chapter.Tab, &chapter.Name, &chapter.Status, &chapter.Content, &chapter.Order)
	if err != nil {
		return chapter, fmt.Errorf("création du dossier ou chapitre : %w", err)
	}
	chapter.Files = []Attachment{}
	return chapter, nil
}

func (a *App) UpdateChapter(chapterID int64, name, content string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("le nom du dossier ou chapitre est obligatoire")
	}
	return a.execAffected("dossier ou chapitre", `UPDATE chapitres SET nom = ?, contenu = ? WHERE id = ?`, name, content, chapterID)
}

func (a *App) SetChapterStatus(chapterID int64, status string) error {
	if !validStatuses[status] {
		return fmt.Errorf("statut invalide")
	}
	return a.execAffected("chapitre du programme", `UPDATE chapitres SET statut = ? WHERE id = ? AND sous_onglet = 'programme'`, status, chapterID)
}

func (a *App) DeleteChapter(chapterID int64) error {
	return a.deleteWithAttachments(
		`SELECT chemin_disque FROM fichiers WHERE chapitre_id = ?`,
		`DELETE FROM chapitres WHERE id = ?`, "dossier ou chapitre", chapterID,
	)
}

func (a *App) CreateSubjectWork(subjectID int64, title, dueDate string) (SubjectWork, error) {
	var work SubjectWork
	title = strings.TrimSpace(title)
	if title == "" {
		return work, fmt.Errorf("l'intitulé du travail est obligatoire")
	}
	err := a.db.QueryRow(`INSERT INTO travaux_matiere(matiere_id, titre, date_echeance, ordre) VALUES (?, ?, NULLIF(?, ''), COALESCE((SELECT MAX(ordre) + 1 FROM travaux_matiere WHERE matiere_id = ?), 0)) RETURNING id, titre, COALESCE(date_echeance, ''), termine, ordre`, subjectID, title, dueDate, subjectID).
		Scan(&work.ID, &work.Title, &work.DueDate, new(int), &work.Order)
	if err != nil {
		return work, fmt.Errorf("création du travail : %w", err)
	}
	return work, nil
}

func (a *App) SetSubjectWorkCompleted(workID int64, completed bool) error {
	return a.execAffected("travail", `UPDATE travaux_matiere SET termine = ? WHERE id = ?`, boolToInt(completed), workID)
}

func (a *App) DeleteSubjectWork(workID int64) error {
	return a.execAffected("travail", `DELETE FROM travaux_matiere WHERE id = ?`, workID)
}

func (a *App) SelectAndImportChapterFiles(chapterID int64) (ImportResult, error) {
	if _, err := a.chapterSubjectID(chapterID); err != nil {
		return ImportResult{}, err
	}
	paths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{Title: "Ajouter des fichiers", Filters: []runtime.FileFilter{{DisplayName: "Documents pris en charge", Pattern: "*.doc;*.docx;*.pages;*.pdf;*.jpg;*.jpeg;*.png;*.xlsx;*.xls;*.numbers;*.zip"}}})
	if err != nil {
		return ImportResult{}, fmt.Errorf("sélection des fichiers : %w", err)
	}
	return a.ImportChapterFiles(chapterID, paths)
}

func (a *App) ImportChapterFiles(chapterID int64, paths []string) (ImportResult, error) {
	if _, err := a.chapterSubjectID(chapterID); err != nil {
		return ImportResult{}, err
	}
	result := ImportResult{Imported: []Attachment{}, Skipped: []string{}}
	for _, path := range paths {
		extension := strings.ToLower(filepath.Ext(path))
		switch extension {
		case ".zip":
			imported, skipped, err := a.importZip(chapterID, path)
			if err != nil {
				return result, err
			}
			result.Imported = append(result.Imported, imported...)
			result.Skipped = append(result.Skipped, skipped...)
		default:
			if !supportedFileExtensions[extension] {
				result.Skipped = append(result.Skipped, filepath.Base(path))
				continue
			}
			attachment, err := a.copyAttachment(chapterID, path, filepath.Base(path))
			if err != nil {
				return result, err
			}
			result.Imported = append(result.Imported, attachment)
		}
	}
	return result, nil
}

func (a *App) RenameAttachment(attachmentID int64, displayName string) error {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return fmt.Errorf("le nom affiché est obligatoire")
	}
	return a.execAffected("fichier", `UPDATE fichiers SET nom_affiche = ? WHERE id = ?`, displayName, attachmentID)
}

func (a *App) MoveAttachment(attachmentID, targetChapterID int64) error {
	targetSubjectID, err := a.chapterSubjectID(targetChapterID)
	if err != nil {
		return err
	}
	var sourceSubjectID int64
	if err := a.db.QueryRow(`SELECT c.matiere_id FROM fichiers f JOIN chapitres c ON c.id = f.chapitre_id WHERE f.id = ?`, attachmentID).Scan(&sourceSubjectID); err != nil {
		return notFound("fichier", err)
	}
	if sourceSubjectID != targetSubjectID {
		return fmt.Errorf("un fichier ne peut être déplacé que dans la même matière")
	}
	return a.execAffected("fichier", `UPDATE fichiers SET chapitre_id = ? WHERE id = ?`, targetChapterID, attachmentID)
}

func (a *App) DeleteAttachment(attachmentID int64) error {
	var path string
	if err := a.db.QueryRow(`SELECT chemin_disque FROM fichiers WHERE id = ?`, attachmentID).Scan(&path); err != nil {
		return notFound("fichier", err)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("suppression du fichier sur disque : %w", err)
	}
	return a.execAffected("fichier", `DELETE FROM fichiers WHERE id = ?`, attachmentID)
}

func (a *App) OpenAttachment(attachmentID int64) error {
	var path string
	if err := a.db.QueryRow(`SELECT chemin_disque FROM fichiers WHERE id = ?`, attachmentID).Scan(&path); err != nil {
		return notFound("fichier", err)
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("accès au fichier : %w", err)
	}
	runtime.BrowserOpenURL(a.ctx, (&url.URL{Scheme: "file", Path: path}).String())
	return nil
}

func (a *App) listAttachments(chapterID int64) ([]Attachment, error) {
	rows, err := a.db.Query(`SELECT id, chapitre_id, nom_affiche, nom_original, type_mime, taille, date_ajout FROM fichiers WHERE chapitre_id = ? ORDER BY ordre, id`, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	attachments := []Attachment{}
	for rows.Next() {
		var attachment Attachment
		if err := rows.Scan(&attachment.ID, &attachment.ChapterID, &attachment.DisplayName, &attachment.OriginalName, &attachment.MimeType, &attachment.Size, &attachment.AddedAt); err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}
	return attachments, rows.Err()
}

func (a *App) importZip(chapterID int64, archivePath string) ([]Attachment, []string, error) {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, nil, fmt.Errorf("ouverture de l'archive : %w", err)
	}
	defer archive.Close()
	imported, skipped := []Attachment{}, []string{}
	var totalUncompressed uint64
	for _, entry := range archive.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		if filepath.IsAbs(entry.Name) || strings.Contains(filepath.Clean(entry.Name), ".."+string(filepath.Separator)) {
			skipped = append(skipped, entry.Name)
			continue
		}
		if !supportedFileExtensions[strings.ToLower(filepath.Ext(entry.Name))] {
			skipped = append(skipped, entry.Name)
			continue
		}
		if len(imported) >= maxArchiveFiles {
			return imported, skipped, fmt.Errorf("archive trop volumineuse : maximum %d fichiers importables", maxArchiveFiles)
		}
		if entry.UncompressedSize64 > maxAttachmentBytes || totalUncompressed+entry.UncompressedSize64 > maxArchiveBytes {
			return imported, skipped, fmt.Errorf("archive trop volumineuse : un fichier dépasse la limite de 100 Mo ou le total dépasse 500 Mo")
		}
		reader, err := entry.Open()
		if err != nil {
			return imported, skipped, fmt.Errorf("lecture de %s : %w", entry.Name, err)
		}
		attachment, copyErr := a.copyAttachmentFromReader(chapterID, filepath.Base(entry.Name), entry.Name, reader)
		reader.Close()
		if copyErr != nil {
			return imported, skipped, copyErr
		}
		totalUncompressed += entry.UncompressedSize64
		imported = append(imported, attachment)
	}
	return imported, skipped, nil
}

func (a *App) copyAttachment(chapterID int64, sourcePath, displayName string) (Attachment, error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return Attachment{}, fmt.Errorf("ouverture du fichier : %w", err)
	}
	defer source.Close()
	return a.copyAttachmentFromReader(chapterID, filepath.Base(sourcePath), displayName, source)
}

func (a *App) copyAttachmentFromReader(chapterID int64, originalName, displayName string, source io.Reader) (Attachment, error) {
	directory := filepath.Join(a.dataDir, "fichiers")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return Attachment{}, fmt.Errorf("création du répertoire des fichiers : %w", err)
	}
	storedPath := filepath.Join(directory, uuid.NewString()+"-"+filepath.Base(originalName))
	destination, err := os.OpenFile(storedPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return Attachment{}, fmt.Errorf("création de la copie : %w", err)
	}
	size, copyErr := io.Copy(destination, io.LimitReader(source, maxAttachmentBytes+1))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil {
		os.Remove(storedPath)
		if copyErr != nil {
			return Attachment{}, fmt.Errorf("copie du fichier : %w", copyErr)
		}
		return Attachment{}, fmt.Errorf("finalisation de la copie : %w", closeErr)
	}
	if size > maxAttachmentBytes {
		os.Remove(storedPath)
		return Attachment{}, fmt.Errorf("fichier trop volumineux : maximum 100 Mo")
	}
	mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(originalName)))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	var attachment Attachment
	err = a.db.QueryRow(`INSERT INTO fichiers(chapitre_id, nom_affiche, nom_original, chemin_disque, type_mime, taille, date_ajout, ordre) VALUES (?, ?, ?, ?, ?, ?, ?, COALESCE((SELECT MAX(ordre) + 1 FROM fichiers WHERE chapitre_id = ?), 0)) RETURNING id, chapitre_id, nom_affiche, nom_original, type_mime, taille, date_ajout`, chapterID, displayName, originalName, storedPath, mimeType, size, time.Now().UTC().Format(time.RFC3339), chapterID).
		Scan(&attachment.ID, &attachment.ChapterID, &attachment.DisplayName, &attachment.OriginalName, &attachment.MimeType, &attachment.Size, &attachment.AddedAt)
	if err != nil {
		os.Remove(storedPath)
		return Attachment{}, fmt.Errorf("enregistrement du fichier : %w", err)
	}
	return attachment, nil
}

func (a *App) chapterSubjectID(chapterID int64) (int64, error) {
	var subjectID int64
	if err := a.db.QueryRow(`SELECT matiere_id FROM chapitres WHERE id = ?`, chapterID).Scan(&subjectID); err != nil {
		return 0, notFound("dossier ou chapitre", err)
	}
	return subjectID, nil
}

func (a *App) requireAffected(result sql.Result, err error, entity string) error {
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%s introuvable", entity)
	}
	return nil
}

func (a *App) execAffected(entity, query string, args ...any) error {
	result, err := a.db.Exec(query, args...)
	return a.requireAffected(result, err, entity)
}

func (a *App) deleteWithAttachments(pathsQuery, deleteQuery, entity string, id int64) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	rows, err := tx.Query(pathsQuery, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			rows.Close()
			tx.Rollback()
			return err
		}
		paths = append(paths, path)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		tx.Rollback()
		return err
	}
	if err := rows.Close(); err != nil {
		tx.Rollback()
		return err
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			tx.Rollback()
			return fmt.Errorf("suppression du fichier sur disque : %w", err)
		}
	}
	result, err := tx.Exec(deleteQuery, id)
	if err := a.requireAffected(result, err, entity); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func notFound(entity string, err error) error {
	if err == sql.ErrNoRows {
		return fmt.Errorf("%s introuvable", entity)
	}
	return err
}
