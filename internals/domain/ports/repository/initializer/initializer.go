package initializer

type InitializerPort interface {
	CheckEnvironment() error
	SetEnvironment() error
}
