package sortvalues

type ValueInterface interface {
	Before(name ...string) ValueInterface
	GetBefore() []string
	After(name ...string) ValueInterface
	GetAfter() []string
	Value() interface{}
	Name() string
}
