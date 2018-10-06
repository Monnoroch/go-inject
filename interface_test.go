package inject

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProviderTests struct {
	suite.Suite
}

func (self *ProviderTests) TestNewProvider() {
	function := func() (int, testAnnotation1) {
		return 0, testAnnotation1{}
	}
	provider := NewProvider(function)
	self.Equal(reflect.ValueOf(function), provider.Function())
	self.False(provider.IsCached())
	self.True(provider.IsValid())
}

func (self *ProviderTests) TestNewReflectValueProvider() {
	function := reflect.ValueOf(func() (int, testAnnotation1) {
		return 0, testAnnotation1{}
	})
	provider := NewProvider(function)
	self.Equal(function, provider.Function())
	self.False(provider.IsCached())
	self.True(provider.IsValid())
}

func (self *ProviderTests) TestCachedProvider() {
	function := func() (int, testAnnotation1) {
		return 0, testAnnotation1{}
	}
	provider := NewProvider(function).Cached(true)
	self.True(provider.IsCached())
	self.True(provider.IsValid())
}

func (self *ProviderTests) TestNotCachedProvider() {
	function := func() (int, testAnnotation1) {
		return 0, testAnnotation1{}
	}
	provider := NewProvider(function).Cached(true).Cached(false)
	self.False(provider.IsCached())
	self.True(provider.IsValid())
}

func (self *ProviderTests) TestNotAFunction() {
	self.False(NewProvider(0).IsValid())
}

func (self *ProviderTests) TestNoResultFunction() {
	self.False(NewProvider(func() {}).IsValid())
}

func (self *ProviderTests) TestNoAnnotationFunction() {
	self.False(NewProvider(func() int {
		return 0
	}).IsValid())
}

func (self *ProviderTests) TestInvalidErrorTypeFunction() {
	self.False(NewProvider(func() (int, testAnnotation1, int) {
		return 0, testAnnotation1{}, 0
	}).IsValid())
}

func (self *ProviderTests) TestArgumentWithoutAnnotationFunction() {
	self.False(NewProvider(func(_ int) (int, testAnnotation1) {
		return 0, testAnnotation1{}
	}).IsValid())
}

func TestProvider(t *testing.T) {
	suite.Run(t, new(ProviderTests))
}

type ProvidersTests struct {
	suite.Suite
}

func (self *ProvidersTests) TestDynamicModule() {
	providers := []Provider{NewProvider(func() {})}
	actualProviders, err := Providers(testModuleWithProviders{providers})
	self.Require().Nil(err)
	self.Equal(providers, actualProviders)
}

func (self *ProvidersTests) TestStaticModule() {
	module := testStaticModule{}
	actualProviders, err := Providers(module)
	self.Require().Nil(err)
	self.Equal([]Provider{NewProvider(reflect.ValueOf(module).MethodByName("Provide"))}, actualProviders)
}

func (self *ProvidersTests) TestErrorModule() {
	testError := errors.New("test error")
	_, err := Providers(testErrorModule{testError})
	self.Equal(testError, err)
}

func TestProviders(t *testing.T) {
	suite.Run(t, new(ProvidersTests))
}
