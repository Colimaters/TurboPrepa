package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "modernc.org/sqlite"
)

type App struct {
	ctx context.Context
	db  *sql.DB
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

var quotes = []Quote{
	{Text: "Le succès, c'est d'aller d'échec en échec sans perdre son enthousiasme.", Author: "Winston Churchill"},
	{Text: "La liberté est le droit de faire tout ce que les lois permettent.", Author: "Montesquieu", Source: "De l'esprit des lois"},
	{Text: "La justice sans la force est impuissante ; la force sans la justice est tyrannique.", Author: "Blaise Pascal", Source: "Pensées"},
	{Text: "Ce n'est pas parce que les choses sont difficiles que nous n'osons pas, c'est parce que nous n'osons pas qu'elles sont difficiles.", Author: "Sénèque", Source: "Lettres à Lucilius"},
	{Text: "Il n'est point de bonheur sans liberté, ni de liberté sans courage.", Author: "Périclès", UncertainAttribution: true},
	{Text: "La victoire appartient au plus opiniâtre.", Author: "Roland Garros"},
	{Text: "Le courage n'est pas l'absence de peur, mais la capacité de la vaincre.", Author: "Nelson Mandela", UncertainAttribution: true},
	{Text: "L'instruction est le premier besoin d'un peuple libre.", Author: "Napoléon Bonaparte", UncertainAttribution: true},
	{Text: "La discipline est la force principale des armées.", Author: "George Washington", UncertainAttribution: true},
	{Text: "Il faut cultiver notre jardin.", Author: "Voltaire", Source: "Candide"},
	{Text: "À vaincre sans péril, on triomphe sans gloire.", Author: "Pierre Corneille", Source: "Le Cid"},
	{Text: "La patience est amère, mais son fruit est doux.", Author: "Jean-Jacques Rousseau", UncertainAttribution: true},
	{Text: "Le bonheur est dans la liberté, et la liberté dans le courage.", Author: "Périclès", UncertainAttribution: true},
	{Text: "La connaissance s'acquiert par l'expérience, tout le reste n'est que de l'information.", Author: "Albert Einstein", UncertainAttribution: true},
	{Text: "Le droit est l'ensemble des conditions qui permettent à la liberté de chacun de s'accorder à la liberté de tous.", Author: "Emmanuel Kant", Source: "Doctrine du droit"},
	{Text: "La grandeur d'un métier est peut-être, avant tout, d'unir des hommes.", Author: "Antoine de Saint-Exupéry", Source: "Terre des hommes"},
	{Text: "La persévérance est la vertu par laquelle on poursuit le bien malgré les obstacles.", Author: "Thomas d'Aquin", UncertainAttribution: true},
	{Text: "Fais de ta vie un rêve, et d'un rêve, une réalité.", Author: "Antoine de Saint-Exupéry", UncertainAttribution: true},
	{Text: "Là où est la volonté, là est le chemin.", Author: "Proverbe anglais", Source: "Where there's a will, there's a way"},
	{Text: "Le travail éloigne de nous trois grands maux : l'ennui, le vice et le besoin.", Author: "Voltaire", Source: "Candide"},
	{Text: "La difficulté attire l'homme de caractère, car c'est en l'étreignant qu'il se réalise lui-même.", Author: "Charles de Gaulle", UncertainAttribution: true},
	{Text: "Un homme n'est jamais si grand que lorsqu'il est à genoux pour aider un enfant.", Author: "Pythagore", UncertainAttribution: true},
	{Text: "Il n'y a point de chemin trop long à qui marche lentement et sans se presser.", Author: "Jean de La Fontaine", Source: "Le Lièvre et la Tortue"},
	{Text: "On ne peut rien fonder sur la faiblesse.", Author: "Charles de Gaulle", Source: "Mémoires de guerre"},
	{Text: "La vraie générosité envers l'avenir consiste à tout donner au présent.", Author: "Albert Camus", Source: "L'Homme révolté"},
	{Text: "Il faut avoir beaucoup étudié pour savoir peu.", Author: "Montesquieu", UncertainAttribution: true},
	{Text: "La volonté de gagner ne signifie rien sans la volonté de se préparer.", Author: "Juma Ikangaa"},
	{Text: "Il est plus facile de faire son devoir que de le connaître.", Author: "Alexandre Dumas", UncertainAttribution: true},
	{Text: "Le succès n'est pas final, l'échec n'est pas fatal : c'est le courage de continuer qui compte.", Author: "Winston Churchill", UncertainAttribution: true},
	{Text: "Pour la patrie, l'honneur et le droit.", Author: "Gendarmerie nationale", Source: "Devise institutionnelle"},
}

var defaultSubjects = []string{
	"Droit pénal général", "Droit pénal spécial", "Procédure pénale", "Droit européen",
	"Droit constitutionnel", "Droit administratif", "Droit de la fonction publique",
	"Libertés publiques", "Anglais", "Dictée", "Culture générale",
	"Connaissance du monde contemporain", "Connaissance de l'institution policière",
	"Note de synthèse", "Cas pratique", "Annales", "Sport", "Lecture",
}

var pastelColors = []string{
	"#BFD7ED", "#C9C5EB", "#C8E6D1", "#F8D7B5", "#F5C6D6", "#D7E5C6",
	"#F4D6A5", "#BFE3E0", "#D8C3E8", "#F6CBC1", "#CCE0F4", "#D4D4D4",
	"#E7D7B8", "#C6D9D1", "#E2C9D4",
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.openDatabase(); err != nil {
		panic(fmt.Sprintf("impossible d'initialiser la base de données : %v", err))
	}
}

func (a *App) openDatabase() error {
	dataDir, err := applicationDataDirectory()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("création du répertoire de données : %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(dataDir, "data.db"))
	if err != nil {
		return fmt.Errorf("ouverture de la base de données : %w", err)
	}
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
	_, err := a.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY);
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
		INSERT OR IGNORE INTO schema_migrations(version) VALUES (1);
	`)
	if err != nil {
		return fmt.Errorf("migration du schéma : %w", err)
	}

	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for index, subject := range defaultSubjects {
		if _, err := tx.Exec(
			`INSERT INTO matieres(nom, couleur, ordre)
			 SELECT ?, ?, ? WHERE NOT EXISTS (SELECT 1 FROM matieres WHERE nom = ?)`,
			subject, pastelColors[index%len(pastelColors)], index, subject,
		); err != nil {
			return fmt.Errorf("initialisation des matières : %w", err)
		}
	}
	return tx.Commit()
}

func (a *App) GetDashboard() (Dashboard, error) {
	if a.db == nil {
		return Dashboard{}, fmt.Errorf("la base de données n'est pas initialisée")
	}
	today := time.Now().Format("2006-01-02")
	days := time.Now().Unix() / 86400
	quote := quotes[int(days%int64(len(quotes)))]
	dashboard := Dashboard{Quote: quote, Tasks: []TodayTask{}, Today: today}

	rows, err := a.db.Query(`
		SELECT statut, COUNT(*) FROM chapitres
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
		boolToInt(completed), taskID, time.Now().Format("2006-01-02"),
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

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
