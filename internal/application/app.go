package application

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "modernc.org/sqlite"
)

type App struct {
	ctx     context.Context
	db      *sql.DB
	dataDir string
	now     func() time.Time
	quotes  []Quote
}

type Quote struct {
	Text                 string `json:"text"`
	Author               string `json:"author"`
	Source               string `json:"source,omitempty"`
	UncertainAttribution bool   `json:"uncertainAttribution"`
}

type Progress struct {
	ToPlan     int `json:"toPlan"`
	Planned    int `json:"planned"`
	InProgress int `json:"inProgress"`
	Mastered   int `json:"mastered"`
}

type TodayTask struct {
	ID        int64  `json:"id"`
	Subject   string `json:"subject"`
	Color     string `json:"color"`
	Title     string `json:"title"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Completed bool   `json:"completed"`
}

type Dashboard struct {
	Quote    Quote       `json:"quote"`
	Progress Progress    `json:"progress"`
	Tasks    []TodayTask `json:"tasks"`
	Today    string      `json:"today"`
}

//go:embed assets/citations.json
var quotesJSON []byte

func NewApp() *App {
	return &App{now: time.Now}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.loadQuotes(); err != nil {
		panic(fmt.Sprintf("impossible de charger les citations : %v", err))
	}
	if err := a.openDatabase(); err != nil {
		panic(fmt.Sprintf("impossible d'initialiser la base de données : %v", err))
	}
}

func (a *App) openDatabase() error {
	if a.dataDir == "" {
		dataDir, err := applicationDataDirectory()
		if err != nil {
			return err
		}
		a.dataDir = dataDir
	}
	if err := os.MkdirAll(a.dataDir, 0o755); err != nil {
		return fmt.Errorf("création du répertoire de données : %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(a.dataDir, "data.db"))
	if err != nil {
		return fmt.Errorf("ouverture de la base de données : %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err = db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return fmt.Errorf("activation des clés étrangères : %w", err)
	}
	a.db = db
	return a.migrate()
}

func applicationDataDirectory() (string, error) {
	if runtime.GOOS == "darwin" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "TurboPrepa"), nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "TurboPrepa"), nil
}

func (a *App) migrate() error {
	if _, err := a.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)`); err != nil {
		return fmt.Errorf("création du suivi des migrations : %w", err)
	}

	for _, migration := range migrations {
		var applied bool
		if err := a.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)`, migration.version).Scan(&applied); err != nil {
			return fmt.Errorf("lecture des migrations : %w", err)
		}
		if applied {
			continue
		}

		tx, err := a.db.Begin()
		if err != nil {
			return fmt.Errorf("démarrage de la migration %d : %w", migration.version, err)
		}
		if err := migration.apply(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d : %w", migration.version, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, migration.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("enregistrement de la migration %d : %w", migration.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("validation de la migration %d : %w", migration.version, err)
		}
	}
	return nil
}

type migration struct {
	version int
	apply   func(*sql.Tx) error
}

var migrations = []migration{
	{version: 1, apply: migrateInitialSchema},
	{version: 2, apply: migratePlanningTaskDetails},
	{version: 3, apply: migrateMatieres},
}

func migrateInitialSchema(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS matieres (
			id INTEGER PRIMARY KEY, nom TEXT NOT NULL, couleur TEXT NOT NULL, ordre INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS chapitres (
			id INTEGER PRIMARY KEY, matiere_id INTEGER NOT NULL, nom TEXT NOT NULL,
			statut TEXT NOT NULL CHECK(statut IN ('a_planifier', 'planifie', 'en_cours', 'maitrise')),
			ordre INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (matiere_id) REFERENCES matieres(id) ON DELETE CASCADE
		);
		CREATE TABLE IF NOT EXISTS taches_planning (
			id INTEGER PRIMARY KEY, matiere_id INTEGER, titre TEXT NOT NULL, date TEXT NOT NULL,
			heure_debut TEXT NOT NULL, heure_fin TEXT NOT NULL, couleur TEXT, terminee INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (matiere_id) REFERENCES matieres(id) ON DELETE SET NULL
		);
	`); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO matieres(nom, couleur, ordre)
		VALUES
			('Droit pénal général', '#BFD7ED', 0), ('Droit pénal spécial', '#C9C5EB', 1),
			('Procédure pénale', '#C8E6D1', 2), ('Droit européen', '#F8D7B5', 3),
			('Droit constitutionnel', '#F5C6D6', 4), ('Droit administratif', '#D7E5C6', 5),
			('Droit de la fonction publique', '#F4D6A5', 6), ('Libertés publiques', '#BFE3E0', 7),
			('Anglais', '#D8C3E8', 8), ('Dictée', '#F6CBC1', 9),
			('Culture générale', '#CCE0F4', 10), ('Connaissance du monde contemporain', '#D4D4D4', 11),
			('Connaissance de l''institution policière', '#E7D7B8', 12), ('Note de synthèse', '#C6D9D1', 13),
			('Cas pratique', '#E2C9D4', 14), ('Annales', '#BFD7ED', 15),
			('Sport', '#C9C5EB', 16), ('Lecture', '#C8E6D1', 17)
		ON CONFLICT DO NOTHING`); err != nil {
		return fmt.Errorf("initialisation des matières : %w", err)
	}
	return nil
}

func migratePlanningTaskDetails(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		ALTER TABLE taches_planning ADD COLUMN chapitre_id INTEGER REFERENCES chapitres(id) ON DELETE SET NULL;
		ALTER TABLE taches_planning ADD COLUMN notes TEXT NOT NULL DEFAULT '';
	`); err != nil {
		return err
	}
	return nil
}

