package sortvalues

type ValueInterface interface {
	Before(name ...string) ValueInterface
	After(name ...string) ValueInterface
	Value() interface{}
	Name() string
}
