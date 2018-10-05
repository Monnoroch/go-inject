package inject

import (
	"fmt"
	"reflect"
)

/// A dynamic provider is a type containing all the data about a provider that
/// the injector needs to be able to provide it.
type Provider struct {
	/// A provider function. Can be either a function value or a reflect.Value of a function.
	Function interface{}
	/// Whether or not to cache this provider.
	Cached bool
}

/// Dynamic providers module. A type that, instead of having provider methods,
/// as with a static providers module, has a method for generating providers dynamically.
type DynamicModule interface {
	Module
	/// Generate a list of providers.
	Providers() ([]Provider, error)
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
