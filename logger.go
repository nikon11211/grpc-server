package grpc

type Logger interface {
	DebugF(msg string, args ...any)
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}
