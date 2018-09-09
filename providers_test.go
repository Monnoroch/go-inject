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

type testNoResultMethodModule struct{}

func (self testNoResultMethodModule) InvalidProvider() {}

func (self *BuildProvidersTests) TestNoResultMethod() {
	_, err := buildProviders(testNoResultMethodModule{})
	self.Contains(err.Error(), "not a module")
}

type testNoAnnotationMethodModule struct{}

func (self testNoAnnotationMethodModule) InvalidProvider() int {
	return 0
}

func (self *BuildProvidersTests) TestNoAnnotationMethod() {
	_, err := buildProviders(testNoAnnotationMethodModule{})
	self.Contains(err.Error(), "not a module")
}

type testInvalidErrorTypeMethodModule struct{}

func (self testInvalidErrorTypeMethodModule) InvalidProvider() (int, int, int) {
	return 0, 0, 0
}

func (self *BuildProvidersTests) TestInvalidErrorTypeMethod() {
	_, err := buildProviders(testInvalidErrorTypeMethodModule{})
	self.Contains(err.Error(), "not a module")
}

type testArgumentWithoutAnnotationMethodModule struct{}

func (self testArgumentWithoutAnnotationMethodModule) InvalidProvider(_ int) (int, int) {
	return 0, 0
}

func (self *BuildProvidersTests) TestArgumentWithoutAnnotationMethod() {
	_, err := buildProviders(testArgumentWithoutAnnotationMethodModule{})
	self.Contains(err.Error(), "not a module")
}

type testSameProviderTwiceModule struct{}

func (self testSameProviderTwiceModule) Provider1() (int, int) {
	return 0, 0
}

func (self testSameProviderTwiceModule) Provider2() (int, int) {
	return 0, 0
}

func (self *BuildProvidersTests) TestSameProviderTwice() {
	_, err := buildProviders(testSameProviderTwiceModule{})
	self.Contains(err.Error(), "Duplicate providers for key")
}

type testSameProviderTwiceModule1 struct{}

func (self testSameProviderTwiceModule1) Provider() (int, int) {
	return 0, 0
}

type testSameProviderTwiceModule2 struct{}

func (self testSameProviderTwiceModule2) Provider() (int, int) {
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
				provider:  reflect.ValueOf(module).MethodByName("Provider"),
				arguments: []providerKey{},
				hasError:  false,
			},
		},
	}, providers)
}

type testValidProviderWithNoArgumentsModule struct{}

func (self testValidProviderWithNoArgumentsModule) Provider() (int32, int64) {
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
				provider:  reflect.ValueOf(module).MethodByName("Provider"),
				arguments: []providerKey{},
				hasError:  false,
			},
		},
	}, providers)
}

type testValidProviderWithArgumentModule struct{}

func (self testValidProviderWithArgumentModule) Provider(_ int32, _ bool) (int32, int64) {
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
				provider: reflect.ValueOf(module).MethodByName("Provider"),
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

func (self testValidProviderWithArgumentsModule) Provider(_ int32, _ bool, _ int64, _ string) (int32, int64) {
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
				provider: reflect.ValueOf(module).MethodByName("Provider"),
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

func (self testValidErrorProviderModule) Provider() (int32, int64, error) {
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
				provider:  reflect.ValueOf(module).MethodByName("Provider"),
				arguments: []providerKey{},
				hasError:  true,
			},
		},
	}, providers)
}

type testTwoValidProvidersModule struct{}

func (self testTwoValidProvidersModule) Provider1() (int32, int64) {
	return 0, 0
}

func (self testTwoValidProvidersModule) Provider2(_ int32, _ int64) (int32, bool) {
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
				provider:  reflect.ValueOf(module).MethodByName("Provider1"),
				arguments: []providerKey{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(bool(false)),
			}: {
				provider: reflect.ValueOf(module).MethodByName("Provider2"),
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

func (self testValidProvidersModule1) Provider1() (int32, int64) {
	return 0, 0
}

type testValidProvidersModule2 struct{}

func (self testValidProvidersModule2) Provider2(_ int32, _ int64) (int32, bool) {
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
				provider:  reflect.ValueOf(module1).MethodByName("Provider1"),
				arguments: []providerKey{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int32(0)),
				annotationType: reflect.TypeOf(bool(false)),
			}: {
				provider: reflect.ValueOf(module2).MethodByName("Provider2"),
				arguments: []providerKey{{
					valueType:      reflect.TypeOf(int32(0)),
					annotationType: reflect.TypeOf(int64(0)),
				}},
				hasError: false,
			},
		},
	}, providers)
}

func TestBuildProviders(t *testing.T) {
	suite.Run(t, new(BuildProvidersTests))
}
