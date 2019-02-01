# sortvalues
Provide sorter for named and anonymous values using topological sorter alghorithm.

## Installation

```bash
go get -u github.com/moisespsena-go/sortvalues
```

## Example

```go
package main

import (
	"fmt"

	"github.com/moisespsena-go/sortvalues"
)

func main() {
	vs := sortvalues.NewValues()
	vs.AnonymousPriority = true
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
```

Output:

    <nil> [anony b c a]
    



