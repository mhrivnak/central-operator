package controller

import (
	"github.com/mhrivnak/central-operator/pkg/controller/watch"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, watch.Add)
}
