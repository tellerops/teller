# ellipsis
Go modules to insert an ellipses into the middle of a long string to shorten it.

## Install
```
go get github.com/jftuga/ellipsis
```

## License
[MIT License](https://github.com/jftuga/ellipsis/blob/main/LICENSE)

## Example

```go
package main

import (
	"fmt"
	"github.com/jftuga/ellipsis"
)

func main() {
	s := "abcdefghijklmnopqurstvwxyz"
	for i := 0; i <= 26; i++ {
		s := ellipsis.Shorten(s, i)
		fmt.Printf("%2d %s\n", i, s)
	}
}
```

## Output

```
 0 
 1 z
 2 yz
 3 xyz
 4 ...z
 5 a...z
 6 a...yz
 7 ab...yz
 8 ab...xyz
 9 abc...xyz
10 abc...wxyz
11 abcd...wxyz
12 abcd...vwxyz
13 abcde...vwxyz
14 abcde...tvwxyz
15 abcdef...tvwxyz
16 abcdef...stvwxyz
17 abcdefg...stvwxyz
18 abcdefg...rstvwxyz
19 abcdefgh...rstvwxyz
20 abcdefgh...urstvwxyz
21 abcdefghi...urstvwxyz
22 abcdefghi...qurstvwxyz
23 abcdefghij...qurstvwxyz
24 abcdefghij...pqurstvwxyz
25 abcdefghijk...pqurstvwxyz
26 abcdefghijklmnopqurstvwxyz
```
