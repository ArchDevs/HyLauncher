package config

var launcherDefaults = LauncherConfig{
	Nick:     "HyLauncher",
	Version:  "0.6.5",
	Instance: "default",
}

var instanceDefaults = InstanceConfig{
	ID:     "default",
	Name:   "Default",
	Branch: "release",
	Build:  "0",
}

func Default[T any](v T) T {
	return v
}

func LauncherDefault() LauncherConfig {
	return Default(launcherDefaults)
}

func InstanceDefault() InstanceConfig {
	return Default(instanceDefaults)
}
