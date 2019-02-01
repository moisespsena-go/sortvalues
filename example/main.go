package main

import (
	"fmt"

	"github.com/moisespsena-go/sortvalues"
)

func main() {
	vs := sortvalues.NewValues()

	// vs.AnonymousPriority = true // default is false

	// vs.DuplicationType = sortvalues.DUPLICATION_ABORT        // default
	//                   or sortvalues.DUPLICATION_OVERRIDE
	//                   or sortvalues.DUPLICATION_SKIP

	vs.Append(
		sortvalues.NewValue("anonymou"),
		sortvalues.NewValue("a", "A"),
		sortvalues.NewValue("b", "B"),
		sortvalues.NewValue("c", "C").
			Before("A", "D").
			After("B"),
		sortvalues.NewValue("d", "D"),
	)
	sorted, err := vs.Sort()
	fmt.Println(err, sorted.Values())
}
