package controllers

type logger interface {
	Error(err error, message string, keysAndValues ...interface{})
	Info(message string, keysAndValues ...interface{})
}
