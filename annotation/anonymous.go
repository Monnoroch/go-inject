package annotation

import (
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/monnoroch/go-inject"
)

// Private annotation and must not have collision in other packages.
// Must not be used directly, only for anonymous type generation.
type anonimousTypeTag struct{}

var anonimousTypeId int64 = 0

/// NextAnonimousAnnotatation generates an unique annotation,
///
/// This is used to override annotations to avoid duplication of providers.
/// For example, some module extends the another provider of another module, performing useful actions,
/// but keeping the provider signature. It is also necessary to allow to create several instances of this module
/// in the one injector. Here we use an anonymous annotation to override the annotation of the extensible provider.
///
/// Example:
///     type server struct {}
///     type serverModule struct {}
///     func (_ serverModule) ProvideEndpoint() (string, server) {
///         return "server.com", server{}
///     }
///
///     type readyServer struct {}
///     type waitModule struct {}
///     func (_ waitModule) ProvideValue(endpoint string, _ server) (string, readyServer, error) {
///         return endpoint, readyServer{}, waitEndpoint(endpoint)
///     }
///     func ModuleWithReadyServer(annotation inject.Annotation) inject.Module {
///         privateServerAnnotation := hackannotation.NextAnonimousAnnotatation()
///         return inject.CombineModules(
///             rewrite.RewriteAnnotations(serverModule{}, map[inject.Annotation]inject.Annotation{
///                 server{}: privateServerAnnotation,
///             }),
///             rewrite.RewriteAnnotations(waitModule{}, map[inject.Annotation]inject.Annotation{
///                 server{}: privateServerAnnotation,
///                 readyServer{}: annotation,
///             }),
///     }
///
///     type server1 struct {}
///     type server2 struct {}
///     injector, err := inject.InjectorOf(
///         ModuleWithReadyServer(server1{}),
///         ModuleWithReadyServer(server2{}),
///     )
///     endpoint1 := injector.MustGet(new(string), server1{}).(string)
///     endpoint2 := injector.MustGet(new(string), server2{}).(string)
func NextAnonimousAnnotatation() inject.Annotation {
	tag := atomic.AddInt64(&anonimousTypeId, 1)
	annotationType := reflect.StructOf([]reflect.StructField{{
		Name: fmt.Sprintf("Tag%d", tag),
		Type: reflect.TypeOf(anonimousTypeTag{}),
	}})
	return reflect.New(annotationType).Elem().Interface()
}
