export namespace main {
	
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
	
	

}

