package application

import (
	"fmt"
	"testing"
	"time"
)

func addQuizQuestions(t *testing.T, app *App, count int) []QuizQuestion {
	t.Helper()
	questions := make([]QuizQuestion, 0, count)
	for index := 0; index < count; index++ {
		question, err := app.CreateQuizQuestion(QuizQuestionInput{
			Question:      fmt.Sprintf("Question %d", index+1),
			Choices:       []string{"Vrai", "Faux"},
			CorrectAnswer: index % 2,
			Theme:         "Droit public",
			Explanation:   "Explication",
		})
		if err != nil {
			t.Fatalf("CreateQuizQuestion(%d) error = %v", index, err)
		}
		questions = append(questions, question)
	}
	return questions
}

func TestQuizRequiresFiveQuestionsAndValidatesQuestionInput(t *testing.T) {
	app := newTestApp(t, time.Date(2026, time.August, 16, 9, 0, 0, 0, time.Local))
	if _, err := app.GetDailyQuiz(); err == nil {
		t.Error("GetDailyQuiz() error = nil, want insufficient question stock error")
	}
	if _, err := app.CreateQuizQuestion(QuizQuestionInput{Question: "Question", Theme: "Thème", Choices: []string{"Oui"}, CorrectAnswer: 0}); err == nil {
		t.Error("CreateQuizQuestion() error = nil, want invalid choice count")
	}
	questions := addQuizQuestions(t, app, 5)
	if err := app.UpdateQuizQuestion(questions[0].ID, QuizQuestionInput{Question: "Modifiée", Theme: "Thème", Choices: []string{"A", "B"}, CorrectAnswer: 1}); err != nil {
		t.Fatalf("UpdateQuizQuestion() error = %v", err)
	}
	list, err := app.ListQuizQuestions()
	if err != nil {
		t.Fatalf("ListQuizQuestions() error = %v", err)
	}
	foundUpdated := false
	for _, question := range list {
		foundUpdated = foundUpdated || question.Question == "Modifiée"
	}
	if len(list) != 5 || !foundUpdated {
		t.Errorf("questions = %#v, want five including the update", list)
	}
}

func TestDailyQuizRotatesWithoutRepeatBeforeExhaustion(t *testing.T) {
	now := time.Date(2026, time.August, 16, 9, 0, 0, 0, time.Local)
	app := newTestApp(t, now)
	addQuizQuestions(t, app, 10)

	first, err := app.GetDailyQuiz()
	if err != nil {
		t.Fatalf("first GetDailyQuiz() error = %v", err)
	}
	if len(first.Questions) != dailyQuizQuestionCount {
		t.Fatalf("first questions = %d, want %d", len(first.Questions), dailyQuizQuestionCount)
	}
	firstIDs := map[int64]bool{}
	for _, question := range first.Questions {
		firstIDs[question.ID] = true
	}
	app.now = func() time.Time { return now.AddDate(0, 0, 1) }
	second, err := app.GetDailyQuiz()
	if err != nil {
		t.Fatalf("second GetDailyQuiz() error = %v", err)
	}
	for _, question := range second.Questions {
		if firstIDs[question.ID] {
			t.Errorf("question %d repeated before all ten questions were used", question.ID)
		}
	}
	app.now = func() time.Time { return now }
	reloaded, err := app.GetDailyQuiz()
	if err != nil {
		t.Fatalf("reloaded GetDailyQuiz() error = %v", err)
	}
	for index, question := range reloaded.Questions {
		if question.ID != first.Questions[index].ID {
			t.Errorf("daily series changed at %d: got %d want %d", index, question.ID, first.Questions[index].ID)
		}
	}
}

func TestQuizSubmissionCorrectionAndStreak(t *testing.T) {
	now := time.Date(2026, time.August, 16, 9, 0, 0, 0, time.Local)
	app := newTestApp(t, now)
	addQuizQuestions(t, app, 10)
	questions, err := app.ListQuizQuestions()
	if err != nil {
		t.Fatalf("ListQuizQuestions() error = %v", err)
	}
	byID := map[int64]QuizQuestion{}
	for _, question := range questions {
		byID[question.ID] = question
	}
	daily, err := app.GetDailyQuiz()
	if err != nil {
		t.Fatalf("GetDailyQuiz() error = %v", err)
	}
	answers := []int{byID[daily.Questions[0].ID].CorrectAnswer, -1, byID[daily.Questions[2].ID].CorrectAnswer, -1, byID[daily.Questions[4].ID].CorrectAnswer}
	result, err := app.SubmitDailyQuiz(answers, false)
	if err != nil {
		t.Fatalf("SubmitDailyQuiz() error = %v", err)
	}
	if result.Total != 5 || result.Score != 3 || len(result.Corrections) != 5 {
		t.Errorf("result = %#v, want a 3/5 correction", result)
	}
	if _, err := app.SubmitDailyQuiz(answers, false); err == nil {
		t.Error("second SubmitDailyQuiz() error = nil, want already completed error")
	}
	daily, err = app.GetDailyQuiz()
	if err != nil {
		t.Fatalf("GetDailyQuiz() after submission error = %v", err)
	}
	if !daily.Completed || daily.Result == nil || daily.Result.Score != 3 {
		t.Errorf("daily = %#v, want persisted completion", daily)
	}
	progress, err := app.GetQuizProgress()
	if err != nil {
		t.Fatalf("GetQuizProgress() error = %v", err)
	}
	if progress.Streak != 1 || progress.TotalScore != 3 || len(progress.History) != 1 {
		t.Errorf("progress = %#v, want one-day streak and 3 total points", progress)
	}

	app.now = func() time.Time { return now.AddDate(0, 0, 2) }
	progress, err = app.GetQuizProgress()
	if err != nil {
		t.Fatalf("GetQuizProgress() after missed day error = %v", err)
	}
	if progress.Streak != 0 {
		t.Errorf("streak = %d, want reset after missed day", progress.Streak)
	}
}

func TestQuizExpirationPersistsUnansweredResponses(t *testing.T) {
	now := time.Date(2026, time.August, 16, 9, 0, 0, 0, time.Local)
	app := newTestApp(t, now)
	addQuizQuestions(t, app, 5)
	if _, err := app.GetDailyQuiz(); err != nil {
		t.Fatalf("GetDailyQuiz() error = %v", err)
	}
	app.now = func() time.Time { return now.Add(2 * time.Minute) }
	result, err := app.SubmitDailyQuiz([]int{-1, -1, -1, -1, -1}, false)
	if err != nil {
		t.Fatalf("SubmitDailyQuiz() error = %v", err)
	}
	if !result.Expired || result.Score != 0 {
		t.Errorf("result = %#v, want expired zero-score quiz", result)
	}
	progress, err := app.GetQuizProgress()
	if err != nil {
		t.Fatalf("GetQuizProgress() error = %v", err)
	}
	if len(progress.History) != 1 || !progress.History[0].Expired {
		t.Errorf("history = %#v, want expired entry", progress.History)
	}
}
