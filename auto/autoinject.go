package autoinject

import (
	"fmt"
	"reflect"

	"github.com/monnoroch/go-inject"
)

/// Default annotation for auto-injected types.
/// Can be used if the program does not have two components of the same type.
type Auto struct{}

/// An interface to be implemented to support auto-injecting a type.
type AutoInjectable interface {
	/// Returns a mapping of field names to annotations.
	/// Omitted fields imply `autoinject.Auto` annotation.
	/// Not implementing this method implies all fields having `autoinject.Auto` annotation.
	ProvideAutoInjectAnnotations() interface{}
}

type autoInjectModule struct {
	typePointer interface{}
	annotation  inject.Annotation
	annotations interface{}
	cached      bool
}

/// Create a module for automatically providing a struct type with the default `autoinject.Auto` annotation.
func AutoInjectModule(typePointer interface{}) autoInjectModule {
	return autoInjectModule{
		typePointer: typePointer,
		annotation:  Auto{},
		annotations: struct{}{},
		cached:      false,
	}
}

/// Auto-inject the value with a custom annotation.
func (self autoInjectModule) WithAnnotation(annotation inject.Annotation) autoInjectModule {
	self.annotation = annotation
	return self
}

/// Auto-inject the value with a custom annotation.
func (self autoInjectModule) WithAnnotations(annotations interface{}) autoInjectModule {
	self.annotations = annotations
	return self
}

/// Make the generated provider cached.
func (self autoInjectModule) Cached() autoInjectModule {
	self.cached = true
	return self
}

/// Make the generated provider not cached.
func (self autoInjectModule) NotCached() autoInjectModule {
	self.cached = false
	return self
}

var autoAnnotationType = reflect.TypeOf(Auto{})
var autoInjectableType = reflect.TypeOf((*AutoInjectable)(nil)).Elem()

func (self autoInjectModule) Providers() ([]inject.Provider, error) {
	valueType := reflect.TypeOf(self.typePointer).Elem()
	annotationType := reflect.TypeOf(self.annotation)
	dereferencedValueType, derefed := getDereferencedType(valueType)

	if dereferencedValueType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%v is not a struct", dereferencedValueType)
	}

	annotationByField := map[string]reflect.Type{}
	if valueType.Implements(autoInjectableType) {
		asAutoInjectable := reflect.ValueOf(self.typePointer).Elem().Interface().(AutoInjectable)
		defaultAnnotations := asAutoInjectable.ProvideAutoInjectAnnotations()
		extractAnnotations(defaultAnnotations, annotationByField)
	}
	extractAnnotations(self.annotations, annotationByField)

	providerArgumentTypes := []reflect.Type{}
	for i := 0; i < dereferencedValueType.NumField(); i += 1 {
		field := dereferencedValueType.Field(i)
		fieldAnnotationType, ok := annotationByField[field.Name]
		if !ok {
			fieldAnnotationType = autoAnnotationType
		}
		providerArgumentTypes = append(providerArgumentTypes, field.Type, fieldAnnotationType)
	}
	provider := inject.Provider{
		Function: reflect.MakeFunc(
			reflect.FuncOf(
				providerArgumentTypes,
				[]reflect.Type{valueType, annotationType},
				false,
			),
			func(arguments []reflect.Value) []reflect.Value {
				result := reflect.New(dereferencedValueType).Elem()
				for i := 0; i < result.NumField()*2; i += 2 {
					result.Field(i / 2).Set(arguments[i])
				}
				if derefed {
					result = result.Addr()
				}
				return []reflect.Value{result, reflect.Zero(annotationType)}
			},
		),
		Cached: self.cached,
	}
	return []inject.Provider{provider}, nil
}

func getDereferencedType(valueType reflect.Type) (reflect.Type, bool) {
	if valueType.Kind() == reflect.Ptr {
		return valueType.Elem(), true
	}
	return valueType, false
}

func extractAnnotations(annotationsStruct interface{}, annotationByField map[string]reflect.Type) {
	annotations := reflect.TypeOf(annotationsStruct)
	for i := 0; i < annotations.NumField(); i += 1 {
		field := annotations.Field(i)
		annotationByField[field.Name] = field.Type
	}
}
