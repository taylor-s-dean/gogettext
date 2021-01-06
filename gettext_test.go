package gogettext

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	poFilePath = "testdata/test.po"
)

type TestSuite struct {
	suite.Suite
}

func TestGettext(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (t *TestSuite) TestNewMessageCatalogFromFile() {
	mc, err := NewMessageCatalogFromFile(poFilePath)
	t.NoError(err)
	t.NotNil(mc)
}
