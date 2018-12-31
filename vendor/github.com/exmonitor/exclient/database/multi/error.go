package multi

import "errors"

var invalidConfigError error = errors.New("invalid config")

var executionFailedError error = errors.New("execution failed")
