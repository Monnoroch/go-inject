package inject

// CheckModule checks if the object is in fact a Module.
// Handles collections of modules correctly.
func CheckModule(module Module) error {
	_, err := buildProviders(module)
	return err
}
