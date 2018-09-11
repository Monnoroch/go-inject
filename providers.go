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
		if err := buildProvidersForLeafModule(module, providers); err != nil {
			return nil, err
		}
	}
	return providers, nil
}

type internalProviderData struct {
	provider  reflect.Value
	output    providerKey
	arguments []providerKey
	module    Module
	hasError  bool
}

type moduleProvidersData struct {
	// A map of provider keys to provider functions.
	providers map[providerKey]internalProviderData
}

func buildProvidersForLeafModule(module Module, providers *providersData) error {
	moduleProviders, err := buildModuleProvidersForLeafModule(module)
	if err != nil {
		return err
	}

	for _, provider := range moduleProviders.providers {
		if existingProvider, ok := providers.providers[provider.output]; ok {
			if !reflect.DeepEqual(existingProvider.provider, provider.provider) {
				return fmt.Errorf(
					"Duplicate providers for key {%v, %v}",
					provider.output.valueType, provider.output.annotationType)
			}
		}
		providers.providers[provider.output] = providerData{
			provider:  provider.provider,
			arguments: provider.arguments,
			hasError:  provider.hasError,
		}
	}
	return nil
}

func buildModuleProvidersForLeafModule(module Module) (*moduleProvidersData, error) {
	providers := moduleProvidersData{
		providers: map[providerKey]internalProviderData{},
	}
	moduleValue := reflect.ValueOf(module)
	moduleType := moduleValue.Type()
	for methodIndex := 0; methodIndex < moduleValue.NumMethod(); methodIndex += 1 {
		method := moduleValue.Method(methodIndex)
		methodDefinition := moduleType.Method(methodIndex)
		if !isProvider(method, methodDefinition) &&
			!isProviderWithError(method, methodDefinition) {
			return nil, fmt.Errorf(
				"%#v is not a module: it has an invalid provider %#v.",
				module, method)
		}

		methodType := method.Type()
		arguments := make([]providerKey, 0, methodType.NumIn()/2)
		for inputIndex := 0; inputIndex < methodType.NumIn(); inputIndex += 2 {
			valueInput := methodType.In(inputIndex)
			annotationInput := methodType.In(inputIndex + 1)
			arguments = append(arguments, providerKey{
				valueType:      valueInput,
				annotationType: annotationInput,
			})
		}

		key := providerKey{
			valueType:      methodType.Out(0),
			annotationType: methodType.Out(1),
		}
		if _, ok := providers.providers[key]; ok {
			return nil, fmt.Errorf("Duplicate providers for key %v inmodule %#v", key, module)
		}

		providers.providers[key] = internalProviderData{
			hasError:  methodType.NumOut() == 3,
			module:    module,
			provider:  method,
			output:    key,
			arguments: arguments,
		}
	}
	return &providers, nil
}

var globalAnnotationType = reflect.TypeOf((*Annotation)(nil)).Elem()
var globalErrorType = reflect.TypeOf((*error)(nil)).Elem()

const providerPrefix = "Provide"

func isProvider(method reflect.Value, methodDefinition reflect.Method) bool {
	if !strings.HasPrefix(methodDefinition.Name, providerPrefix) {
		return false
	}

	methodType := method.Type()
	if methodType.NumOut() != 2 {
		return false
	}
	return hasAnnotationOutput(methodType) && hasInputsWithAnnotations(methodType)
}

func isProviderWithError(method reflect.Value, methodDefinition reflect.Method) bool {
	if !strings.HasPrefix(methodDefinition.Name, providerPrefix) {
		return false
	}

	methodType := method.Type()
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
