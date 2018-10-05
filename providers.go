package inject

import (
	"fmt"
	"reflect"
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

func buildProvidersFromDynamicModule(module DynamicModule, providers *providersData) error {
	dynamicProviders, err := module.Providers()
	if err != nil {
		return err
	}

	for _, dynamicProvider := range dynamicProviders {
		if err := buildProvidersFromDynamicProvider(dynamicProvider, providers); err != nil {
			return err
		}
	}
	return nil
}

func buildProvidersFromDynamicProvider(dynamicProvider Provider, providers *providersData) error {
	function, ok := dynamicProvider.Function.(reflect.Value)
	if !ok {
		function = reflect.ValueOf(dynamicProvider.Function)
	}
	functionType := function.Type()
	if !isProviderFunction(functionType) && !isProviderWithErrorFunction(functionType) {
		return fmt.Errorf("%#v is an invalid provider.", dynamicProvider)
	}

	arguments := make([]providerKey, 0, functionType.NumIn()/2)
	for inputIndex := 0; inputIndex < functionType.NumIn(); inputIndex += 2 {
		valueInput := functionType.In(inputIndex)
		annotationInput := functionType.In(inputIndex + 1)
		arguments = append(arguments, providerKey{
			valueType:      valueInput,
			annotationType: annotationInput,
		})
	}

	provider := providerData{
		provider:  function,
		arguments: arguments,
		cached:    dynamicProvider.Cached,
		hasError:  functionType.NumOut() == 3,
	}
	key := providerKey{
		valueType:      functionType.Out(0),
		annotationType: functionType.Out(1),
	}
	if existingProvider, ok := providers.providers[key]; ok {
		if !reflect.DeepEqual(existingProvider.provider, provider.provider) {
			return fmt.Errorf(
				"Duplicate providers for key {%v, %v}",
				key.valueType, key.annotationType)
		}
	}
	providers.providers[key] = provider
	return nil
}

var globalAnnotationType = reflect.TypeOf((*Annotation)(nil)).Elem()
var globalErrorType = reflect.TypeOf((*error)(nil)).Elem()

func isProviderFunction(methodType reflect.Type) bool {
	if methodType.Kind() != reflect.Func {
		return false
	}
	if methodType.NumOut() != 2 {
		return false
	}
	return hasAnnotationOutput(methodType) && hasInputsWithAnnotations(methodType)
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
