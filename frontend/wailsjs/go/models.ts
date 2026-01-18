export namespace app {
	
	export class CoverImage {
	    s3Key: string;
	
	    static createFrom(source: any = {}) {
	        return new CoverImage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.s3Key = source["s3Key"];
	    }
	}
	export class BlogPost {
	    _id: string;
	    title: string;
	    // Go type: time
	    publishedAt: any;
	    slug: string;
	    coverImage: CoverImage;
	    bodyExcerpt: string;
	    author: string;
	
	    static createFrom(source: any = {}) {
	        return new BlogPost(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._id = source["_id"];
	        this.title = source["title"];
	        this.publishedAt = this.convertValues(source["publishedAt"], null);
	        this.slug = source["slug"];
	        this.coverImage = this.convertValues(source["coverImage"], CoverImage);
	        this.bodyExcerpt = source["bodyExcerpt"];
	        this.author = source["author"];
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
	export class ConnectivityInfo {
	    can_reach_game_server: boolean;
	    can_reach_itchio_server: boolean;
	    game_server_error?: string;
	    itchio_server_error?: string;
	    response_time_ms: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectivityInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.can_reach_game_server = source["can_reach_game_server"];
	        this.can_reach_itchio_server = source["can_reach_itchio_server"];
	        this.game_server_error = source["game_server_error"];
	        this.itchio_server_error = source["itchio_server_error"];
	        this.response_time_ms = source["response_time_ms"];
	    }
	}
	
	export class DiskSpaceInfo {
	    install_directory: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new DiskSpaceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.install_directory = source["install_directory"];
	        this.error = source["error"];
	    }
	}
	export class ServerVersionInfo {
	    latest_version: number;
	    found_versions: boolean;
	    checked_urls?: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ServerVersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.latest_version = source["latest_version"];
	        this.found_versions = source["found_versions"];
	        this.checked_urls = source["checked_urls"];
	        this.error = source["error"];
	    }
	}
	export class InstallationInfo {
	    game_installed: boolean;
	    current_version: string;
	    install_path: string;
	    jre_installed: boolean;
	    butler_installed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new InstallationInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game_installed = source["game_installed"];
	        this.current_version = source["current_version"];
	        this.install_path = source["install_path"];
	        this.jre_installed = source["jre_installed"];
	        this.butler_installed = source["butler_installed"];
	    }
	}
	export class PlatformInfo {
	    os: string;
	    arch: string;
	    go_version: string;
	    num_cpu: number;
	    num_goroutine: number;
	
	    static createFrom(source: any = {}) {
	        return new PlatformInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.go_version = source["go_version"];
	        this.num_cpu = source["num_cpu"];
	        this.num_goroutine = source["num_goroutine"];
	    }
	}
	export class DiagnosticReport {
	    // Go type: time
	    timestamp: any;
	    app_version: string;
	    platform: PlatformInfo;
	    connectivity: ConnectivityInfo;
	    local_installation: InstallationInfo;
	    server_versions: ServerVersionInfo;
	    disk_space: DiskSpaceInfo;
	
	    static createFrom(source: any = {}) {
	        return new DiagnosticReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.app_version = source["app_version"];
	        this.platform = this.convertValues(source["platform"], PlatformInfo);
	        this.connectivity = this.convertValues(source["connectivity"], ConnectivityInfo);
	        this.local_installation = this.convertValues(source["local_installation"], InstallationInfo);
	        this.server_versions = this.convertValues(source["server_versions"], ServerVersionInfo);
	        this.disk_space = this.convertValues(source["disk_space"], DiskSpaceInfo);
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
	
	export class GameVersions {
	    current: string;
	    latest: string;
	
	    static createFrom(source: any = {}) {
	        return new GameVersions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current = source["current"];
	        this.latest = source["latest"];
	    }
	}
	
	

}

export namespace config {
	
	export class GameSettings {
	    minMemory: number;
	    maxMemory: number;
	    width: number;
	    height: number;
	    fullscreen: boolean;
	    javaArgs: string;
	    gameDir: string;
	    channel: string;
	    gameVersion: number;
	    onlineFix: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GameSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.minMemory = source["minMemory"];
	        this.maxMemory = source["maxMemory"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.fullscreen = source["fullscreen"];
	        this.javaArgs = source["javaArgs"];
	        this.gameDir = source["gameDir"];
	        this.channel = source["channel"];
	        this.gameVersion = source["gameVersion"];
	        this.onlineFix = source["onlineFix"];
	    }
	}
	export class Profile {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}

}

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

