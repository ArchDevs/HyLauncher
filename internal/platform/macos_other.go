//go:build !darwin

package platform

func RemoveQuarantine(path string) error {
	return nil
}

func AdHocSign(path string) error {
	return nil
}

func RemoveSignature(path string) error {
	return nil
}

func FixMacOSApp(appPath string) error {
	return nil
}
