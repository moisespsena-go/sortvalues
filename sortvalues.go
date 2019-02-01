package sortvalues

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/moisespsena/go-topsort"
)

const (
	DUPLICATION_ABORT    DuplicationType = 0
	DUPLICATION_OVERRIDE DuplicationType = 1
	DUPLICATION_SKIP     DuplicationType = 2
)

type DuplicationType int

func (dt DuplicationType) String() string {
	switch dt {
	case 0:
		return "ABORT"
	case 1:
		return "OVERRIDE"
	case 2:
		return "SKIP"
	default:
		return "<invalid value>"
	}
}

type ErrDuplicate struct {
	Value *Value
}

func (e ErrDuplicate) Error() string {
	return fmt.Sprintf("Duplicate value %q: %s", e.Value.Name, e.Value.Value)
}

var ErrUnamed = errors.New("Unnamed value")

type Slice []*Value

func (s Slice) Values() []interface{} {
	values := make([]interface{}, len(s))
	for i, v := range s {
		values[i] = v.Value
	}
	return values
}

type Value struct {
	Value       interface{}
	Name        string
	BeforeNames []string
	AfterNames  []string
}

func NewValue(value interface{}, name ...string) *Value {
	if len(name) == 0 {
		name = make([]string, 1)
	}
	return &Value{Value: value, Name: name[0]}
}

func (v *Value) Before(name ...string) *Value {
	if v.Name == "" {
		panic(ErrUnamed)
	}
	v.BeforeNames = append(v.BeforeNames, name...)
	return v
}

func (v *Value) After(name ...string) *Value {
	if v.Name == "" {
		panic(ErrUnamed)
	}
	v.AfterNames = append(v.AfterNames, name...)
	return v
}

type Values struct {
	Named             map[string]int
	NamedSlice        Slice
	Anonymous         Slice
	DuplicationType   DuplicationType
	AnonymousPriority bool
}

func NewValues(duplicationType ...DuplicationType) *Values {
	if len(duplicationType) == 0 {
		duplicationType = make([]DuplicationType, 1)
	}
	return &Values{DuplicationType: duplicationType[0]}
}

func (vs *Values) AppendOption(dt DuplicationType, v ...*Value) error {
	if vs.Named == nil {
		vs.Named = map[string]int{}
	}

	for _, v := range v {
		if v.Name == "" {
			vs.Anonymous = append(vs.Anonymous, v)
		} else if i, ok := vs.Named[v.Name]; ok {
			switch dt {
			case DUPLICATION_OVERRIDE:
				vs.NamedSlice[i] = v
			case DUPLICATION_ABORT:
				return &ErrDuplicate{v}
			case DUPLICATION_SKIP:
			default:
				return fmt.Errorf("Invalid duplication type %d", dt)
			}
		} else {
			vs.Named[v.Name] = len(vs.NamedSlice)
			vs.NamedSlice = append(vs.NamedSlice, v)
		}
	}
	return nil
}

func (vs *Values) Append(v ...*Value) error {
	return vs.AppendOption(vs.DuplicationType, v...)
}

func (vs *Values) Sort() (values Slice, err error) {
	notFound := make(map[string][]string)

	graph := topsort.NewGraph()

	for _, v := range vs.NamedSlice {
		graph.AddNode(v.Name)
	}

	for _, v := range vs.NamedSlice {
		for _, to := range v.BeforeNames {
			if _, ok := vs.Named[to]; ok {
				_ = graph.AddEdge(v.Name, to)
			} else {
				if _, ok := notFound[v.Name]; !ok {
					notFound[v.Name] = make([]string, 1)
				}
				notFound[v.Name] = append(notFound[v.Name], to)
			}
		}
		for _, from := range v.AfterNames {
			if _, ok := vs.Named[from]; ok {
				graph.AddEdge(from, v.Name)
			} else {
				if _, ok := notFound[v.Name]; ok {
					notFound[v.Name] = make([]string, 1)
				}
				notFound[v.Name] = append(notFound[v.Name], from)
			}
		}
	}

	if len(notFound) > 0 {
		var msgs []string
		for n, items := range notFound {
			msgs = append(msgs, fmt.Sprintf("Required by %q: %v.", n, strings.Join(items, ", ")))
		}
		panic(fmt.Errorf("Values dependency error:\n - %v\n", strings.Join(msgs, "\n - ")))
	}

	names, err := graph.DepthFirst()

	if err != nil {
		panic(fmt.Errorf("Topological values sorter error: %v", err))
	}

	values = make(Slice, len(vs.Anonymous)+len(vs.NamedSlice))
	var i int

	if vs.AnonymousPriority {
		vs.addAnonymousTo(values)
		i = len(vs.Anonymous)
	}

	vs.addNamedTo(values[i:], names)
	i += len(names)

	if !vs.AnonymousPriority {
		vs.addAnonymousTo(values[i:])
	}

	return
}

func (vs *Values) addAnonymousTo(s Slice) {
	for i, v := range vs.Anonymous {
		s[i] = v
	}
}

func (vs *Values) addNamedTo(s Slice, names []string) {
	for i, name := range names {
		s[i] = vs.NamedSlice[vs.Named[name]]
	}
}
