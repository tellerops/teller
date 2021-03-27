wincred
=======

Go wrapper around the Windows Credential Manager API functions.

![Go](https://github.com/danieljoos/wincred/workflows/Go/badge.svg)
[![GoDoc](https://godoc.org/github.com/danieljoos/wincred?status.svg)](https://godoc.org/github.com/danieljoos/wincred)


Installation
------------

```Go
go get github.com/danieljoos/wincred
```


Usage
-----

See the following examples:

### Create and store a new generic credential object
```Go
package main

import (
    "fmt"
    "github.com/danieljoos/wincred"
)

func main() {
    cred := wincred.NewGenericCredential("myGoApplication")
    cred.CredentialBlob = []byte("my secret")
    err := cred.Write()
    
    if err != nil {
        fmt.Println(err)
    }
} 
```

### Retrieve a credential object
```Go
package main

import (
    "fmt"
    "github.com/danieljoos/wincred"
)

func main() {
    cred, err := wincred.GetGenericCredential("myGoApplication")
    if err == nil {
        fmt.Println(string(cred.CredentialBlob))
    }
} 
```

### Remove a credential object
```Go
package main

import (
    "fmt"
    "github.com/danieljoos/wincred"
)

func main() {
    cred, err := wincred.GetGenericCredential("myGoApplication")
    if err != nil {
        fmt.Println(err)
        return
    }
    cred.Delete()
} 
```

### List all available credentials
```Go
package main

import (
    "fmt"
    "github.com/danieljoos/wincred"
)

func main() {
    creds, err := wincred.List()
    if err != nil {
        fmt.Println(err)
        return
    }
    for i := range(creds) {
        fmt.Println(creds[i].TargetName)
    }
}
```
