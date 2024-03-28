# Geamon

A library used create background service for golang

> Windows not supported!!

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
	backd := geamon.Geamon{
		Stdout: file,
		Stderr: file,
		PidFile: "./logs/pid/hello.pid",
		ProcessTitle: "hellod",
	}
	if err := backd.Run(); err != nil {
		panic(err)
	}
	defer backd.ReleasePidFile()

	// do blocked services
}
```
