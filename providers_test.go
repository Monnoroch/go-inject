package inject

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type BuildProvidersTests struct {
	suite.Suite
}

func (self *BuildProvidersTests) TestEmptyModule() {
	type testEmptyModule struct{}
	providers, err := buildProviders(testEmptyModule{})
	self.Require().Nil(err)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{},
	}, providers)
}

type testBadMethodNameModule struct{}

func (self testBadMethodNameModule) NotAProvider() {}

func (self *BuildProvidersTests) TestBadMethodName() {
	_, err := buildProviders(testBadMethodNameModule{})
	self.Contains(err.Error(), "not a module")
}

type testNoResultMethodModule struct{}

func (self testNoResultMethodModule) ProvideInvalid() {}

func (self *BuildProvidersTests) TestNoResultMethod() {
	_, err := buildProviders(testNoResultMethodModule{})
	self.Contains(err.Error(), "not a module")
}

type testNoAnnotationMethodModule struct{}

func (self testNoAnnotationMethodModule) ProvideInvalid() int {
	return 0
}

func (self *BuildProvidersTests) TestNoAnnotationMethod() {
	_, err := buildProviders(testNoAnnotationMethodModule{})
	self.Contains(err.Error(), "not a module")
}

type testInvalidErrorTypeMethodModule struct{}

func (self testInvalidErrorTypeMethodModule) ProvideInvalid() (int, int, int) {
	return 0, 0, 0
}

func (self *BuildProvidersTests) TestInvalidErrorTypeMethod() {
	_, err := buildProviders(testInvalidErrorTypeMethodModule{})
	self.Contains(err.Error(), "not a module")
}

type testArgumentWithoutAnnotationMethodModule struct{}

func (self testArgumentWithoutAnnotationMethodModule) ProvideInvalid(_ int) (int, int) {
	return 0, 0
}

func (self *BuildProvidersTests) TestArgumentWithoutAnnotationMethod() {
	_, err := buildProviders(testArgumentWithoutAnnotationMethodModule{})
	self.Contains(err.Error(), "not a module")
}

type testSameProviderTwiceModule struct{}

func (self testSameProviderTwiceModule) Provide1() (int, int) {
	return 0, 0
}

func (self testSameProviderTwiceModule) Provide2() (int, int) {
	return 0, 0
}

func (self *BuildProvidersTests) TestSameProviderTwice() {
	_, err := buildProviders(testSameProviderTwiceModule{})
	self.Contains(err.Error(), "Duplicate providers for key")
}

type testSameProviderTwiceModule1 struct{}

func (self testSameProviderTwiceModule1) Provide() (int, int) {
	return 0, 0
}

type testSameProviderTwiceModule2 struct{}

func (self testSameProviderTwiceModule2) Provide() (int, int) {
	return 0, 0
}

func (self *BuildProvidersTests) TestSameProviderTwiceActossModules() {
	_, err := buildProviders(CombineModules(testSameProviderTwiceModule1{}, testSameProviderTwiceModule2{}))
	self.Contains(err.Error(), "Duplicate providers for key")
}

func (self *BuildProvidersTests) TestIdenticalProviderTwiceActossModules() {
	module := testSameProviderTwiceModule1{}
	providers, err := buildProviders(CombineModules(module, testSameProviderTwiceModule1{}))
	self.Nil(err)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(int(0)),
			}: {
				provider:  reflect.ValueOf(module).MethodByName("Provide"),
				arguments: []providerKey{},
				hasError:  false,
			},
		},
	}, providers)
}

type testValidProviderWithNoArgumentsModule struct{}

func (self testValidProviderWithNoArgumentsModule) Provide() (int32, int64) {
	return 0, 0
}

func (self *BuildProvidersTests) TestValidProviderWithNoArguments() {
	module := testValidProviderWithNoArgumentsModule{}
	providers, err := buildProviders(module)
	self.Nil(err)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(int64(0)),
			}: {
				provider:  reflect.ValueOf(module).MethodByName("Provide"),
				arguments: []providerKey{},
				hasError:  false,
			},
		},
	}, providers)
}

type testValidProviderWithArgumentModule struct{}

func (self testValidProviderWithArgumentModule) Provide(_ int32, _ bool) (int32, int64) {
	return 0, 0
}

func (self *BuildProvidersTests) TestValidProviderWithArgument() {
	module := testValidProviderWithArgumentModule{}
	providers, err := buildProviders(module)
	self.Nil(err)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(int64(0)),
			}: {
				provider: reflect.ValueOf(module).MethodByName("Provide"),
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(int32(0)),
					annotationType: reflect.TypeOf(bool(false)),
				}},
				hasError: false,
			},
		},
	}, providers)
}

