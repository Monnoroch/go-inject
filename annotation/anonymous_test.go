package annotation

import (
	"testing"

	"github.com/monnoroch/go-inject"
	"github.com/monnoroch/go-inject/rewrite"
	"github.com/stretchr/testify/require"
)

func TestNextAnonimousTypeNotEquals(t *testing.T) {
	require.NotEqual(t, NextAnonimousAnnotatation(), NextAnonimousAnnotatation())
}

type private struct{}

type testDynamicModule struct {
	value string
}

func (self testDynamicModule) ProvideValue() (string, private) {
	return self.value, private{}
}

func TestNextAnonimousTypeWithInjector(t *testing.T) {
	testValue1 := "test_1"
	testValue2 := "test_2"

	testAnnotation1 := NextAnonimousAnnotatation()
	testAnnotation2 := NextAnonimousAnnotatation()
	require.NotEqual(t, testAnnotation1, testAnnotation2)
	injector, err := inject.InjectorOf(
		rewrite.RewriteAnnotations(testDynamicModule{value: testValue1}, rewrite.AnnotationsMapping{
			private{}: testAnnotation1,
		}),
		rewrite.RewriteAnnotations(testDynamicModule{value: testValue2}, rewrite.AnnotationsMapping{
			private{}: testAnnotation2,
		}),
	)
	require.Nil(t, err)

	require.Equal(t, testValue1, injector.MustGet(new(string), testAnnotation1))
	require.Equal(t, testValue2, injector.MustGet(new(string), testAnnotation2))
}
