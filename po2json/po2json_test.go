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

func (t *TestSuite) TestLoadFile_InvalidFilePath() {
	poJSON, err := LoadFile("./this/doesnt/exist")
	t.EqualError(err, "open ./this/doesnt/exist: no such file or directory")
	t.Nil(poJSON)
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

func (t *TestSuite) TestLoadBytes_Valid() {
	fileContents, err := ioutil.ReadFile(poFilePath)
	t.NoError(err)

	poJSON, err := LoadBytes(fileContents)
	t.NoError(err)

	enc := json.NewEncoder(ioutil.Discard)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	t.NoError(enc.Encode(poJSON))
}

func (t *TestSuite) TestLoadBytes_MsgctxtUnexpectedState() {
	_, err := LoadBytes([]byte(`
msgid "Log in"
msgctxt "Войти"
msgstr "derp"
`))
	t.EqualError(err, "Invalid .po file. Found msgctxt, expected one of {msgid_plural, msgstr}.")
}

func (t *TestSuite) TestLoadBytes_MsgidUnexpectedState() {
	_, err := LoadBytes([]byte(`
msgid "Log in"
msgstr "Войти"
msgid "Dialog title"
msgstr "Войти"
`))
	t.EqualError(err, "Invalid .po file. Found msgid, expected one of {msgid_plural, msgstr}.")
}

func (t *TestSuite) TestLoadBytes_MsgstrUnexpectedState() {
	_, err := LoadBytes([]byte(`
msgid "Log in"
msgstr "Войти"

msgctxt "Dialog title"
msgstr "Derp"
`))
	t.EqualError(err, "Invalid .po file. Found msgstr, expected one of {msgid}.")
}

func (t *TestSuite) TestLoadBytes_MsgidPluralUnexpectedState() {
	_, err := LoadBytes([]byte(`
msgid "singular"
msgid_plural "plural"
msgid_plural "plural again"
msgstr[0] "message"
`))
	t.EqualError(err, "Invalid .po file. Found msgid_plural, expected one of {msgstr_plural}.")
}

func (t *TestSuite) TestLoadBytes_MsgstrPluralUnexpectedState() {
	_, err := LoadBytes([]byte(`
msgid "singular"
msgstr[0] "message"
`))
	t.EqualError(err, "Invalid .po file. Found msgstr_plural, expected one of {msgid_plural, msgstr}.")
}

func (t *TestSuite) TestLoadBytes_AppendAllFields() {
	_, err := LoadBytes([]byte(`
msgctxt ""
"Context"
msgid ""
"Message ID"
msgid_plural ""
"Message ID plural
msgstr[0] "Message string plural"

msgid ""
"Message ID"
msgstr ""
"Message"

""
`))
	t.EqualError(err, "Encountered invalid state. Please ensure the input file is in a valid .po format.")
}

func (t *TestSuite) TestLoadBytes_DuplicateMsgid() {
	_, err := LoadBytes([]byte(`
msgid "Log in"
msgstr "Войти"

msgid "Log in"
msgstr "Войти"
`))
	t.EqualError(err, `Invalid .po file. Found duplicate msgstr for msgid "Log in".`)
}

func (t *TestSuite) TestLoadBytes_DuplicateHeaderKey() {
	_, err := LoadBytes([]byte(`
msgid ""
msgstr ""
"HeaderKey: Header value\n"
"HeaderKey: Another value\n"
`))
	t.EqualError(err, `Invalid .po file. Found duplicate header key "HeaderKey".`)
}

func (t *TestSuite) TestLoadBytes_MissingMsgid0() {
	_, err := LoadBytes([]byte(`
msgstr "Войти"
`))
	t.EqualError(err, "Invalid .po file. Found msgstr, expected one of {msgctxt, msgid}.")
}

func (t *TestSuite) TestLoadBytes_MissingMsgid1() {
	_, err := LoadBytes([]byte(`
msgctxt "Dialog title"
msgstr "Войти"
`))
	t.EqualError(err, "Invalid .po file. Found msgstr, expected one of {msgid}.")
}

func (t *TestSuite) TestLoadBytes_DuplicateMsgctxt() {
	_, err := LoadBytes([]byte(`
msgctxt "Dialog title"
msgctxt "Dialog title"
`))
	t.EqualError(err, "Invalid .po file. Found msgctxt, expected one of {msgid}.")
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
