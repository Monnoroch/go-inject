package inject

// A special module that combines multiple modules.
type combinedModule struct {
	modules []Module
}

// Implement moduleCollection.
func (self combinedModule) Modules() []Module {
	return self.modules
}

// Combine multiple modules into one.
func CombineModules(modules ...Module) Module {
	return combinedModule{
		modules: modules,
	}
}
