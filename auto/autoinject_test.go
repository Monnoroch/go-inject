package autoinject

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/monnoroch/go-inject"
)

type Annotation struct{}

type AutoInjectTests struct {
	suite.Suite
}

func (self *AutoInjectTests) TestEmptyStruct() {
	type Struct struct{}

	provider := self.getProvider(AutoInjectModule(new(Struct)))
	self.False(provider.IsCached())

	value, annotation := self.call(provider, []reflect.Value{})

	self.Equal(Struct{}, value)

	_, ok := annotation.(Auto)
	self.True(ok)
}

func (self *AutoInjectTests) TestStructWithFields() {
	testValue := 10

	type Struct struct {
		Value int
	}

	provider := self.getProvider(AutoInjectModule(new(Struct)))
	self.False(provider.IsCached())

	value, annotation := self.call(provider, []reflect.Value{
		reflect.ValueOf(int(testValue)), reflect.ValueOf(Auto{}),
	})

	self.Equal(Struct{
		Value: testValue,
	}, value)

	_, ok := annotation.(Auto)
	self.True(ok)
}

func (self *AutoInjectTests) TestPointerToEmptyStruct() {
	type Struct struct{}

	provider := self.getProvider(AutoInjectModule(new(*Struct)))
	value, _ := self.call(provider, []reflect.Value{})
	self.Equal(&Struct{}, value)
}

func (self *AutoInjectTests) TestPointerToStructWithFields() {
	testValue := 10

	type Struct struct {
		Value int
	}

	provider := self.getProvider(AutoInjectModule(new(*Struct)))
	value, _ := self.call(provider, []reflect.Value{
		reflect.ValueOf(int(testValue)), reflect.ValueOf(Auto{}),
	})

	self.Equal(&Struct{
		Value: testValue,
	}, value)
}

func (self *AutoInjectTests) TestCustomAnnotation() {
	type Struct struct{}

	provider := self.getProvider(AutoInjectModule(new(Struct)).WithAnnotation(Annotation{}))
	_, annotation := self.call(provider, []reflect.Value{})
	_, ok := annotation.(Annotation)
	self.True(ok)
}

func (self *AutoInjectTests) TestCustomAnnotations() {
	testValue := 10
	type Struct struct {
		Value int
	}

	provider := self.getProvider(AutoInjectModule(new(Struct)).WithFieldAnnotations(struct {
		Value Annotation
	}{}))
	value, annotation := self.call(provider, []reflect.Value{
		reflect.ValueOf(int(testValue)), reflect.ValueOf(Annotation{}),
	})

	self.Equal(Struct{
		Value: testValue,
	}, value)

	_, ok := annotation.(Auto)
	self.True(ok)
}

type AutoInjectableStruct struct {
	Value int
}

func (self AutoInjectableStruct) ProvideAutoInjectAnnotations() interface{} {
	return struct{}{}
}

func (self *AutoInjectTests) TestAutoInjectable() {
	testValue := 10

	provider := self.getProvider(AutoInjectModule(new(AutoInjectableStruct)))
	value, annotation := self.call(provider, []reflect.Value{
		reflect.ValueOf(int(testValue)), reflect.ValueOf(Auto{}),
	})

	self.Equal(AutoInjectableStruct{
		Value: testValue,
	}, value)

	_, ok := annotation.(Auto)
	self.True(ok)
}

type CustomAutoInjectableStruct struct {
	Value int
}

func (self CustomAutoInjectableStruct) ProvideAutoInjectAnnotations() interface{} {
	return struct {
		Value Annotation
	}{}
}

func (self *AutoInjectTests) TestCustomAutoInjectable() {
	testValue := 10

	provider := self.getProvider(AutoInjectModule(new(CustomAutoInjectableStruct)))
	value, annotation := self.call(provider, []reflect.Value{
		reflect.ValueOf(int(testValue)), reflect.ValueOf(Annotation{}),
	})

	self.Equal(CustomAutoInjectableStruct{
		Value: testValue,
	}, value)

	_, ok := annotation.(Auto)
	self.True(ok)
}

func (self *AutoInjectTests) TestCached() {
	provider := self.getProvider(AutoInjectModule(new(struct{})).Cached())
	self.True(provider.IsCached())
}

func (self *AutoInjectTests) TestNotCached() {
	provider := self.getProvider(AutoInjectModule(new(struct{})).Cached().NotCached())
	self.False(provider.IsCached())
}

func (self *AutoInjectTests) TestPrimitive() {
	_, err := AutoInjectModule(new(int)).Providers()
	self.Contains(err.Error(), "int is not a struct")
}

func (self *AutoInjectTests) TestDoublePointerToStruct() {
	type Struct struct{}
	_, err := AutoInjectModule(new(**Struct)).Providers()
	self.Contains(err.Error(), "*autoinject.Struct is not a struct")
}

func (self *AutoInjectTests) getProvider(module inject.DynamicModule) inject.Provider {
	providers, err := module.Providers()
	self.Require().Nil(err)
	self.Equal(1, len(providers))
	return providers[0]
}

func (self *AutoInjectTests) call(provider inject.Provider, arguments []reflect.Value) (interface{}, interface{}) {
	outputs := provider.Function().Call(arguments)
	self.Equal(2, len(outputs))

	return outputs[0].Interface(), outputs[1].Interface()
}

func TestAutoInject(t *testing.T) {
	suite.Run(t, new(AutoInjectTests))
}
