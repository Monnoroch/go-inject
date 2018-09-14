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

func buildProvidersForAutoInjectModule(module autoInjectModule, providers *providersData) error {
	key := providerKey{
		valueType:      reflect.TypeOf(module.typePointer).Elem(),
		annotationType: reflect.TypeOf(module.annotation),
	}

	if key.valueType.Kind() != reflect.Struct {
		return fmt.Errorf("%v is not a struct", reflect.TypeOf(module.typePointer))
	}

	annotationByField := map[string]reflect.Type{}
	if key.valueType.Implements(autoInjectableType) {
		asAutoInjectable := reflect.ValueOf(module.typePointer).Elem().Interface().(AutoInjectable)
		defaultAnnotations := asAutoInjectable.ProvideAutoInjectAnnotations()
		extractAnnotations(defaultAnnotations, annotationByField)
	}
	extractAnnotations(module.annotations, annotationByField)

	arguments := []providerArgument{}
	providerArgumentTypes := []reflect.Type{}
	for i := 0; i < key.valueType.NumField(); i += 1 {
		field := key.valueType.Field(i)
		annotationType, ok := annotationByField[field.Name]
		if !ok {
			annotationType = autoAnnotationType
		}
		arguments = append(arguments, providerArgument{providerKey{
			valueType:      field.Type,
			annotationType: annotationType,
		}, nil})
		providerArgumentTypes = append(providerArgumentTypes, field.Type, annotationType)
	}

	provider := providerData{
		provider: reflect.MakeFunc(
			reflect.FuncOf(
				providerArgumentTypes,
				[]reflect.Type{key.valueType, key.annotationType},
				false,
			),
			func(arguments []reflect.Value) []reflect.Value {
				result := reflect.New(key.valueType).Elem()
				for i := 0; i < result.NumField()*2; i += 2 {
					result.Field(i / 2).Set(arguments[i])
				}
				return []reflect.Value{result, reflect.Zero(key.annotationType)}
			},
		),
		arguments: arguments,
		cached:    module.cached,
		hasError:  false,
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

func extractAnnotations(annotationsStruct interface{}, annotationByField map[string]reflect.Type) {
	annotations := reflect.TypeOf(annotationsStruct)
	for i := 0; i < annotations.NumField(); i += 1 {
		field := annotations.Field(i)
		annotationByField[field.Name] = field.Type
	}
}
