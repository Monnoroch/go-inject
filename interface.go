/// A dependency injection library for Go.
///
/// This library provides the `Injector` type that is used for providing values and
/// after being configured with a collection of `Module`-s.
/// The library also provides iinterfaces for defining modules both from structs with provider methods and
/// dynamically generated providers.
package inject

import (
	"reflect"
)

/// Module is the interface that has to be implemented by all modules.
/// It is empty, so implementation is trivial.
/// In addition to this interface all Modules have to have methods that have two or three outputs:
/// - A value type.
/// - An annotation type.
/// - Optionally, an error.
/// These methods can have inputs that should come in pairs: values and their annotations.
type Module interface{}

/// Annotation is the interface that has to be implemented by all annotations.
/// It is empty, so implementation is trivial.
type Annotation interface{}

/// A dynamic provider is a type containing all the data about a provider that
/// the injector needs to be able to provide it.
type Provider struct {
	/// A provider function.
	function reflect.Value
	/// Whether or not to cache this provider.
	cached bool
}

/// Create a new provider from either a function or a `reflect.Value` with a function.
func NewProvider(function interface{}) Provider {
	reflectFunction, ok := function.(reflect.Value)
	if !ok {
		reflectFunction = reflect.ValueOf(function)
	}
	return Provider{
		function: reflectFunction,
	}
}

/// Return the `reflect.Value` with the function of this provider.
func (self Provider) Function() reflect.Value {
	return self.function
}

/// Test if the provider is valid: has the right number and types of inputs and outputs.
func (self Provider) IsValid() bool {
	functionType := self.Function().Type()
	return isProvider(functionType) || isProviderWithError(functionType)
}

/// Create a cached or non-cached version of this provider.
func (self Provider) Cached(cached bool) Provider {
	self.cached = cached
	return self
}

/// Test if this provider is cached or not.
func (self Provider) IsCached() bool {
	return self.cached
}

/// Dynamic providers module. A type that, instead of having provider methods,
/// as with a static providers module, has a method for generating providers dynamically.
type DynamicModule interface {
	Module
	/// Generate a list of providers.
	Providers() ([]Provider, error)
}

/// Get the list of providers from the module.
func Providers(module Module) ([]Provider, error) {
	dynamicModule, ok := module.(DynamicModule)
	if !ok {
		dynamicModule = staticProvidersModule{module}
	}
	return dynamicModule.Providers()
}
