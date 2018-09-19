package inject

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Injector is a component for providing values exported by modules.
type Injector struct {
	providers *providersData
	cacheLock sync.Mutex
	cache     map[providerKey]valueErrorPair
}

type valueErrorPair struct {
	value interface{}
	err   error
}

// Create an injector from the list of modules.
func InjectorOf(modules ...Module) (*Injector, error) {
	providers, err := buildProviders(CombineModules(modules...))
	if err != nil {
		return nil, err
	}

	return &Injector{
		providers: providers,
		cache:     map[providerKey]valueErrorPair{},
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
	return self.getLocked(key)
}

func (self *Injector) getLocked(key providerKey) (interface{}, error) {
	self.cacheLock.Lock()
	defer self.cacheLock.Unlock()
	return self.getCached(key)
}

func (self *Injector) getCached(key providerKey) (providedValue interface{}, err error) {
	if provider, ok := self.providers.providers[key]; ok && provider.cached {
		if value, ok := self.cache[key]; ok {
			return value.value, value.err
		}
		defer func() {
			self.cache[key] = valueErrorPair{
				value: providedValue,
				err:   err,
			}
		}()
	}
	return self.get(key)
}

type lazyProviderError struct {
	cause error
}

var injectOutsideInjectorCallError = errors.New("Trying to call a lazy provider outside of an Injector.Get call")

func (self *Injector) get(key providerKey) (interface{}, error) {
	provider, ok := self.providers.providers[key]
	if !ok {
		return nil, provideError{key: key, cause: errors.New("No provider found")}
	}

	arguments := make([]reflect.Value, len(provider.arguments)*2)
	injectionTime := true
	for index, argument := range provider.arguments {
		argumentKey := argument.key
		offset := index * 2
		if lazyArgumentType := getLazyArgumentType(argumentKey); lazyArgumentType != nil {
			strictArgumentKey := providerKey{valueType: lazyArgumentType, annotationType: argumentKey.annotationType}
			arguments[offset] = reflect.MakeFunc(argumentKey.valueType, func(_ []reflect.Value) []reflect.Value {
				if !injectionTime {
					panic(injectOutsideInjectorCallError)
				}

				result, err := self.getCached(strictArgumentKey)
				if err != nil {
					panic(lazyProviderError{cause: err})
				}
				return []reflect.Value{reflect.ValueOf(result)}
			})
		} else {
			argumentValue, err := self.getCached(argumentKey)
			if err != nil {
				return nil, provideError{key: key, cause: err}
			}
			arguments[offset] = getValueForArgument(argumentValue, argumentKey.valueType)
		}
		annotationType := argument.originalAnnotationType
		if annotationType == nil {
			annotationType = argumentKey.annotationType
		}
		arguments[offset+1] = reflect.Zero(annotationType)
	}

	outputs, err := callProviderHandlingLazyErrors(provider.provider, arguments)
	injectionTime = false
	if err != nil {
		return nil, provideError{key: key, cause: err}
	}

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

func callProviderHandlingLazyErrors(
	provider reflect.Value,
	arguments []reflect.Value,
) (result []reflect.Value, resultingErr error) {
	defer func() {
		if err := recover(); err != nil {
			if lazyProviderErr, ok := err.(lazyProviderError); ok {
				resultingErr = lazyProviderErr.cause
			} else {
				panic(err)
			}
		}
	}()
	return provider.Call(arguments), nil
}

func getLazyArgumentType(key providerKey) reflect.Type {
	if key.valueType.Kind() != reflect.Func {
		return nil
	}
	// Types that are aliases of function types have Kind() == Func,
	// but have distinct names.
	if !strings.HasPrefix(key.valueType.String(), "func() ") {
		return nil
	}
	if key.valueType.NumOut() != 1 {
		return nil
	}
	return key.valueType.Out(0)
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
