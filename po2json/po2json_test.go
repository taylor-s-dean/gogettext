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
	poJSON, err := LoadFile(poFilePath)
	t.NoError(err)

	enc := json.NewEncoder(ioutil.Discard)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	t.NoError(enc.Encode(poJSON))
}

func (t *TestSuite) TestLoadString_Valid() {
	fileContents, err := ioutil.ReadFile(poFilePath)
	t.NoError(err)

	poJSON, err := LoadString(string(fileContents))
	t.NoError(err)

	enc := json.NewEncoder(ioutil.Discard)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	t.NoError(enc.Encode(poJSON))
}

func (t *TestSuite) TestLoadString_MissingEmptyLine() {
	_, err := LoadString(`
msgid "Log in"
msgstr "Войти"
msgid "Dialog title"
msgstr "Войти"
`)
	t.Error(err)
}

func (t *TestSuite) TestLoadString_DuplicateMsgid() {
	_, err := LoadString(`
msgid "Log in"
msgstr "Войти"

msgid "Log in"
msgstr "Войти"
`)
	t.Error(err)
}

func (t *TestSuite) TestLoadString_MissingMsgid0() {
	_, err := LoadString(`
msgstr "Войти"
`)
	t.Error(err)
}

func (t *TestSuite) TestLoadString_MissingMsgid1() {
	_, err := LoadString(`
msgctxt "Dialog title"
msgstr "Войти"
`)
	t.Error(err)
}

func (t *TestSuite) TestLoadString_DuplicateMsgctxt() {
	_, err := LoadString(`
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

	b.ResetTimer()

	for i := 0; i < 100; i++ {
		LoadBytes(fileContents)
	}
}
