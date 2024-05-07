# pydeohub
pydeohub in GoLang


### Create go.mod
```

go mod init github.com/StechLabs/pydeohub
go mod tidy
```

## Example
```
mkdir myhub
cd myhub
go mod init myhub
go get github.com/StechLabs/pydeohub
touch main.go
```

```golang
package main

import (
	"fmt"

	"github.com/StechLabs/pydeohub/videohub" // Import the videohub package
)

func main() {
	ip := "192.168.0.150"

	fmt.Println("IP: ", ip)

	vh := videohub.NewVideohub(ip)

	// Now you can use methods of the Videohub struct, like vh.Route(), vh.InputLabel(), etc.
	// Use vh to perform some action, for example:
	vh.Route(0, 0) // Route output 1 to input 2
	vh.InputLabel(1, "Camera 2")
	vh.OutputLabel(0, "Switcher 1")
}
```

### Run application
```go run main.go```