# Geamon

A library used create daemon for golang

```sh
go get github.com/mengdu/geamon
```

**Demo**

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mengdu/geamon"
)

func main() {
	file, err := os.OpenFile("./logs/deamon.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	deamon := geamon.Geamon{
		Stdout: file,
		Stderr: file,
		MaxRestart: 3,
		PidFile: "/var/run/hello.pid",
	}
	if err := deamon.Run(); err != nil {
		panic(err)
	}

  // do blocked services
}
```
