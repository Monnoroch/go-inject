package inject

import (
	"fmt"
	"reflect"
	"strings"
)

/// A wrapper type around a regular module that implements DynamicModule to make
/// provider table generation code more uniform.
type staticProvidersModule struct {
	module Module
}

const providerPrefix = "Provide"
const cachedProviderPrefix = providerPrefix + "Cached"

func (self staticProvidersModule) Providers() ([]Provider, error) {
	providerKeys := map[providerKey]struct{}{}
	providers := []Provider{}
	moduleValue := reflect.ValueOf(self.module)
	moduleType := moduleValue.Type()
	for methodIndex := 0; methodIndex < moduleValue.NumMethod(); methodIndex += 1 {
		method := moduleValue.Method(methodIndex)
		methodDefinition := moduleType.Method(methodIndex)
		provider := NewProvider(method).Cached(strings.HasPrefix(methodDefinition.Name, cachedProviderPrefix))
		if !strings.HasPrefix(methodDefinition.Name, providerPrefix) || !provider.IsValid() {
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
		providers = append(providers, provider)
	}

	return providers, nil
}