type testValidProviderWithArgumentsModule struct{}

func (self testValidProviderWithArgumentsModule) Provide(_ int32, _ bool, _ int64, _ string) (int32, int64) {
	return 0, 0
}

func (self *BuildProvidersTests) TestValidProviderWithArguments() {
	module := testValidProviderWithArgumentsModule{}
	providers, err := buildProviders(module)
	self.Nil(err)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(int64(0)),
			}: {
				provider: reflect.ValueOf(module).MethodByName("Provide"),
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(int32(0)),
					annotationType: reflect.TypeOf(bool(false)),
				}, {
					valueType:      reflect.TypeOf(int64(0)),
					annotationType: reflect.TypeOf(string("")),
				}},
				hasError: false,
			},
		},
	}, providers)
}

type testValidErrorProviderModule struct{}

func (self testValidErrorProviderModule) Provide() (int32, int64, error) {
	return 0, 0, nil
}

func (self *BuildProvidersTests) TestValidErrorProvider() {
	module := testValidErrorProviderModule{}
	providers, err := buildProviders(module)
	self.Nil(err)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(int64(0)),
			}: {
				provider:  reflect.ValueOf(module).MethodByName("Provide"),
				arguments: []providerKey{},
				hasError:  true,
			},
		},
	}, providers)
}

type testTwoValidProvidersModule struct{}

func (self testTwoValidProvidersModule) Provide1() (int32, int64) {
	return 0, 0
}

func (self testTwoValidProvidersModule) Provide2(_ int32, _ int64) (int32, bool) {
	return 0, false
}

func (self *BuildProvidersTests) TestTwoValidProviders() {
	module := testTwoValidProvidersModule{}
	providers, err := buildProviders(module)
	self.Nil(err)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(int64(0)),
			}: {
				provider:  reflect.ValueOf(module).MethodByName("Provide1"),
				arguments: []providerKey{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(bool(false)),
			}: {
				provider: reflect.ValueOf(module).MethodByName("Provide2"),
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(int32(0)),
					annotationType: reflect.TypeOf(int64(0)),
				}},
				hasError: false,
			},
		},
	}, providers)
}

type testValidProvidersModule1 struct{}

func (self testValidProvidersModule1) Provide1() (int32, int64) {
	return 0, 0
}

type testValidProvidersModule2 struct{}

func (self testValidProvidersModule2) Provide2(_ int32, _ int64) (int32, bool) {
	return 0, false
}

func (self *BuildProvidersTests) TestValidProvidersInDifferentModules() {
	module1 := testValidProvidersModule1{}
	module2 := testValidProvidersModule2{}
	providers, err := buildProviders(CombineModules(module1, module2))
	self.Nil(err)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(int64(0)),
			}: {
				provider:  reflect.ValueOf(module1).MethodByName("Provide1"),
				arguments: []providerKey{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(bool(false)),
			}: {
				provider: reflect.ValueOf(module2).MethodByName("Provide2"),
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(int32(0)),
					annotationType: reflect.TypeOf(int64(0)),
				}},
				hasError: false,
			},
		},
	}, providers)
}

type testCachedProviderModule struct{}

func (self testCachedProviderModule) ProvideCachedValue() (int32, int64) {
	return 0, 0
}

func (self *BuildProvidersTests) TestCachedProvider() {
	module := testCachedProviderModule{}
	providers, err := buildProviders(module)
	self.Require().Nil(err)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(int64(0)),
			}: {
				provider:  reflect.ValueOf(module).MethodByName("ProvideCachedValue"),
				arguments: []providerKey{},
				hasError:  false,
				cached:    true,
			},
		},
	}, providers)
}

func (self *BuildProvidersTests) TestAutoInjectEmptyStruct() {
	type Struct struct{}
	providers, err := buildProviders(AutoInjectModule(new(Struct), Annotation1{}, struct{}{}))
	self.Require().Nil(err)
	removeProviderFunctions(providers)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(Struct{}),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				arguments: []providerKey{},
			},
		},
	}, providers)
}

