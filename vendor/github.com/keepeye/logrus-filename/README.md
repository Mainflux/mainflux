This package is extracted from [onrik/logrus][1].

USAGE
======

```go
package main

import (
	"github.com/sirupsen/logrus"
	"github.com/keepeye/logrus-filename"
)

func main() {
	filenameHook := filename.NewHook()
	filenameHook.Field = "line"
	logrus.AddHook(filenameHook)
	logrus.Info("aha")
}
```

output:

```go
INFO[0000] aha                                       line="box-api-server/test.go:12"
```

[1]: https://github.com/onrik/logrus
