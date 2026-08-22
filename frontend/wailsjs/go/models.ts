export namespace main {
	
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
	
	
	
	

}

