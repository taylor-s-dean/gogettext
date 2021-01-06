package po2json

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	poFilePath = "../testdata/test.po"
)

type TestSuite struct {
	suite.Suite
}

func TestPO2JSON(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (t *TestSuite) TestLoadFile_Valid() {
	loader := Loader{}
	poJSON, err := loader.LoadFile(poFilePath)
	t.NoError(err)

	enc := json.NewEncoder(ioutil.Discard)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	t.NoError(enc.Encode(poJSON))
}

func (t *TestSuite) TestLoadString_Valid() {
	fileContents, err := ioutil.ReadFile(poFilePath)
	t.NoError(err)

	loader := Loader{}
	poJSON, err := loader.LoadString(string(fileContents))
	t.NoError(err)

	enc := json.NewEncoder(ioutil.Discard)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	t.NoError(enc.Encode(poJSON))
}

func (t *TestSuite) TestLoadString_DuplicateMsgid() {
	loader := Loader{}
	_, err := loader.LoadString(`
msgid "Log in"
msgstr "Войти"

msgid "Log in"
msgstr "Войти"
`)
	t.Error(err)
}

func (t *TestSuite) TestLoadString_MissingMsgid0() {
	loader := Loader{}
	_, err := loader.LoadString(`
msgstr "Войти"
`)
	t.Error(err)
}

func (t *TestSuite) TestLoadString_MissingMsgid1() {
	loader := Loader{}
	_, err := loader.LoadString(`
msgctxt "Dialog title"
msgstr "Войти"
`)
	t.Error(err)
}

func (t *TestSuite) TestLoadString_DuplicateMsgctxt() {
	loader := Loader{}
	_, err := loader.LoadString(`
msgctxt "Dialog title"
msgctxt "Dialog title"
`)
	t.Error(err)
}

func BenchmarkLoadBytes(b *testing.B) {
	fileContents, err := ioutil.ReadFile(poFilePath)
	if err != nil {
		b.FailNow()
	}

	loader := Loader{}
	b.ResetTimer()

	for i := 0; i < 100; i++ {
		loader.LoadBytes(fileContents)
	}
}
