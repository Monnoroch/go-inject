package inject

import (
	"fmt"
	"reflect"
)

/// Default annotation for auto-injected types.
type Auto struct{}

/// An interface to be implemented to support auto-injecting a type.
type AutoInjectable interface {
	/// Returns a mapping of field names to annotations.
	/// Omitted fields imply `inject.Auto` annotation.
	/// Not implementing this method implies all fields having `inject.Auto` annotation.
	ProvideAutoInjectAnnotations() interface{}
}

type autoInjectModule struct {
	typePointer interface{}
	annotation  Annotation
	annotations interface{}
	cached      bool
}

/// Create a module for providing a struct type with an annotation with dependencies
/// inferred from the annotations struct.
func AutoInjectModule(typePointer interface{}, annotation Annotation, annotations interface{}) Module {
	return autoInjectModule{
		typePointer: typePointer,
		annotation:  annotation,
		annotations: annotations,
		cached:      false,
	}
}

/// Create a module for providing a struct type with an annotation with dependencies
/// inferred from the annotations struct.
func AutoInjectCachedModule(typePointer interface{}, annotation Annotation, annotations interface{}) Module {
	return autoInjectModule{
		typePointer: typePointer,
		annotation:  annotation,
		annotations: annotations,
		cached:      true,
	}
}

var autoAnnotationType = reflect.TypeOf(Auto{})
var autoInjectableType = reflect.TypeOf((*AutoInjectable)(nil)).Elem()

func (self autoInjectModule) Providers() ([]Provider, error) {
	valueType := reflect.TypeOf(self.typePointer).Elem()
	annotationType := reflect.TypeOf(self.annotation)

	if valueType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%v is not a struct", reflect.TypeOf(self.typePointer))
	}

	annotationByField := map[string]reflect.Type{}
	if valueType.Implements(autoInjectableType) {
		asAutoInjectable := reflect.ValueOf(self.typePointer).Elem().Interface().(AutoInjectable)
		defaultAnnotations := asAutoInjectable.ProvideAutoInjectAnnotations()
		extractAnnotations(defaultAnnotations, annotationByField)
	}
	extractAnnotations(self.annotations, annotationByField)

	providerArgumentTypes := []reflect.Type{}
	for i := 0; i < valueType.NumField(); i += 1 {
		field := valueType.Field(i)
		fieldAnnotationType, ok := annotationByField[field.Name]
		if !ok {
			fieldAnnotationType = autoAnnotationType
		}
		providerArgumentTypes = append(providerArgumentTypes, field.Type, fieldAnnotationType)
	}
	provider := Provider{
		Function: reflect.MakeFunc(
			reflect.FuncOf(
				providerArgumentTypes,
				[]reflect.Type{valueType, annotationType},
				false,
			),
			func(arguments []reflect.Value) []reflect.Value {
				result := reflect.New(valueType).Elem()
				for i := 0; i < result.NumField()*2; i += 2 {
					result.Field(i / 2).Set(arguments[i])
				}
				return []reflect.Value{result, reflect.Zero(annotationType)}
			},
		),
		Cached: self.cached,
	}
	return []Provider{provider}, nil
}

func extractAnnotations(annotationsStruct interface{}, annotationByField map[string]reflect.Type) {
	annotations := reflect.TypeOf(annotationsStruct)
	for i := 0; i < annotations.NumField(); i += 1 {
		field := annotations.Field(i)
		annotationByField[field.Name] = field.Type
	}
}
