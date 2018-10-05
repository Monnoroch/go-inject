package inject

import (
	"fmt"
	"reflect"
	"strings"
)

type providerKey struct {
	valueType      reflect.Type
	annotationType reflect.Type
}

type providerData struct {
	provider  reflect.Value
	arguments []providerKey
	hasError  bool
	cached    bool
}

type providersData struct {
	// A map of provider keys to provider functions.
	providers map[providerKey]providerData
}

/// A wrapper type around a regular module that implements DynamicModule to make
/// provider table generation code more uniform.
type staticProvidersModule struct {
	module Module
}

func (self staticProvidersModule) Providers() ([]Provider, error) {
	providerKeys := map[providerKey]struct{}{}
	providers := []Provider{}
	moduleValue := reflect.ValueOf(self.module)
	moduleType := moduleValue.Type()
	for methodIndex := 0; methodIndex < moduleValue.NumMethod(); methodIndex += 1 {
		method := moduleValue.Method(methodIndex)
		methodDefinition := moduleType.Method(methodIndex)
		if !isProvider(method, methodDefinition) &&
			!isProviderWithError(method, methodDefinition) {
			return nil, fmt.Errorf(
				"%#v is not a module: it has an invalid provider %#v.",
				self.module, method)
		}

		methodType := method.Type()
		key := providerKey{
			valueType:      methodType.Out(0),
			annotationType: methodType.Out(1),
		}
		if _, ok := providerKeys[key]; ok {
			return nil, fmt.Errorf("Duplicate providers for key %v in module %#v", key, self.module)
		}
		providerKeys[key] = struct{}{}

		providers = append(providers, Provider{
			Function: method,
			Cached:   strings.HasPrefix(methodDefinition.Name, cachedProviderPrefix),
		})
	}

	return providers, nil
}

func buildProviders(module Module) (*providersData, error) {
	providers := &providersData{
		providers: map[providerKey]providerData{},
	}
	for _, module := range flattenModule(module) {
		dynamicModule, ok := module.(DynamicModule)
		if !ok {
			dynamicModule = staticProvidersModule{module}
		}
		if err := buildProvidersFromDynamicModule(dynamicModule, providers); err != nil {
			return nil, err
		}
	}
	return providers, nil
}

var globalAnnotationType = reflect.TypeOf((*Annotation)(nil)).Elem()
var globalErrorType = reflect.TypeOf((*error)(nil)).Elem()

const providerPrefix = "Provide"
const cachedProviderPrefix = providerPrefix + "Cached"

func isProvider(method reflect.Value, methodDefinition reflect.Method) bool {
	if !strings.HasPrefix(methodDefinition.Name, providerPrefix) {
		return false
	}
	return isProviderFunction(method.Type())
}

func isProviderFunction(methodType reflect.Type) bool {
	if methodType.Kind() != reflect.Func {
		return false
	}
	if methodType.NumOut() != 2 {
		return false
	}
	return hasAnnotationOutput(methodType) && hasInputsWithAnnotations(methodType)
}

func isProviderWithError(method reflect.Value, methodDefinition reflect.Method) bool {
	if !strings.HasPrefix(methodDefinition.Name, providerPrefix) {
		return false
	}
	return isProviderWithErrorFunction(method.Type())
}

func isProviderWithErrorFunction(methodType reflect.Type) bool {
	if methodType.Kind() != reflect.Func {
		return false
	}
	if methodType.NumOut() != 3 {
		return false
	}
	if !isError(methodType.Out(2)) {
		return false
	}
	return hasAnnotationOutput(methodType) && hasInputsWithAnnotations(methodType)
}

func hasAnnotationOutput(methodType reflect.Type) bool {
	return isAnnotation(methodType.Out(1))
}

func hasInputsWithAnnotations(methodType reflect.Type) bool {
	numberOfInputs := methodType.NumIn()
	if numberOfInputs%2 != 0 {
		return false
	}

	for inputIndex := 1; inputIndex < numberOfInputs; inputIndex += 2 {
		annotationInput := methodType.In(inputIndex)
		if !isAnnotation(annotationInput) {
			return false
		}
	}
	return true
}

func isAnnotation(valueType reflect.Type) bool {
	return valueType.Implements(globalAnnotationType)
}

func isError(valueType reflect.Type) bool {
	return valueType == globalErrorType || valueType.Implements(globalErrorType)
}
