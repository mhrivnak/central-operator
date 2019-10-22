package controller

import (
	"github.com/operator-framework/operator-sdk/pkg/ansible/proxy/controllermap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, *controllermap.ControllerMap) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, cMap *controllermap.ControllerMap) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, cMap); err != nil {
			return err
		}
	}
	return nil
}