func (self *BuildProvidersTests) TestAutoInjectStructAuto() {
	type Struct1 struct{}
	type Struct2 struct {
		Struct1 Struct1
	}
	type Struct3 struct {
		Struct1 Struct1
		Struct2 Struct2
	}
	providers, err := buildProviders(CombineModules(
		AutoInjectModule(new(Struct1), Auto{}, struct{}{}),
		AutoInjectModule(new(Struct2), Auto{}, struct{}{}),
		AutoInjectModule(new(Struct3), Auto{}, struct{}{}),
	))
	self.Require().Nil(err)
	removeProviderFunctions(providers)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(Struct1{}),
				annotationType: reflect.TypeOf(Auto{}),
			}: {
				arguments: []providerKey{},
			},
			{
				valueType:      reflect.TypeOf(Struct2{}),
				annotationType: reflect.TypeOf(Auto{}),
			}: {
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(Struct1{}),
					annotationType: reflect.TypeOf(Auto{}),
				}},
			},
			{
				valueType:      reflect.TypeOf(Struct3{}),
				annotationType: reflect.TypeOf(Auto{}),
			}: {
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(Struct1{}),
					annotationType: reflect.TypeOf(Auto{}),
				}, {
					valueType:      reflect.TypeOf(Struct2{}),
					annotationType: reflect.TypeOf(Auto{}),
				}},
			},
		},
	}, providers)
}

func (self *BuildProvidersTests) TestAutoInjectStructCustomAnnotations() {
	type Struct1 struct{}
	type Struct2 struct {
		Struct1 Struct1
	}
	type Struct3 struct {
		Struct1 Struct1
		Struct2 Struct2
	}
	providers, err := buildProviders(CombineModules(
		AutoInjectModule(new(Struct1), Annotation1{}, struct{}{}),
		AutoInjectModule(new(Struct2), Annotation2{}, struct {
			Struct1 Annotation1
		}{}),
		AutoInjectModule(new(Struct3), Annotation3{}, struct {
			Struct1 Annotation1
			Struct2 Annotation2
		}{}),
	))
	self.Require().Nil(err)
	removeProviderFunctions(providers)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(Struct1{}),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				arguments: []providerKey{},
			},
			{
				valueType:      reflect.TypeOf(Struct2{}),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(Struct1{}),
					annotationType: reflect.TypeOf(Annotation1{}),
				}},
			},
			{
				valueType:      reflect.TypeOf(Struct3{}),
				annotationType: reflect.TypeOf(Annotation3{}),
			}: {
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(Struct1{}),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, {
					valueType:      reflect.TypeOf(Struct2{}),
					annotationType: reflect.TypeOf(Annotation2{}),
				}},
			},
		},
	}, providers)
}

type Struct1Default struct{}
type Struct2Default struct {
	Struct1 Struct1Default
}

func (self Struct2Default) ProvideAutoInjectAnnotations() interface{} {
	return struct {
		Struct1 Annotation1
	}{}
}

type Struct3Default struct {
	Struct1 Struct1Default
	Struct2 Struct2Default
}

func (self Struct3Default) ProvideAutoInjectAnnotations() interface{} {
	return struct {
		Struct1 Annotation1
		Struct2 Annotation2
	}{}
}
func (self *BuildProvidersTests) TestAutoInjectStructDefaultAnnotations() {
	providers, err := buildProviders(CombineModules(
		AutoInjectModule(new(Struct1Default), Annotation1{}, struct{}{}),
		AutoInjectModule(new(Struct2Default), Annotation2{}, struct{}{}),
		AutoInjectModule(new(Struct3Default), Annotation3{}, struct{}{}),
	))
	self.Require().Nil(err)
	removeProviderFunctions(providers)
	self.Equal(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(Struct1Default{}),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				arguments: []providerKey{},
			},
			{
				valueType:      reflect.TypeOf(Struct2Default{}),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(Struct1Default{}),
					annotationType: reflect.TypeOf(Annotation1{}),
				}},
			},
			{
				valueType:      reflect.TypeOf(Struct3Default{}),
				annotationType: reflect.TypeOf(Annotation3{}),
			}: {
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(Struct1Default{}),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, {
					valueType:      reflect.TypeOf(Struct2Default{}),
					annotationType: reflect.TypeOf(Annotation2{}),
				}},
			},
		},
	}, providers)
}

func (self *BuildProvidersTests) TestAutoInjectPrimitive() {
	_, err := buildProviders(AutoInjectModule(new(int), Annotation1{}, struct{}{}))
	self.Contains(err.Error(), "int is not a struct")
}

func removeProviderFunctions(providers *providersData) {
	for key, provider := range providers.providers {
		provider.provider = reflect.Value{}
		providers.providers[key] = provider
	}
}

func TestBuildProviders(t *testing.T) {
	suite.Run(t, new(BuildProvidersTests))
}
