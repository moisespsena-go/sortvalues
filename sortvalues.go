package sortvalues

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/moisespsena-go/topsort"
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
	Value ValueInterface
}

func (e ErrDuplicate) Error() string {
	return fmt.Sprintf("Duplicate value %q: %s", e.Value.Name(), e.Value.Value())
}

var ErrUnamed = errors.New("Unnamed value")

type Slice []ValueInterface

func (s Slice) Values() []interface{} {
	values := make([]interface{}, len(s))
	for i, v := range s {
		values[i] = v.Value()
	}
	return values
}

type Value struct {
	value      interface{}
	name       string
	afterNames []string
	after      []string
}

func NewValue(value interface{}, name ...string) *Value {
	if len(name) == 0 {
		name = make([]string, 1)
	}
	return &Value{value: value, name: name[0]}
}

func (v *Value) Value() interface{} {
	return v.value
}

func (v *Value) Name() string {
	return v.name
}

func (v *Value) GetBefore() []string {
	return v.afterNames
}

func (v *Value) GetAfter() []string {
	return v.after
}

func (v *Value) Before(name ...string) ValueInterface {
	if v.Name() == "" {
		panic(ErrUnamed)
	}
	v.afterNames = append(v.afterNames, name...)
	return v
}

func (v *Value) After(name ...string) ValueInterface {
	if v.Name() == "" {
		panic(ErrUnamed)
	}
	v.after = append(v.after, name...)
	return v
}

type Sorter struct {
	Named             map[string]int
	NamedSlice        Slice
	Anonymous         Slice
	DuplicationType   DuplicationType
	AnonymousPriority bool
	mu                sync.Mutex
}

func New(duplicationType ...DuplicationType) *Sorter {
	if len(duplicationType) == 0 {
		duplicationType = make([]DuplicationType, 1)
	}
	return &Sorter{DuplicationType: duplicationType[0]}
}

func (s *Sorter) AppendOption(dt DuplicationType, v ...ValueInterface) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Named == nil {
		s.Named = map[string]int{}
	}

	for _, v := range v {
		if v.Name() == "" {
			s.Anonymous = append(s.Anonymous, v)
		} else if i, ok := s.Named[v.Name()]; ok {
			switch dt {
			case DUPLICATION_OVERRIDE:
				s.NamedSlice[i] = v
			case DUPLICATION_ABORT:
				return &ErrDuplicate{v}
			case DUPLICATION_SKIP:
			default:
				return fmt.Errorf("Invalid duplication type %d", dt)
			}
		} else {
			s.Named[v.Name()] = len(s.NamedSlice)
			s.NamedSlice = append(s.NamedSlice, v)
		}
	}
	return nil
}

func (s *Sorter) Append(v ...ValueInterface) error {
	return s.AppendOption(s.DuplicationType, v...)
}

func (s *Sorter) Sort() (values Slice, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	notFound := make(map[string][]string)

	graph := topsort.NewGraph()

	for _, v := range s.NamedSlice {
		graph.AddNode(v.Name())
	}

	for _, v := range s.NamedSlice {
		for _, to := range v.GetBefore() {
			if _, ok := s.Named[to]; ok {
				_ = graph.AddEdge(v.Name(), to)
			} else {
				if _, ok := notFound[v.Name()]; !ok {
					notFound[v.Name()] = make([]string, 1)
				}
				notFound[v.Name()] = append(notFound[v.Name()], to)
			}
		}
		for _, from := range v.GetAfter() {
			if _, ok := s.Named[from]; ok {
				graph.AddEdge(from, v.Name())
			} else {
				if _, ok := notFound[v.Name()]; ok {
					notFound[v.Name()] = make([]string, 1)
				}
				notFound[v.Name()] = append(notFound[v.Name()], from)
			}
		}
	}

	if len(notFound) > 0 {
		var msgs []string
		for n, items := range notFound {
			msgs = append(msgs, fmt.Sprintf("Required by %q: %v.", n, strings.Join(items, ", ")))
		}
		panic(fmt.Errorf("Sorter dependency error:\n - %v\n", strings.Join(msgs, "\n - ")))
	}

	names, err := graph.DepthFirst()

	if err != nil {
		panic(fmt.Errorf("Topological values sorter error: %v", err))
	}

	values = make(Slice, len(s.Anonymous)+len(s.NamedSlice))
	var i int

	if s.AnonymousPriority {
		s.addAnonymousTo(values)
		i = len(s.Anonymous)
	}

	s.addNamedTo(values[i:], names)
	i += len(names)

	if !s.AnonymousPriority {
		s.addAnonymousTo(values[i:])
	}

	return
}

func (s *Sorter) addAnonymousTo(sl Slice) {
	for i, v := range s.Anonymous {
		sl[i] = v
	}
}

func (s *Sorter) addNamedTo(sl Slice, names []string) {
	for i, name := range names {
		sl[i] = s.NamedSlice[s.Named[name]]
	}
}
