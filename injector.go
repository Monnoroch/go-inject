package inject

import (
	"errors"
	"fmt"
	"reflect"
)

// Injector is a component for providing values exported by modules.
type Injector struct {
	providers *providersData
}

// Create an injector from the list of modules.
func InjectorOf(modules ...Module) (*Injector, error) {
	providers, err := buildProviders(CombineModules(modules...))
	if err != nil {
		return nil, err
	}

	return &Injector{
		providers: providers,
	}, nil
}

// Get the annotated value from the injector, panic if there was an error.
func (self *Injector) MustGet(pointerToType interface{}, annotation Annotation) interface{} {
	result, err := self.Get(pointerToType, annotation)
	if err != nil {
		panic(err)
	}
	return result
}

// Get the annotated value from the injector.
func (self *Injector) Get(pointerToType interface{}, annotation Annotation) (interface{}, error) {
	valueType := reflect.TypeOf(pointerToType).Elem()
	annotationType := reflect.TypeOf(annotation)
	key := providerKey{
		valueType:      valueType,
		annotationType: annotationType,
	}
	return self.get(key)
}

func (self *Injector) get(key providerKey) (interface{}, error) {
	provider, ok := self.providers.providers[key]
	if !ok {
		return nil, provideError{key: key, cause: errors.New("No provider found")}
	}

	arguments := make([]reflect.Value, len(provider.arguments)*2)
	for index, argumentKey := range provider.arguments {
		argument, err := self.get(argumentKey)
		if err != nil {
			return nil, provideError{key: key, cause: err}
		}

		arguments[index*2] = getValueForArgument(argument, argumentKey.valueType)
		arguments[index*2+1] = reflect.Zero(argumentKey.annotationType)
	}

	outputs := provider.provider.Call(arguments)
	output := outputs[0].Interface()
	if !provider.hasError {
		return output, nil
	}

	if err := outputs[2].Interface(); err != nil {
		return output, provideError{key: key, cause: err.(error)}
	} else {
		return output, nil
	}
}

func getValueForArgument(argument interface{}, valueType reflect.Type) reflect.Value {
	// When a provider returns `nil`, the return type is lost and we need to create a value
	// of that type explicitly.
	if argument == nil {
		return reflect.Zero(valueType)
	}
	return reflect.ValueOf(argument)
}

type provideError struct {
	key   providerKey
	cause error
}

func (self provideError) Error() string {
	return fmt.Sprintf("error providing type %v with annotation %v: %s",
		self.key.valueType, self.key.annotationType, self.cause.Error())
}
