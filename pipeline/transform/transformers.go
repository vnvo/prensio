package transform

import "github.com/dop251/goja"

func registerTransformers(vm *goja.Runtime) error {
	vm.Set("tObfString", tObfString)
	return nil
}

func tObfString(value string) string {
	return "****"
}
