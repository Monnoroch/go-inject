package inject

// A module that is a collection of modules should implement this interface.
type moduleCollection interface {
	Module
	// Get a list of modules in the collection.
	Modules() []Module
}

func flattenModule(module Module) []Module {
	actualModules := make([]Module, 0, 1)
	return flattenModuleFill(module, actualModules)
}

func flattenModuleFill(module Module, output []Module) []Module {
	if combined, ok := module.(moduleCollection); ok {
		output = flattenModulesFill(combined.Modules(), output)
	} else {
		output = append(output, module)
	}
	return output
}

func flattenModulesFill(modules []Module, output []Module) []Module {
	for _, module := range modules {
		output = flattenModuleFill(module, output)
	}
	return output
}
