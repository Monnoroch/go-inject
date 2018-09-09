package inject

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type FlattenModulesTests struct {
	suite.Suite
}

func (self *FlattenModulesTests) TestLeafModule() {
	type testModule struct{}
	module := testModule{}
	self.Equal([]Module{module}, flattenModule(module))
}

func (self *FlattenModulesTests) TestModuleCollection() {
	type testModule struct {
		value string
	}
	module1 := testModule{"1"}
	module2 := testModule{"2"}
	self.Equal([]Module{module1, module2}, flattenModule(CombineModules(module1, module2)))
}

func TestFlattenModules(t *testing.T) {
	suite.Run(t, new(FlattenModulesTests))
}
