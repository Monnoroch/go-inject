# Go-inject — Dependency Injection Library for Go

A dependency injection system for Go inspired by Guice. See [the post](https://monnoroch.github.io/posts/2018/10/27/go-inject-dependency-injection-library-for-go.html) for detailed guide.

## Installation

#### Go get

```
go get github.com/monnoroch/go-inject
```

#### Dep

```
dep ensure -add github.com/monnoroch/go-inject
```

## Examples

#### Inject an int value

Here's a simple module with a single provider:

```
import (
	"github.com/monnoroch/go-inject"
)

type singleValue struct{}

type MyModule struct{}

func (_ MyModule) ProvideValue() (int, singleValue) {
	return 10, singleValue{}
}

func main() {
	injector, _ := inject.InjectorOf(MyModule{})
	// Will be 10.
	value := injector.MustGet(new(int), singleValue{}).(int)
}
```

#### Dependencies

Let's add a second provider that depends on the first one:

```
type doubleValue struct{}

func (_ MyModule) ProvideValueDouble(value int, _ singleValue) (int, doubleValue) {
	return value * 2, doubleValue{}
}

func main() {
	injector, _ := inject.InjectorOf(MyModule{})
	// Will be 20.
	value := injector.MustGet(new(int), doubleValue{}).(int)
}
```

#### Cross-module dependencies

Let's add a second module, which can depend on values provided by the first module:

```
type tripleValue struct{}

type MyAnotherModule struct{}

func (_ MyAnotherModule) ProvideValue(
	value int, _ singleValue,
	doubledValue int, _ doubleValue,
) (int, tripleValue) {
	return doubledValue + value, tripleValue{}
}

func main() {
	injector, _ := inject.InjectorOf(
		MyModule{},
		MyAnotherModule{},
	)
	// Will be 30.
	value := injector.MustGet(new(int), tripleValue{}).(int)
}
```

#### Lazy dependencies

You can inject a function that returns a value. That way the value will be computed lazily.

```
type tripleValueLazy struct{}

func (_ MyAnotherModule) ProvideValueLazy(
	value int, _ singleValue,
	doubledValue func() int, _ doubleValue,
) (int, tripleValueLazy) {
	return doubledValue() + value, tripleValueLazy{}
}

func main() {
	injector, _ := inject.InjectorOf(
		MyModule{},
		MyAnotherModule{},
	)
	// Will be 30.
	value := injector.MustGet(new(int), tripleValueLazy{}).(int)
}
```

#### Auto injecting dependencies in struct fields

```
import (
	"github.com/monnoroch/go-inject"
	"github.com/monnoroch/go-inject/auto"
)

type MyStruct struct {
	Value        int
	DoubledValue int
}

type myAnnotation struct{}

func main() {
	injector, _ := inject.InjectorOf(
		MyModule{},
		autoinject.AutoInjectModule(new(MyStruct)).
			WithAnnotation(myAnnotation{}).
			WithFieldAnnotations(struct {
				Value        singleValue
				DoubledValue doubleValue
			}{}),
	)
	// Will be MyStruct{Value: 10, DoubledValue: 20}.
	value := injector.MustGet(new(MyStruct), myAnnotation{}).(MyStruct)
}
```

You can also inplement `autoinject.AutoInjectable` to provide default way to construct `MyStruct`:

```
func (self MyStruct) ProvideAutoInjectAnnotations() interface{} {
	return struct {
		Value        singleValue
		DoubledValue doubleValue
	}{}
}

func main() {
	injector, _ := inject.InjectorOf(
		MyModule{},
		autoinject.AutoInjectModule(new(MyStruct)).
			WithAnnotation(myAnnotation{}),
	)
	// Will be MyStruct{Value: 10, DoubledValue: 20}.
	value := injector.MustGet(new(MyStruct), myAnnotation{}).(MyStruct)
}
```

You can also use the default `autoinject.Auto` annotation to simplify code even further:

```
func main() {
	injector, _ := inject.InjectorOf(
		MyModule{},
		autoinject.AutoInjectModule(new(MyStruct)),
	)
	// Will be MyStruct{Value: 10, DoubledValue: 20}.
	value := injector.MustGet(new(MyStruct), autoinject.Auto{}).(MyStruct)
}
```

#### Other features

Providing singletons, private providers, annotation rewriting, dynamic modules and other features are explained in more detail in the [guide](https://monnoroch.github.io/posts/2018/10/27/go-inject-dependency-injection-library-for-go.html).
