package config

var launcherDefaults = LauncherConfig{
	Nick:       "HyLauncher",
	Instance:   "default",
	DiscordRPC: true,
}

var instanceDefaults = InstanceConfig{
	ID:     "default",
	Name:   "Default",
	Branch: "release",
	Build:  "auto",
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