func (a *App) loadQuotes() error {
	var quotes []Quote
	if err := json.Unmarshal(quotesJSON, &quotes); err != nil {
		return fmt.Errorf("lecture du JSON : %w", err)
	}
	if len(quotes) == 0 {
		return fmt.Errorf("le stock est vide")
	}
	for index, quote := range quotes {
		if quote.Text == "" || quote.Author == "" {
			return fmt.Errorf("citation %d incomplète", index+1)
		}
	}
	a.quotes = quotes
	return nil
}

func (a *App) GetDashboard() (Dashboard, error) {
	if a.db == nil {
		return Dashboard{}, fmt.Errorf("la base de données n'est pas initialisée")
	}
	if len(a.quotes) == 0 {
		return Dashboard{}, fmt.Errorf("les citations ne sont pas initialisées")
	}
	now := a.now()
	today := now.Format("2006-01-02")
	days := civilDayIndex(now)
	quote := a.quotes[int(days%int64(len(a.quotes)))]
	dashboard := Dashboard{Quote: quote, Tasks: []TodayTask{}, Today: today}

	rows, err := a.db.Query(`
		SELECT statut, COUNT(*) FROM chapitres
		WHERE sous_onglet = 'programme'
		GROUP BY statut`)
	if err != nil {
		return Dashboard{}, fmt.Errorf("lecture de la progression : %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return Dashboard{}, fmt.Errorf("lecture d'un statut de chapitre : %w", err)
		}
		switch status {
		case "a_planifier":
			dashboard.Progress.ToPlan = count
		case "planifie":
			dashboard.Progress.Planned = count
		case "en_cours":
			dashboard.Progress.InProgress = count
		case "maitrise":
			dashboard.Progress.Mastered = count
		}
	}
	if err := rows.Err(); err != nil {
		return Dashboard{}, fmt.Errorf("parcours de la progression : %w", err)
	}

	taskRows, err := a.db.Query(`
		SELECT t.id, COALESCE(m.nom, 'Sans matière'), COALESCE(t.couleur, m.couleur, '#D4D4D4'),
		       t.titre, t.heure_debut, t.heure_fin, t.terminee
		FROM taches_planning t
		LEFT JOIN matieres m ON m.id = t.matiere_id
		WHERE t.date = ?
		ORDER BY t.heure_debut, t.heure_fin, t.id`, today)
	if err != nil {
		return Dashboard{}, fmt.Errorf("lecture des tâches du jour : %w", err)
	}
	defer taskRows.Close()
	for taskRows.Next() {
		var task TodayTask
		var completed int
		if err := taskRows.Scan(&task.ID, &task.Subject, &task.Color, &task.Title, &task.StartTime, &task.EndTime, &completed); err != nil {
			return Dashboard{}, fmt.Errorf("lecture d'une tâche : %w", err)
		}
		task.Completed = completed == 1
		dashboard.Tasks = append(dashboard.Tasks, task)
	}
	if err := taskRows.Err(); err != nil {
		return Dashboard{}, fmt.Errorf("parcours des tâches du jour : %w", err)
	}
	return dashboard, nil
}

func (a *App) ToggleTodayTask(taskID int64, completed bool) error {
	if a.db == nil {
		return fmt.Errorf("la base de données n'est pas initialisée")
	}
	result, err := a.db.Exec(
		`UPDATE taches_planning SET terminee = ? WHERE id = ? AND date = ?`,
		boolToInt(completed), taskID, a.now().Format("2006-01-02"),
	)
	if err != nil {
		return fmt.Errorf("mise à jour de la tâche : %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification de la tâche : %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("tâche du jour introuvable")
	}
	return nil
}

func civilDayIndex(value time.Time) int64 {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix() / 86400
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
