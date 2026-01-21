package config

func SaveNick(nick string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.Nick = nick
	return Save(cfg)
}

func GetNick() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	return cfg.Nick, nil
}

func SaveLocalGameVersion(version int) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.CurrentGameVersion = version
	return Save(cfg)
}

func GetLocalGameVersion() (int, error) {
	cfg, err := Load()
	if err != nil {
		return 0, err
	}
	return cfg.CurrentGameVersion, nil
}

func SaveBranch(branch string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.Branch = branch
	return Save(cfg)
}

func GetBranch() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	return cfg.Branch, nil
}
