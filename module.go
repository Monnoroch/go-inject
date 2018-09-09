package inject

// Module is the interface that has to be implemented by all modules.
// It is empty, so implementation is trivial.
// In addition to this interface all Modules have to have methods that have two or three outputs:
// - A value type.
// - An annotation type.
// - Optionally, an error.
// These methods can have inputs that should come in pairs: values and their annotations.
type Module interface{}
