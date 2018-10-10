/// rewrite provides tools to dynamically transform modules to chenge their providers using reflection.
package rewrite

import (
	"fmt"
	"reflect"

	"github.com/monnoroch/go-inject"
)

/// Generate a module that takes all input module's providers and replaces specified annotations
/// according to the `annotationsToRewrite` map.
func RewriteAnnotations(
	module inject.Module,
	annotationsToRewrite map[inject.Annotation]inject.Annotation,
) inject.DynamicModule {
	return rewriteAnnotationsModule{
		module:               module,
		annotationsToRewrite: annotationsToRewrite,
	}
}

type rewriteAnnotationsModule struct {
	module               inject.Module
	annotationsToRewrite map[inject.Annotation]inject.Annotation
}

func (self rewriteAnnotationsModule) Providers() ([]inject.Provider, error) {
	providers, err := inject.Providers(self.module)
	if err != nil {
		return nil, err
	}

	annotationsToRewrite := map[reflect.Type]reflect.Type{}
	for from, to := range self.annotationsToRewrite {
		annotationsToRewrite[reflect.TypeOf(from)] = reflect.TypeOf(to)
	}

	newProviders := make([]inject.Provider, len(providers))
	for index, provider := range providers {
		if !provider.IsValid() {
			return nil, fmt.Errorf("invalid provider %v", provider)
		}

		function := provider.Function()
		functionType := function.Type()

		providerArgumentTypes := make([]reflect.Type, functionType.NumIn())
		for inputIndex := 0; inputIndex < functionType.NumIn(); inputIndex += 2 {
			annotationType := functionType.In(inputIndex + 1)
			if rewrittenType, ok := annotationsToRewrite[annotationType]; ok {
				annotationType = rewrittenType
			}
			providerArgumentTypes[inputIndex] = functionType.In(inputIndex)
			providerArgumentTypes[inputIndex+1] = annotationType
		}

		valueType := functionType.Out(0)
		annotationType := functionType.Out(1)
		if rewrittenType, ok := annotationsToRewrite[annotationType]; ok {
			annotationType = rewrittenType
		}

		returnTypes := []reflect.Type{valueType, annotationType}
		// Provider with an error.
		if functionType.NumOut() == 3 {
			returnTypes = append(returnTypes, reflect.TypeOf((*error)(nil)).Elem())
		}
		newProviders[index] = inject.NewProvider(reflect.MakeFunc(
			reflect.FuncOf(
				providerArgumentTypes,
				returnTypes,
				false,
			),
			func(arguments []reflect.Value) []reflect.Value {
				newArguments := make([]reflect.Value, functionType.NumIn())
				for inputIndex := 0; inputIndex < functionType.NumIn(); inputIndex += 2 {
					newArguments[inputIndex] = arguments[inputIndex]
					newArguments[inputIndex+1] = reflect.Zero(functionType.In(inputIndex + 1))
				}
				resutls := function.Call(newArguments)
				resutls[1] = reflect.Zero(annotationType)
				return resutls
			},
		)).Cached(provider.IsCached())
	}
	return newProviders, nil
}
