package constant

import (
	"reflect"

	"github.com/monnoroch/go-inject"
)

type constantModule struct {
	value      interface{}
	annotation inject.Annotation
}

func (self constantModule) Providers() ([]inject.Provider, error) {
	annotationType := reflect.TypeOf(self.annotation)
	return []inject.Provider{inject.NewProvider(reflect.MakeFunc(
		reflect.FuncOf(
			[]reflect.Type{},
			[]reflect.Type{
				reflect.TypeOf(self.value),
				annotationType,
			},
			false,
		),
		func(_ []reflect.Value) []reflect.Value {
			return []reflect.Value{
				reflect.ValueOf(self.value),
				reflect.Zero(annotationType),
			}
		},
	),
	)}, nil
}

/// Creates a module that provides a constant value with a specified annotation.
func ConstantModule(value interface{}, annotation inject.Annotation) inject.Module {
	return constantModule{value: value, annotation: annotation}
}
