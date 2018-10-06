/// A dependency injection library for Go.
///
/// This library provides the `Injector` type that is used for providing values and
/// after being configured with a collection of `Module`-s.
/// The library also provides iinterfaces for defining modules both from structs with provider methods and
/// dynamically generated providers.
package inject

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
