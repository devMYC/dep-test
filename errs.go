package main

import "fmt"

// ArgError for argument errors
type ArgError struct {
	FuncName string
	ArgName  string
	Reason   string
}

func (ae *ArgError) Error() string {
	return fmt.Sprintf("Invalid argument \"%s\" %s [%s]", ae.ArgName, ae.Reason, ae.FuncName)
}
