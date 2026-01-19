export namespace diagnostics {
	
	export class SystemInfo {
	    num_cpu: number;
	    goos: string;
	    goarch: string;
	    go_version: string;
	    num_goroutine: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.num_cpu = source["num_cpu"];
	        this.goos = source["goos"];
	        this.goarch = source["goarch"];
	        this.go_version = source["go_version"];
	        this.num_goroutine = source["num_goroutine"];
	    }
	}
	export class CrashReport {
	    // Go type: time
	    timestamp: any;
	    app_version: string;
	    os: string;
	    arch: string;
	    error?: hyerrors.AppError;
	    system_info: SystemInfo;
	    recent_logs?: string[];
	
	    static createFrom(source: any = {}) {
	        return new CrashReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.app_version = source["app_version"];
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.error = this.convertValues(source["error"], hyerrors.AppError);
	        this.system_info = this.convertValues(source["system_info"], SystemInfo);
	        this.recent_logs = source["recent_logs"];
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

export namespace hyerrors {
	
	export class AppError {
	    type: string;
	    message: string;
	    technical: string;
	    // Go type: time
	    timestamp: any;
	    stack: string;
	
	    static createFrom(source: any = {}) {
	        return new AppError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.message = source["message"];
	        this.technical = source["technical"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.stack = source["stack"];
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

