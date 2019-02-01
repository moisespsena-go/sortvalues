package main

import (
	"fmt"

	"github.com/moisespsena-go/sortvalues"
)

func main() {
	var s sortvalues.Sorter

	// vs.AnonymousPriority = true // default is false

	// vs.DuplicationType = sortvalues.DUPLICATION_ABORT        // default
	//                   or sortvalues.DUPLICATION_OVERRIDE
	//                   or sortvalues.DUPLICATION_SKIP

	err := s.Append(
		sortvalues.NewValue("anonymou"),
		sortvalues.NewValue("a", "A"),
		sortvalues.NewValue("b", "B"),
		sortvalues.NewValue("c", "C").
			Before("A", "D").
			After("B"),
		sortvalues.NewValue("d", "D"),
	)
	if err != nil {
		fmt.Println("failed to add values: %v", err)
		return
	}

	sorted, err := s.Sort()
	if err != nil {
		fmt.Println("sort failed: %v", err)
		return
	}

	fmt.Println(sorted.Values())
}
