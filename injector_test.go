package inject

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

const testValue = 17

var testError = errors.New("test error")

type Annotation1 struct{}
type Annotation2 struct{}
type Annotation3 struct{}

type InjectorTests struct {
	suite.Suite
	injector *Injector
}

func (self *InjectorTests) TestNotFound() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{},
	})
	_, err := self.injector.Get((*int)(nil), Annotation1{})
	cause := err.(provideError).cause
	self.Require().NotNil(cause)
	self.Contains(cause.Error(), "No provider found")
}

func (self *InjectorTests) TestNotFoundTransitive() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func(_ int, _ Annotation2) (int, Annotation1) {
					return testValue, Annotation1{}
				}),
				arguments: []providerArgument{{providerKey{
					valueType:      reflect.TypeOf(int(0)),
					annotationType: reflect.TypeOf(Annotation2{}),
				}, nil}},
				hasError: false,
			},
		},
	})
	_, err := self.injector.Get((*int)(nil), Annotation2{})
	cause := err.(provideError).cause
	self.Require().NotNil(cause)
	self.Contains(cause.Error(), "No provider found")
}

func (self *InjectorTests) TestGet() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					return testValue, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
		},
	})
	value := self.getInt(Annotation1{})
	self.Equal(testValue, value)
}

func (self *InjectorTests) TestGetNil() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf((*int)(nil)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (*int, Annotation1) {
					return nil, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
		},
	})
	value := self.getIntPtr(Annotation1{})
	self.Nil(value)
}

func (self *InjectorTests) TestErrorGet() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1, error) {
					return testValue, Annotation1{}, nil
				}),
				arguments: []providerArgument{},
				hasError:  true,
			},
		},
	})
	value := self.getInt(Annotation1{})
	self.Equal(testValue, value)
}

func (self *InjectorTests) TestGetError() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1, error) {
					return testValue, Annotation1{}, testError
				}),
				arguments: []providerArgument{},
				hasError:  true,
			},
		},
	})
	_, err := self.injector.Get((*int)(nil), Annotation1{})
	self.Equal(testError, err.(provideError).cause)
}

func (self *InjectorTests) TestPanic() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					panic(testError)
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
		},
	})
	self.PanicsWithValue(testError, func() {
		self.getInt(Annotation1{})
	})
}

func (self *InjectorTests) TestCachedGet() {
	counter := testValue
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					defer func() {
						counter += 1
					}()
					return counter, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
				cached:    true,
			},
		},
	})
	self.Equal(testValue, self.getInt(Annotation1{}))
	self.Equal(testValue, self.getInt(Annotation1{}))
}

func (self *InjectorTests) TestGetTransitiveError() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1, error) {
					return testValue, Annotation1{}, testError
				}),
				arguments: []providerArgument{},
				hasError:  true,
			},
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				provider: reflect.ValueOf(func(value int, _ Annotation1) (int, Annotation2) {
					return value * 2, Annotation2{}
				}),
				arguments: []providerArgument{{providerKey{
					valueType:      reflect.TypeOf(int(0)),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, nil}},
				hasError: false,
			},
		},
	})
	_, err := self.injector.Get((*int)(nil), Annotation2{})
	self.Equal(testError, err.(provideError).cause.(provideError).cause)
}

func (self *InjectorTests) TestGetRecalculates() {
	counter := testValue
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					defer func() {
						counter += 1
					}()
					return counter, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
		},
	})
	self.Equal(testValue, self.getInt(Annotation1{}))
	self.Equal(testValue+1, self.getInt(Annotation1{}))
	self.Equal(testValue+2, self.getInt(Annotation1{}))
}

func (self *InjectorTests) TestGetWithOriginalAnnotation() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					return testValue, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				provider: reflect.ValueOf(func(value int, _ Annotation3) (int, Annotation2) {
					return value * 2, Annotation2{}
				}),
				arguments: []providerArgument{{
					key: providerKey{
						valueType:      reflect.TypeOf(int(0)),
						annotationType: reflect.TypeOf(Annotation1{}),
					},
					originalAnnotationType: reflect.TypeOf(Annotation3{}),
				}},
				hasError: false,
			},
		},
	})
	self.Equal(testValue*2, self.getInt(Annotation2{}))
}

func (self *InjectorTests) TestGetLazy() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					return testValue, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				provider: reflect.ValueOf(func(value func() int, _ Annotation1) (int, Annotation2) {
					return value(), Annotation2{}
				}),
				arguments: []providerArgument{{providerKey{
					valueType:      reflect.TypeOf(func() int { return 0 }),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, nil}},
				hasError: false,
			},
		},
	})
	value := self.getInt(Annotation2{})
	self.Equal(testValue, value)
}

func (self *InjectorTests) TestGetLazyNil() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf((*int)(nil)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (*int, Annotation1) {
					return nil, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf((*int)(nil)),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				provider: reflect.ValueOf(func(value func() *int, _ Annotation1) (*int, Annotation2) {
					return value(), Annotation2{}
				}),
				arguments: []providerArgument{{providerKey{
					valueType:      reflect.TypeOf(func() *int { return nil }),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, nil}},
				hasError: false,
			},
		},
	})
	self.Nil(self.getIntPtr(Annotation2{}))
}

func (self *InjectorTests) TestGetLazyError() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1, error) {
					return 0, Annotation1{}, testError
				}),
				arguments: []providerArgument{},
				hasError:  true,
			},
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				provider: reflect.ValueOf(func(value func() int, _ Annotation1) (int, Annotation2) {
					return value(), Annotation2{}
				}),
				arguments: []providerArgument{{providerKey{
					valueType:      reflect.TypeOf(func() int { return 0 }),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, nil}},
				hasError: false,
			},
		},
	})
	_, err := self.injector.Get(new(int), Annotation2{})
	self.Equal(testError, err.(provideError).cause.(provideError).cause)
}

