package application

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const dailyQuizQuestionCount = 5

type QuizQuestion struct {
	ID            int64    `json:"id"`
	Question      string   `json:"question"`
	Choices       []string `json:"choices"`
	CorrectAnswer int      `json:"correctAnswer"`
	Theme         string   `json:"theme"`
	Explanation   string   `json:"explanation"`
}

type QuizQuestionInput struct {
	Question      string   `json:"question"`
	Choices       []string `json:"choices"`
	CorrectAnswer int      `json:"correctAnswer"`
	Theme         string   `json:"theme"`
	Explanation   string   `json:"explanation"`
}

type DailyQuizQuestion struct {
	ID          int64    `json:"id"`
	Question    string   `json:"question"`
	Choices     []string `json:"choices"`
	Theme       string   `json:"theme"`
	Explanation string   `json:"explanation,omitempty"`
}

type QuizCorrection struct {
	QuestionID    int64  `json:"questionId"`
	Answer        int    `json:"answer"`
	CorrectAnswer int    `json:"correctAnswer"`
	Correct       bool   `json:"correct"`
	Explanation   string `json:"explanation,omitempty"`
}

type QuizResult struct {
	Score       int              `json:"score"`
	Total       int              `json:"total"`
	Expired     bool             `json:"expired"`
	Corrections []QuizCorrection `json:"corrections"`
}

type DailyQuiz struct {
	Date      string              `json:"date"`
	Questions []DailyQuizQuestion `json:"questions"`
	StartedAt string              `json:"startedAt"`
	Completed bool                `json:"completed"`
	Result    *QuizResult         `json:"result,omitempty"`
}

type QuizHistoryEntry struct {
	Date    string `json:"date"`
	Score   int    `json:"score"`
	Total   int    `json:"total"`
	Expired bool   `json:"expired"`
}

type QuizProgress struct {
	Streak     int                `json:"streak"`
	TotalScore int                `json:"totalScore"`
	History    []QuizHistoryEntry `json:"history"`
}

