package gogettext

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	poFilePath = "testdata/test.po"
)

type TestSuite struct {
	suite.Suite
	mc *MessageCatalog
}

func TestGettext(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (t *TestSuite) SetupSuite() {
	var err error
	t.mc, err = NewMessageCatalogFromFile(poFilePath)
	t.NoError(err)
	t.NotNil(t.mc)
}

func (t *TestSuite) TestNewMessageCatalogFromFile() {
	mc, err := NewMessageCatalogFromFile(poFilePath)
	t.NoError(err)
	t.NotNil(mc)
}

func (t *TestSuite) TestNewMessageCatalogFromString() {
	fileContents, err := ioutil.ReadFile(poFilePath)
	t.NoError(err)
	mc, err := NewMessageCatalogFromString(string(fileContents))
	t.NoError(err)
	t.NotNil(mc)
}

func (t *TestSuite) TestNewMessageCatalogFromBytes() {
	fileContents, err := ioutil.ReadFile(poFilePath)
	t.NoError(err)
	mc, err := NewMessageCatalogFromBytes(fileContents)
	t.NoError(err)
	t.NotNil(mc)
}

func (t *TestSuite) TestMessageCatalog_Gettext() {
	messages, err := t.mc.GetMessages()
	t.NoError(err)
	t.NotNil(messages)
}
