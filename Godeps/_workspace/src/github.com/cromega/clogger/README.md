clogger
=======

a very basic wrapper around log.Logger and syslog that provides a common interface and severity filtering

## Usage

```go
package main

import (
  "github.com/cromega/clogger"
  "os"
)

func main() {
  var logger clogger.Logger
  target := os.Getenv("LOG")

  if target == "local" {
    logger = clogger.CreateIoWriter(os.Stdout)
  } else {
    logger = clogger.CreateSyslog("udp", "logs2.papertrailapp:12345", "app")
  }

  logger.SetLevel(clogger.Debug)
  logger.Info("logging is awesome")
}
```

Levels are Debug, Info, Warning, Error, Fatal (Critical) and Off.
