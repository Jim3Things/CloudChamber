package test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type mySuite struct {
	SuiteX
}

func (ts *mySuite) TestBogus() {
	ts.Require().Bogus()
}

func TestMySuite(t *testing.T) {
	suite.Run(t, new(mySuite))
}
