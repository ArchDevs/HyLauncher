export namespace hyerrors {
	
	export class Frame {
	    function: string;
	    file: string;
	    line: number;
	
	    static createFrom(source: any = {}) {
	        return new Frame(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.function = source["function"];
	        this.file = source["file"];
	        this.line = source["line"];
	    }
	}
	export class Error {
	    id: string;
	    category: string;
	    severity: number;
	    message: string;
	    details?: string;
	    // Go type: time
	    timestamp: any;
	    stack?: Frame[];
	    context?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Error(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.severity = source["severity"];
	        this.message = source["message"];
	        this.details = source["details"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.stack = this.convertValues(source["stack"], Frame);
	        this.context = source["context"];
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

export namespace model {
	
	export class InstanceModel {
	    InstanceID: string;
	    InstanceName: string;
	    Branch: string;
	    BuildVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new InstanceModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.InstanceID = source["InstanceID"];
	        this.InstanceName = source["InstanceName"];
	        this.Branch = source["Branch"];
	        this.BuildVersion = source["BuildVersion"];
	    }
	}

}

export namespace service {
	
	export class LogEntry {
	    // Go type: time
	    timestamp: any;
	    severity: number;
	    category: string;
	    message: string;
	    details?: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.severity = source["severity"];
	        this.category = source["category"];
	        this.message = source["message"];
	        this.details = source["details"];
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
	export class SystemInfo {
	    os: string;
	    arch: string;
	    num_cpu: number;
	    go_version: string;
	    num_goroutine: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.num_cpu = source["num_cpu"];
	        this.go_version = source["go_version"];
	        this.num_goroutine = source["num_goroutine"];
	    }
	}
	export class CrashReport {
	    id: string;
	    // Go type: time
	    timestamp: any;
	    app_version: string;
	    error?: hyerrors.Error;
	    system: SystemInfo;
	    recent_logs?: LogEntry[];
	
	    static createFrom(source: any = {}) {
	        return new CrashReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.app_version = source["app_version"];
	        this.error = this.convertValues(source["error"], hyerrors.Error);
	        this.system = this.convertValues(source["system"], SystemInfo);
	        this.recent_logs = this.convertValues(source["recent_logs"], LogEntry);
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
	
	export class NewsArticle {
	    title: string;
	    dest_url: string;
	    description: string;
	    image_url: string;
	
	    static createFrom(source: any = {}) {
	        return new NewsArticle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.dest_url = source["dest_url"];
	        this.description = source["description"];
	        this.image_url = source["image_url"];
	    }
	}

}

export namespace updater {
	
	export class Asset {
	    url: string;
	    sha256: string;
	
	    static createFrom(source: any = {}) {
	        return new Asset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.sha256 = source["sha256"];
	    }
	}

}

