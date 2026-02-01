package hyerrors

func Internal(message string) *Error {
	return New(CategoryValidation, SeverityCritical, message)
}

func Game(message string) *Error {
	return New(CategoryGame, SeverityError, message)
}

func GameCritical(message string) *Error {
	return New(CategoryGame, SeverityCritical, message)
}

func Java(message string) *Error {
	return New(CategoryJava, SeverityError, message)
}

func Network(message string) *Error {
	return New(CategoryNetwork, SeverityError, message)
}

func Validation(message string) *Error {
	return New(CategoryValidation, SeverityWarning, message)
}

func FileSystem(message string) *Error {
	return New(CategoryFileSystem, SeverityCritical, message)
}

func Config(message string) *Error {
	return New(CategoryConfig, SeverityError, message)
}

func Update(message string) *Error {
	return New(CategoryUpdate, SeverityError, message)
}

func WrapGame(err error, message string) *Error {
	return Wrap(err, CategoryGame, message)
}

func WrapJava(err error, message string) *Error {
	return Wrap(err, CategoryJava, message)
}

func WrapNetwork(err error, message string) *Error {
	return Wrap(err, CategoryNetwork, message)
}

func WrapFileSystem(err error, message string) *Error {
	return Wrap(err, CategoryFileSystem, message)
}

func WrapConfig(err error, message string) *Error {
	return Wrap(err, CategoryConfig, message)
}

func WrapUpdate(err error, message string) *Error {
	return Wrap(err, CategoryUpdate, message)
}
