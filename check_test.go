package inject

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CheckModuleTests struct {
	suite.Suite
}

func (self *CheckModuleTests) TestEmptyModuleIsValid() {
	type testEmptyModule struct{}
	self.Nil(CheckModule(testEmptyModule{}))
}

type checkTestInvalidModule struct{}

func (self checkTestInvalidModule) Provide() {}

func (self *CheckModuleTests) TestInvalid() {
	self.NotNil(CheckModule(checkTestInvalidModule{}))
}

type checkTestValidModule struct{}

func (self checkTestValidModule) Provide() (int, int) {
	return 0, 0
}

func (self *CheckModuleTests) TestValid() {
	self.Nil(CheckModule(checkTestValidModule{}))
}

func TestCheckModule(t *testing.T) {
	suite.Run(t, new(CheckModuleTests))
}
