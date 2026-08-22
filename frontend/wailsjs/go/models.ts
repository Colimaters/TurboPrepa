export namespace application {
	
	export class Attachment {
	    id: number;
	    chapterId: number;
	    displayName: string;
	    originalName: string;
	    mimeType: string;
	    size: number;
	    addedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Attachment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.chapterId = source["chapterId"];
	        this.displayName = source["displayName"];
	        this.originalName = source["originalName"];
	        this.mimeType = source["mimeType"];
	        this.size = source["size"];
	        this.addedAt = source["addedAt"];
	    }
	}
	export class Chapter {
	    id: number;
	    matiereId: number;
	    tab: string;
	    name: string;
	    status: string;
	    content: string;
	    order: number;
	    files: Attachment[];
	
	    static createFrom(source: any = {}) {
	        return new Chapter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.matiereId = source["matiereId"];
	        this.tab = source["tab"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.content = source["content"];
	        this.order = source["order"];
	        this.files = this.convertValues(source["files"], Attachment);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class QuizCorrection {
	    questionId: number;
	    answer: number;
	    correctAnswer: number;
	    correct: boolean;
	    explanation?: string;
	
	    static createFrom(source: any = {}) {
	        return new QuizCorrection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.questionId = source["questionId"];
	        this.answer = source["answer"];
	        this.correctAnswer = source["correctAnswer"];
	        this.correct = source["correct"];
	        this.explanation = source["explanation"];
	    }
	}
	export class QuizResult {
	    score: number;
	    total: number;
	    expired: boolean;
	    corrections: QuizCorrection[];
	
	    static createFrom(source: any = {}) {
	        return new QuizResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.score = source["score"];
	        this.total = source["total"];
	        this.expired = source["expired"];
	        this.corrections = this.convertValues(source["corrections"], QuizCorrection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DailyQuizQuestion {
	    id: number;
	    question: string;
	    choices: string[];
	    theme: string;
	    explanation?: string;
	
	    static createFrom(source: any = {}) {
	        return new DailyQuizQuestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.question = source["question"];
	        this.choices = source["choices"];
	        this.theme = source["theme"];
	        this.explanation = source["explanation"];
	    }
	}
	export class DailyQuiz {
	    date: string;
	    questions: DailyQuizQuestion[];
	    startedAt: string;
	    completed: boolean;
	    result?: QuizResult;
	
	    static createFrom(source: any = {}) {
	        return new DailyQuiz(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.questions = this.convertValues(source["questions"], DailyQuizQuestion);
	        this.startedAt = source["startedAt"];
	        this.completed = source["completed"];
	        this.result = this.convertValues(source["result"], QuizResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class TodayTask {
	    id: number;
	    subject: string;
	    color: string;
	    title: string;
	    startTime: string;
	    endTime: string;
	    completed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TodayTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.subject = source["subject"];
	        this.color = source["color"];
	        this.title = source["title"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.completed = source["completed"];
	    }
	}
	export class Progress {
	    toPlan: number;
	    planned: number;
	    inProgress: number;
	    mastered: number;
	
	    static createFrom(source: any = {}) {
	        return new Progress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.toPlan = source["toPlan"];
	        this.planned = source["planned"];
	        this.inProgress = source["inProgress"];
	        this.mastered = source["mastered"];
	    }
	}
	export class Quote {
	    text: string;
	    author: string;
	    source?: string;
	    uncertainAttribution: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Quote(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.author = source["author"];
	        this.source = source["source"];
	        this.uncertainAttribution = source["uncertainAttribution"];
	    }
	}
	export class Dashboard {
	    quote: Quote;
	    progress: Progress;
	    tasks: TodayTask[];
	    today: string;
	
	    static createFrom(source: any = {}) {
	        return new Dashboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.quote = this.convertValues(source["quote"], Quote);
	        this.progress = this.convertValues(source["progress"], Progress);
	        this.tasks = this.convertValues(source["tasks"], TodayTask);
	        this.today = source["today"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PlanningSelection {
	    chapterId: number;
	    startDays: number[];
	    revisionCount: number;
	    durationMinutes: number;
	    spacingDays: number;
	
	    static createFrom(source: any = {}) {
	        return new PlanningSelection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.chapterId = source["chapterId"];
	        this.startDays = source["startDays"];
	        this.revisionCount = source["revisionCount"];
	        this.durationMinutes = source["durationMinutes"];
	        this.spacingDays = source["spacingDays"];
	    }
	}
	export class GeneratePlanningInput {
	    selections: PlanningSelection[];
	    startDate: string;
	
	    static createFrom(source: any = {}) {
	        return new GeneratePlanningInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.selections = this.convertValues(source["selections"], PlanningSelection);
	        this.startDate = source["startDate"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ImportResult {
	    imported: Attachment[];
	    skipped: string[];
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imported = this.convertValues(source["imported"], Attachment);
	        this.skipped = source["skipped"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SubjectWork {
	    id: number;
	    title: string;
	    dueDate: string;
	    completed: boolean;
	    order: number;
	
	    static createFrom(source: any = {}) {
	        return new SubjectWork(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.dueDate = source["dueDate"];
	        this.completed = source["completed"];
	        this.order = source["order"];
	    }
	}
	export class MatiereSummary {
	    id: number;
	    name: string;
	    color: string;
	    order: number;
	    chapters: number;
	    mastered: number;
	
	    static createFrom(source: any = {}) {
	        return new MatiereSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.color = source["color"];
	        this.order = source["order"];
	        this.chapters = source["chapters"];
	        this.mastered = source["mastered"];
	    }
	}
	export class MatiereDetail {
	    subject: MatiereSummary;
	    chapters: Chapter[];
	    works: SubjectWork[];
	
	    static createFrom(source: any = {}) {
	        return new MatiereDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.subject = this.convertValues(source["subject"], MatiereSummary);
	        this.chapters = this.convertValues(source["chapters"], Chapter);
	        this.works = this.convertValues(source["works"], SubjectWork);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class PlanningChapter {
	    id: number;
	    name: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new PlanningChapter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.status = source["status"];
	    }
	}
	export class PlanningSubject {
	    id: number;
	    name: string;
	    color: string;
	    chapters: PlanningChapter[];
	
	    static createFrom(source: any = {}) {
	        return new PlanningSubject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.color = source["color"];
	        this.chapters = this.convertValues(source["chapters"], PlanningChapter);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PlanningData {
	    subjects: PlanningSubject[];
	
	    static createFrom(source: any = {}) {
	        return new PlanningData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.subjects = this.convertValues(source["subjects"], PlanningSubject);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class PlanningTask {
	    id: number;
	    matiereId?: number;
	    chapitreId?: number;
	    title: string;
	    date: string;
	    startTime: string;
	    endTime: string;
	    color: string;
	    notes: string;
	    completed: boolean;
	    subjectName: string;
	    chapterName: string;
	
	    static createFrom(source: any = {}) {
	        return new PlanningTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.matiereId = source["matiereId"];
	        this.chapitreId = source["chapitreId"];
	        this.title = source["title"];
	        this.date = source["date"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.color = source["color"];
	        this.notes = source["notes"];
	        this.completed = source["completed"];
	        this.subjectName = source["subjectName"];
	        this.chapterName = source["chapterName"];
	    }
	}
	export class PlanningTaskInput {
	    matiereId?: number;
	    chapitreId?: number;
	    title: string;
	    date: string;
	    startTime: string;
	    endTime: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new PlanningTaskInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.matiereId = source["matiereId"];
	        this.chapitreId = source["chapitreId"];
	        this.title = source["title"];
	        this.date = source["date"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.notes = source["notes"];
	    }
	}
	
	
	export class QuizHistoryEntry {
	    date: string;
	    score: number;
	    total: number;
	    expired: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QuizHistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.score = source["score"];
	        this.total = source["total"];
	        this.expired = source["expired"];
	    }
	}
	export class QuizProgress {
	    streak: number;
	    totalScore: number;
	    history: QuizHistoryEntry[];
	
	    static createFrom(source: any = {}) {
	        return new QuizProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.streak = source["streak"];
	        this.totalScore = source["totalScore"];
	        this.history = this.convertValues(source["history"], QuizHistoryEntry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class QuizQuestion {
	    id: number;
	    question: string;
	    choices: string[];
	    correctAnswer: number;
	    theme: string;
	    explanation: string;
	
	    static createFrom(source: any = {}) {
	        return new QuizQuestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.question = source["question"];
	        this.choices = source["choices"];
	        this.correctAnswer = source["correctAnswer"];
	        this.theme = source["theme"];
	        this.explanation = source["explanation"];
	    }
	}
	export class QuizQuestionInput {
	    question: string;
	    choices: string[];
	    correctAnswer: number;
	    theme: string;
	    explanation: string;
	
	    static createFrom(source: any = {}) {
	        return new QuizQuestionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.question = source["question"];
	        this.choices = source["choices"];
	        this.correctAnswer = source["correctAnswer"];
	        this.theme = source["theme"];
	        this.explanation = source["explanation"];
	    }
	}
	
	
	
	
	export class WorkdaySlot {
	    period: string;
	    start: string;
	    end: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WorkdaySlot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.period = source["period"];
	        this.start = source["start"];
	        this.end = source["end"];
	        this.enabled = source["enabled"];
	    }
	}
	export class WorkdayPreferences {
	    slots: WorkdaySlot[];
	
	    static createFrom(source: any = {}) {
	        return new WorkdayPreferences(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slots = this.convertValues(source["slots"], WorkdaySlot);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

