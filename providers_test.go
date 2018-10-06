package inject

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testAnnotation1 struct{}
type testAnnotation2 struct{}

type BuildProvidersTests struct {
	suite.Suite
}

type testModuleWithProviders struct {
	providers []Provider
}

func (self testModuleWithProviders) Providers() ([]Provider, error) {
	return self.providers, nil
}

func (self *BuildProvidersTests) TestEmptyModule() {
	providers := self.buildProviders(testModuleWithProviders{[]Provider{}})
	self.Equal(map[providerKey]providerData{}, providers)
}

func (self *BuildProvidersTests) TestNoArgumentsProvider() {
	function := func() (int, testAnnotation1) {
		return 0, testAnnotation1{}
	}
	providers := self.buildProviders(testModuleWithProviders{[]Provider{NewProvider(function)}})
	self.Equal(map[providerKey]providerData{
		{
			valueType:      reflect.TypeOf(int(0)),
			annotationType: reflect.TypeOf(testAnnotation1{}),
		}: {
			provider:  reflect.ValueOf(function),
			arguments: []providerKey{},
		},
	}, providers)
}

func (self *BuildProvidersTests) TestReflectValueProvider() {
	function := func() (int, testAnnotation1) {
		return 0, testAnnotation1{}
	}
	providers := self.buildProviders(testModuleWithProviders{[]Provider{NewProvider(reflect.ValueOf(function))}})
	self.Equal(map[providerKey]providerData{
		{
			valueType:      reflect.TypeOf(int(0)),
			annotationType: reflect.TypeOf(testAnnotation1{}),
		}: {
			provider:  reflect.ValueOf(function),
			arguments: []providerKey{},
		},
	}, providers)
}

type testStaticModule struct{}

func (self testStaticModule) Provide() (int, testAnnotation1) {
	return 0, testAnnotation1{}
}

func (self *BuildProvidersTests) TestStaticModule() {
	module := testStaticModule{}
	providers := self.buildProviders(module)
	self.Equal(map[providerKey]providerData{
		{
			valueType:      reflect.TypeOf(int(0)),
			annotationType: reflect.TypeOf(testAnnotation1{}),
		}: {
			provider:  reflect.ValueOf(module).MethodByName("Provide"),
			arguments: []providerKey{},
		},
	}, providers)
}

func (self *BuildProvidersTests) TestProviderWithArguments() {
	function := func(_ bool, _ testAnnotation2) (int, testAnnotation1) {
		return 0, testAnnotation1{}
	}
	providers := self.buildProviders(testModuleWithProviders{[]Provider{NewProvider(function)}})
	self.Equal(map[providerKey]providerData{
		{
			valueType:      reflect.TypeOf(int(0)),
			annotationType: reflect.TypeOf(testAnnotation1{}),
		}: {
			provider: reflect.ValueOf(function),
			arguments: []providerKey{{
				valueType:      reflect.TypeOf(bool(false)),
				annotationType: reflect.TypeOf(testAnnotation2{}),
			}},
		},
	}, providers)
}

func (self *BuildProvidersTests) TestProviderWithError() {
	function := func() (int, testAnnotation1, error) {
		return 0, testAnnotation1{}, nil
	}
	providers := self.buildProviders(testModuleWithProviders{[]Provider{NewProvider(function)}})
	self.Equal(map[providerKey]providerData{
		{
			valueType:      reflect.TypeOf(int(0)),
			annotationType: reflect.TypeOf(testAnnotation1{}),
		}: {
			provider:  reflect.ValueOf(function),
			arguments: []providerKey{},
			hasError:  true,
		},
	}, providers)
}

func (self *BuildProvidersTests) TestCachedProvider() {
	function := func() (int, testAnnotation1) {
		return 0, testAnnotation1{}
	}
	providers := self.buildProviders(testModuleWithProviders{[]Provider{NewProvider(function).Cached(true)}})
	self.Equal(map[providerKey]providerData{
		{
			valueType:      reflect.TypeOf(int(0)),
			annotationType: reflect.TypeOf(testAnnotation1{}),
		}: {
			provider:  reflect.ValueOf(function),
			arguments: []providerKey{},
			cached:    true,
		},
	}, providers)
}

func (self *BuildProvidersTests) TestMultipleModules() {
	function1 := func() (int, testAnnotation1) {
		return 0, testAnnotation1{}
	}
	function2 := func() (int, testAnnotation2) {
		return 0, testAnnotation2{}
	}
	providers := self.buildProviders(CombineModules(
		testModuleWithProviders{[]Provider{NewProvider(function1)}},
		testModuleWithProviders{[]Provider{NewProvider(function2)}},
	))
	self.Equal(map[providerKey]providerData{
		{
			valueType:      reflect.TypeOf(int(0)),
			annotationType: reflect.TypeOf(testAnnotation1{}),
		}: {
			provider:  reflect.ValueOf(function1),
			arguments: []providerKey{},
		},
		{
			valueType:      reflect.TypeOf(int(0)),
			annotationType: reflect.TypeOf(testAnnotation2{}),
		}: {
			provider:  reflect.ValueOf(function2),
			arguments: []providerKey{},
		},
	}, providers)
}

func (self *BuildProvidersTests) TestDuplicatedProvider() {
	function := func() (int, testAnnotation1) {
		return 0, testAnnotation1{}
	}
	providers := self.buildProviders(testModuleWithProviders{[]Provider{
		NewProvider(function),
		NewProvider(function),
	}})
	self.Equal(map[providerKey]providerData{
		{
			valueType:      reflect.TypeOf(int(0)),
			annotationType: reflect.TypeOf(testAnnotation1{}),
		}: {
			provider:  reflect.ValueOf(function),
			arguments: []providerKey{},
		},
	}, providers)
}

func (self *BuildProvidersTests) TestDuplicatedProviders() {
	err := self.buildProvidersError(testModuleWithProviders{[]Provider{
		NewProvider(func() (int, testAnnotation1) {
			return 0, testAnnotation1{}
		}), NewProvider(func() (int, testAnnotation1) {
			return 0, testAnnotation1{}
		}),
	}})
	self.Contains(err.Error(), "Duplicate providers for key")
}

func (self *BuildProvidersTests) TestDuplicatedProvidersAcrossModules() {
	err := self.buildProvidersError(CombineModules(
		testModuleWithProviders{[]Provider{NewProvider(func() (int, testAnnotation1) {
			return 0, testAnnotation1{}
		})}},
		testModuleWithProviders{[]Provider{NewProvider(func() (int, testAnnotation1) {
			return 0, testAnnotation1{}
		})}},
	))
	self.Contains(err.Error(), "Duplicate providers for key")
}

func (self *BuildProvidersTests) TestInvalidProvider() {
	err := self.buildProvidersError(testModuleWithProviders{[]Provider{NewProvider(0)}})
	self.Contains(err.Error(), "invalid provider")
}

type testErrorModule struct {
	err error
}

func (self testErrorModule) Providers() ([]Provider, error) {
	return nil, self.err
}

func (self *BuildProvidersTests) TestProvidersError() {
	testError := errors.New("test error")
	err := self.buildProvidersError(testErrorModule{testError})
	self.Equal(testError, err)
}

func (self *BuildProvidersTests) buildProviders(module Module) map[providerKey]providerData {
	providers, err := buildProviders(module)
	self.Require().Nil(err)
	return providers.providers
}

func (self *BuildProvidersTests) buildProvidersError(module Module) error {
	_, err := buildProviders(module)
	self.Require().NotNil(err)
	return err
}

func TestBuildProviders(t *testing.T) {
	suite.Run(t, new(BuildProvidersTests))
}