func migrateQuiz(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE quiz_questions (
			id INTEGER PRIMARY KEY,
			question TEXT NOT NULL,
			choix_json TEXT NOT NULL,
			bonne_reponse INTEGER NOT NULL,
			theme TEXT NOT NULL,
			explication TEXT NOT NULL DEFAULT ''
		);
		CREATE TABLE quiz_sessions (
			date TEXT PRIMARY KEY,
			questions_json TEXT NOT NULL,
			demarre_le TEXT NOT NULL
		);
		CREATE TABLE quiz_historique (
			id INTEGER PRIMARY KEY,
			date TEXT NOT NULL UNIQUE,
			score INTEGER NOT NULL,
			total INTEGER NOT NULL,
			reponses_json TEXT NOT NULL,
			expire INTEGER NOT NULL DEFAULT 0 CHECK(expire IN (0, 1)),
			termine_le TEXT NOT NULL
		);
		CREATE INDEX idx_quiz_historique_date ON quiz_historique(date DESC);
	`)
	return err
}

func (a *App) ListQuizQuestions() ([]QuizQuestion, error) {
	rows, err := a.db.Query(`SELECT id, question, choix_json, bonne_reponse, theme, explication FROM quiz_questions ORDER BY theme, id`)
	if err != nil {
		return nil, fmt.Errorf("lecture des questions : %w", err)
	}
	defer rows.Close()

	questions := []QuizQuestion{}
	for rows.Next() {
		question, err := scanQuizQuestion(rows)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours des questions : %w", err)
	}
	return questions, nil
}

func (a *App) CreateQuizQuestion(input QuizQuestionInput) (QuizQuestion, error) {
	if err := validateQuizQuestion(input); err != nil {
		return QuizQuestion{}, err
	}
	choices, err := json.Marshal(input.Choices)
	if err != nil {
		return QuizQuestion{}, fmt.Errorf("encodage des choix : %w", err)
	}
	result, err := a.db.Exec(`INSERT INTO quiz_questions(question, choix_json, bonne_reponse, theme, explication) VALUES (?, ?, ?, ?, ?)`,
		strings.TrimSpace(input.Question), choices, input.CorrectAnswer, strings.TrimSpace(input.Theme), strings.TrimSpace(input.Explanation))
	if err != nil {
		return QuizQuestion{}, fmt.Errorf("création de la question : %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return QuizQuestion{}, fmt.Errorf("lecture de la question créée : %w", err)
	}
	return QuizQuestion{ID: id, Question: strings.TrimSpace(input.Question), Choices: input.Choices, CorrectAnswer: input.CorrectAnswer, Theme: strings.TrimSpace(input.Theme), Explanation: strings.TrimSpace(input.Explanation)}, nil
}

func (a *App) UpdateQuizQuestion(id int64, input QuizQuestionInput) error {
	if id < 1 {
		return fmt.Errorf("question invalide")
	}
	if err := validateQuizQuestion(input); err != nil {
		return err
	}
	choices, err := json.Marshal(input.Choices)
	if err != nil {
		return fmt.Errorf("encodage des choix : %w", err)
	}
	result, err := a.db.Exec(`UPDATE quiz_questions SET question = ?, choix_json = ?, bonne_reponse = ?, theme = ?, explication = ? WHERE id = ?`,
		strings.TrimSpace(input.Question), choices, input.CorrectAnswer, strings.TrimSpace(input.Theme), strings.TrimSpace(input.Explanation), id)
	if err != nil {
		return fmt.Errorf("mise à jour de la question : %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification de la question : %w", err)
	}
	if count == 0 {
		return fmt.Errorf("question introuvable")
	}
	return nil
}

func (a *App) DeleteQuizQuestion(id int64) error {
	if id < 1 {
		return fmt.Errorf("question invalide")
	}
	result, err := a.db.Exec(`DELETE FROM quiz_questions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("suppression de la question : %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification de la question : %w", err)
	}
	if count == 0 {
		return fmt.Errorf("question introuvable")
	}
	return nil
}

func (a *App) GetDailyQuiz() (DailyQuiz, error) {
	today := a.now().Format("2006-01-02")
	ids, startedAt, err := a.dailyQuizQuestionIDs(today)
	if err != nil {
		return DailyQuiz{}, err
	}
	questions, err := a.loadDailyQuizQuestions(ids)
	if err != nil {
		return DailyQuiz{}, err
	}
	quiz := DailyQuiz{Date: today, Questions: questions, StartedAt: startedAt}

	var result QuizResult
	var answersJSON string
	var expired int
	err = a.db.QueryRow(`SELECT score, total, reponses_json, expire FROM quiz_historique WHERE date = ?`, today).Scan(&result.Score, &result.Total, &answersJSON, &expired)
	if err == sql.ErrNoRows {
		return quiz, nil
	}
	if err != nil {
		return DailyQuiz{}, fmt.Errorf("lecture du résultat du jour : %w", err)
	}
	if err := json.Unmarshal([]byte(answersJSON), &result.Corrections); err != nil {
		return DailyQuiz{}, fmt.Errorf("lecture des corrections du jour : %w", err)
	}
	result.Expired = expired == 1
	quiz.Completed = true
	quiz.Result = &result
	return quiz, nil
}

