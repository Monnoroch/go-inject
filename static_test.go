package inject

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type StaticProvidersTests struct {
	suite.Suite
}

func (self *StaticProvidersTests) TestEmptyModule() {
	type emptyModule struct{}
	providers := self.getProviders(emptyModule{})
	self.Equal([]Provider{}, providers)
}

type testProviderModule struct{}

func (self testProviderModule) Provide() (int32, int64) {
	return 0, 0
}

func (self *StaticProvidersTests) TestProvider() {
	module := testProviderModule{}
	providers := self.getProviders(module)
	self.Equal([]Provider{{
		Function: reflect.ValueOf(module).MethodByName("Provide"),
		Cached:   false,
	}}, providers)
}

type testCachedProviderModule struct{}

func (self testCachedProviderModule) ProvideCached() (int32, int64) {
	return 0, 0
}

func (self *StaticProvidersTests) TestCachedProvider() {
	module := testCachedProviderModule{}
	providers := self.getProviders(module)
	self.Equal([]Provider{{
		Function: reflect.ValueOf(module).MethodByName("ProvideCached"),
		Cached:   true,
	}}, providers)
}

type testBadMethodNameModule struct{}

func (self testBadMethodNameModule) NotAProvider() {}

func (self *StaticProvidersTests) TestBadMethodName() {
	_, err := staticProvidersModule{testBadMethodNameModule{}}.Providers()
	self.Contains(err.Error(), "not a module")
}

type testNoResultMethodModule struct{}

func (self testNoResultMethodModule) ProvideInvalid() {}

func (self *StaticProvidersTests) TestNoResultMethod() {
	_, err := staticProvidersModule{testNoResultMethodModule{}}.Providers()
	self.Contains(err.Error(), "not a module")
}

type testNoAnnotationMethodModule struct{}

func (self testNoAnnotationMethodModule) ProvideInvalid() int {
	return 0
}

func (self *StaticProvidersTests) TestNoAnnotationMethod() {
	_, err := staticProvidersModule{testNoAnnotationMethodModule{}}.Providers()
	self.Contains(err.Error(), "not a module")
}

type testInvalidErrorTypeMethodModule struct{}

func (self testInvalidErrorTypeMethodModule) ProvideInvalid() (int, int, int) {
	return 0, 0, 0
}

func (self *StaticProvidersTests) TestInvalidErrorTypeMethod() {
	_, err := staticProvidersModule{testInvalidErrorTypeMethodModule{}}.Providers()
	self.Contains(err.Error(), "not a module")
}

type testArgumentWithoutAnnotationMethodModule struct{}

func (self testArgumentWithoutAnnotationMethodModule) ProvideInvalid(_ int) (int, int) {
	return 0, 0
}

func (self *StaticProvidersTests) TestArgumentWithoutAnnotationMethod() {
	_, err := staticProvidersModule{testArgumentWithoutAnnotationMethodModule{}}.Providers()
	self.Contains(err.Error(), "not a module")
}

type testSameProviderTwiceModule struct{}

func (self testSameProviderTwiceModule) Provide1() (int, int) {
	return 0, 0
}

func (self testSameProviderTwiceModule) Provide2() (int, int) {
	return 0, 0
}

func (self *StaticProvidersTests) TestSameProviderTwice() {
	_, err := staticProvidersModule{testSameProviderTwiceModule{}}.Providers()
	self.Contains(err.Error(), "Duplicate providers for key")
}

func (self *StaticProvidersTests) getProviders(module Module) []Provider {
	providers, err := staticProvidersModule{module}.Providers()
	self.Require().Nil(err)
	return providers
}

func TestStaticProviders(t *testing.T) {
	suite.Run(t, new(StaticProvidersTests))
}