func (self *InjectorTests) TestGetLazyPanic() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					panic(testError)
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				provider: reflect.ValueOf(func(value func() int, _ Annotation1) (int, Annotation2) {
					return value(), Annotation2{}
				}),
				arguments: []providerArgument{{providerKey{
					valueType:      reflect.TypeOf(func() int { return 0 }),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, nil}},
				hasError: false,
			},
		},
	})
	self.PanicsWithValue(testError, func() {
		self.injector.Get(new(int), Annotation2{})
	})
}

func (self *InjectorTests) TestGetLazyDoesNotCallProviderUntilRequested() {
	calledLazyProvider := false
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					calledLazyProvider = true
					return testValue, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				provider: reflect.ValueOf(func(value func() int, _ Annotation1) (int, Annotation2) {
					return 1, Annotation2{}
				}),
				arguments: []providerArgument{{providerKey{
					valueType:      reflect.TypeOf(func() int { return 0 }),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, nil}},
				hasError: false,
			},
		},
	})
	_ = self.getInt(Annotation2{})
	self.False(calledLazyProvider)
}

func (self *InjectorTests) TestGetLazyCached() {
	counter := testValue
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					defer func() {
						counter += 1
					}()
					return counter, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
				cached:    true,
			},
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				provider: reflect.ValueOf(func(value func() int, _ Annotation1) (int, Annotation2) {
					return value(), Annotation2{}
				}),
				arguments: []providerArgument{{providerKey{
					valueType:      reflect.TypeOf(func() int { return 0 }),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, nil}},
				hasError: false,
			},
		},
	})
	self.Equal(testValue, self.getInt(Annotation2{}))
	self.Equal(testValue, self.getInt(Annotation2{}))
}

func (self *InjectorTests) TestGetTransitive() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					return testValue, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				provider: reflect.ValueOf(func(value int, _ Annotation1) (int, Annotation2) {
					return value * 2, Annotation2{}
				}),
				arguments: []providerArgument{{providerKey{
					valueType:      reflect.TypeOf(int(0)),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, nil}},
				hasError: false,
			},
		},
	})
	value := self.getInt(Annotation2{})
	self.Equal(testValue*2, value)
}

func (self *InjectorTests) TestGetTransitiveMultiple() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					return testValue, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation2{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation2) {
					return testValue * 2, Annotation2{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation3{}),
			}: {
				provider: reflect.ValueOf(func(value1 int, _ Annotation1,
					value2 int, _ Annotation2) (int, Annotation3) {
					return value1 + value2, Annotation3{}
				}),
				arguments: []providerArgument{{providerKey{
					valueType:      reflect.TypeOf(int(0)),
					annotationType: reflect.TypeOf(Annotation1{}),
				}, nil}, {providerKey{
					valueType:      reflect.TypeOf(int(0)),
					annotationType: reflect.TypeOf(Annotation2{}),
				}, nil}},
				hasError: false,
			},
		},
	})
	value := self.getInt(Annotation3{})
	self.Equal(testValue*3, value)
}

func (self *InjectorTests) TestMustGet() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{
			{
				valueType:      reflect.TypeOf(int(0)),
				annotationType: reflect.TypeOf(Annotation1{}),
			}: {
				provider: reflect.ValueOf(func() (int, Annotation1) {
					return testValue, Annotation1{}
				}),
				arguments: []providerArgument{},
				hasError:  false,
			},
		},
	})
	value := self.injector.MustGet((*int)(nil), Annotation1{}).(int)
	self.Equal(testValue, value)
}

func (self *InjectorTests) TestMustGetPanic() {
	self.initInjector(&providersData{
		providers: map[providerKey]providerData{},
	})
	defer func() {
		err := recover()
		self.NotNil(err.(provideError).cause)
	}()
	_ = self.injector.MustGet((*int)(nil), Annotation1{}).(int)
}

func (self *InjectorTests) getInt(annotation Annotation) int {
	value, err := self.injector.Get((*int)(nil), annotation)
	self.Require().Nil(err)
	intValue, ok := value.(int)
	self.Require().True(ok)
	return intValue
}

func (self *InjectorTests) getIntPtr(annotation Annotation) *int {
	value, err := self.injector.Get((**int)(nil), annotation)
	self.Require().Nil(err)
	intPtrValue, ok := value.(*int)
	self.Require().True(ok)
	return intPtrValue
}

func (self *InjectorTests) initInjector(providers *providersData) {
	self.injector = &Injector{
		providers: providers,
		cache:     map[providerKey]valueErrorPair{},
	}
}

func TestInjector(t *testing.T) {
	suite.Run(t, new(InjectorTests))
}

type InjectorOfTests struct {
	suite.Suite
}

func (self *InjectorOfTests) TestValidModule() {
	type testEmptyModule struct{}
	_, err := InjectorOf(testEmptyModule{})
	self.Nil(err)
}

func (self *InjectorOfTests) TestCombinedModule() {
	type testEmptyModule struct{}
	_, err := InjectorOf(CombineModules(testEmptyModule{}, testEmptyModule{}))
	self.Nil(err)
}

type injectorTestInvalidModule struct{}

func (self injectorTestInvalidModule) Provide() {}

func (self *InjectorOfTests) TestInvalidModule() {
	_, err := InjectorOf(injectorTestInvalidModule{})
	self.NotNil(err)
}

func TestInjectorOf(t *testing.T) {
	suite.Run(t, new(InjectorOfTests))
}