func (a *App) SubmitDailyQuiz(answers []int, expired bool) (QuizResult, error) {
	today := a.now().Format("2006-01-02")
	ids, startedAt, err := a.dailyQuizQuestionIDs(today)
	if err != nil {
		return QuizResult{}, err
	}
	if len(answers) != len(ids) {
		return QuizResult{}, fmt.Errorf("les réponses doivent correspondre aux %d questions du quiz", len(ids))
	}
	var completed bool
	if err := a.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM quiz_historique WHERE date = ?)`, today).Scan(&completed); err != nil {
		return QuizResult{}, fmt.Errorf("lecture du quiz du jour : %w", err)
	}
	if completed {
		return QuizResult{}, fmt.Errorf("le quiz du jour est déjà terminé")
	}
	started, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return QuizResult{}, fmt.Errorf("heure de démarrage du quiz invalide : %w", err)
	}
	if !a.now().Before(started.Add(2 * time.Minute)) {
		expired = true
	}

	questions, err := a.loadQuizQuestions(ids)
	if err != nil {
		return QuizResult{}, err
	}
	result := QuizResult{Total: len(questions), Expired: expired, Corrections: make([]QuizCorrection, 0, len(questions))}
	for index, question := range questions {
		answer := answers[index]
		if answer < -1 || answer >= len(question.Choices) {
			return QuizResult{}, fmt.Errorf("réponse invalide pour la question %d", index+1)
		}
		correction := QuizCorrection{QuestionID: question.ID, Answer: answer, CorrectAnswer: question.CorrectAnswer, Correct: answer == question.CorrectAnswer, Explanation: question.Explanation}
		if correction.Correct {
			result.Score++
		}
		result.Corrections = append(result.Corrections, correction)
	}
	corrections, err := json.Marshal(result.Corrections)
	if err != nil {
		return QuizResult{}, fmt.Errorf("encodage des corrections : %w", err)
	}
	if _, err := a.db.Exec(`INSERT INTO quiz_historique(date, score, total, reponses_json, expire, termine_le) VALUES (?, ?, ?, ?, ?, ?)`,
		today, result.Score, result.Total, corrections, boolToInt(expired), a.now().Format(time.RFC3339)); err != nil {
		return QuizResult{}, fmt.Errorf("enregistrement du résultat : %w", err)
	}
	return result, nil
}

func (a *App) GetQuizProgress() (QuizProgress, error) {
	progress := QuizProgress{History: []QuizHistoryEntry{}}
	if err := a.db.QueryRow(`SELECT COALESCE(SUM(score), 0) FROM quiz_historique`).Scan(&progress.TotalScore); err != nil {
		return QuizProgress{}, fmt.Errorf("lecture du score cumulé : %w", err)
	}
	rows, err := a.db.Query(`SELECT date, score, total, expire FROM quiz_historique ORDER BY date DESC LIMIT 30`)
	if err != nil {
		return QuizProgress{}, fmt.Errorf("lecture de l'historique : %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var entry QuizHistoryEntry
		var expired int
		if err := rows.Scan(&entry.Date, &entry.Score, &entry.Total, &expired); err != nil {
			return QuizProgress{}, fmt.Errorf("lecture d'un historique : %w", err)
		}
		entry.Expired = expired == 1
		progress.History = append(progress.History, entry)
	}
	if err := rows.Err(); err != nil {
		return QuizProgress{}, fmt.Errorf("parcours de l'historique : %w", err)
	}

	expected := a.now().Format("2006-01-02")
	for _, entry := range progress.History {
		if entry.Date != expected {
			break
		}
		progress.Streak++
		day, err := time.Parse("2006-01-02", expected)
		if err != nil {
			return QuizProgress{}, fmt.Errorf("date d'historique invalide : %w", err)
		}
		expected = day.AddDate(0, 0, -1).Format("2006-01-02")
	}
	return progress, nil
}

func (a *App) dailyQuizQuestionIDs(date string) ([]int64, string, error) {
	var sessionJSON, startedAt string
	err := a.db.QueryRow(`SELECT questions_json, demarre_le FROM quiz_sessions WHERE date = ?`, date).Scan(&sessionJSON, &startedAt)
	if err == nil {
		var ids []int64
		if err := json.Unmarshal([]byte(sessionJSON), &ids); err != nil {
			return nil, "", fmt.Errorf("lecture de la série du jour : %w", err)
		}
		return ids, startedAt, nil
	}
	if err != sql.ErrNoRows {
		return nil, "", fmt.Errorf("lecture de la série du jour : %w", err)
	}

	rows, err := a.db.Query(`SELECT id FROM quiz_questions ORDER BY id`)
	if err != nil {
		return nil, "", fmt.Errorf("lecture du stock de questions : %w", err)
	}
	defer rows.Close()
	var allIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, "", fmt.Errorf("lecture d'une question : %w", err)
		}
		allIDs = append(allIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("parcours du stock de questions : %w", err)
	}
	if len(allIDs) < dailyQuizQuestionCount {
		return nil, "", fmt.Errorf("ajoutez au moins %d questions pour lancer le quiz", dailyQuizQuestionCount)
	}
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, "", fmt.Errorf("date du quiz invalide : %w", err)
	}
	start := int((civilDayIndex(parsedDate) * dailyQuizQuestionCount) % int64(len(allIDs)))
	ids := make([]int64, dailyQuizQuestionCount)
	for index := range ids {
		ids[index] = allIDs[(start+index)%len(allIDs)]
	}
	serialized, err := json.Marshal(ids)
	if err != nil {
		return nil, "", fmt.Errorf("encodage de la série du jour : %w", err)
	}
	startedAt = a.now().Format(time.RFC3339)
	if _, err := a.db.Exec(`INSERT INTO quiz_sessions(date, questions_json, demarre_le) VALUES (?, ?, ?)`, date, serialized, startedAt); err != nil {
		return nil, "", fmt.Errorf("création de la série du jour : %w", err)
	}
	return ids, startedAt, nil
}

func (a *App) loadDailyQuizQuestions(ids []int64) ([]DailyQuizQuestion, error) {
	questions, err := a.loadQuizQuestions(ids)
	if err != nil {
		return nil, err
	}
	daily := make([]DailyQuizQuestion, len(questions))
	for index, question := range questions {
		daily[index] = DailyQuizQuestion{ID: question.ID, Question: question.Question, Choices: question.Choices, Theme: question.Theme}
	}
	return daily, nil
}

func (a *App) loadQuizQuestions(ids []int64) ([]QuizQuestion, error) {
	questions := make([]QuizQuestion, 0, len(ids))
	for _, id := range ids {
		row := a.db.QueryRow(`SELECT id, question, choix_json, bonne_reponse, theme, explication FROM quiz_questions WHERE id = ?`, id)
		question, err := scanQuizQuestion(row)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("une question de la série du jour a été supprimée")
		}
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}
	return questions, nil
}

type quizQuestionScanner interface {
	Scan(dest ...any) error
}

func scanQuizQuestion(scanner quizQuestionScanner) (QuizQuestion, error) {
	var question QuizQuestion
	var choicesJSON string
	if err := scanner.Scan(&question.ID, &question.Question, &choicesJSON, &question.CorrectAnswer, &question.Theme, &question.Explanation); err != nil {
		return QuizQuestion{}, err
	}
	if err := json.Unmarshal([]byte(choicesJSON), &question.Choices); err != nil {
		return QuizQuestion{}, fmt.Errorf("lecture des choix de la question %d : %w", question.ID, err)
	}
	return question, nil
}

func validateQuizQuestion(input QuizQuestionInput) error {
	if strings.TrimSpace(input.Question) == "" {
		return fmt.Errorf("l'énoncé est obligatoire")
	}
	if strings.TrimSpace(input.Theme) == "" {
		return fmt.Errorf("le thème est obligatoire")
	}
	if len(input.Choices) < 2 {
		return fmt.Errorf("au moins deux choix sont requis")
	}
	for _, choice := range input.Choices {
		if strings.TrimSpace(choice) == "" {
			return fmt.Errorf("les choix ne peuvent pas être vides")
		}
	}
	if input.CorrectAnswer < 0 || input.CorrectAnswer >= len(input.Choices) {
		return fmt.Errorf("la bonne réponse est invalide")
	}
	return nil
}
