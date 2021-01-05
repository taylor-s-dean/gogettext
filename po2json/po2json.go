package gogettext

import (
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

const testPOFileContents = `
msgid ""
msgstr ""
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Language: ru\n"
"Plural-Forms:  nplurals=3; plural=(n%10==1 && n%100!=11 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2);\n"

msgid "%d user likes this."
msgid_plural "%d users like this."
msgstr[0] "one"
msgstr[1] "few"
msgstr[2] "many"
msgstr[3] "other"

msgctxt "This is some context"
"about the string."
msgid "Accept language "
"%{accept_language} was rejected"
msgstr "Принять "
"языки %{accept_language} были отклонены"

msgctxt "Button label"
msgid "Log in"
msgstr "Войти"

msgctxt "Dialog title"
msgid "Log in"
msgstr "Вход в систему"

msgid "One piggy went to the market."
msgstr "Одна свинья ушла на рынок."

msgctxt "Context with plural"
msgid "One piggy went to the market."
msgstr "Одна свинья ушла на рынок."

#, fuzzy
msgctxt ""
"Context with plural"
msgid ""
"One piggy went to the market."
msgid_plural ""
"One piggy went to the market."
msgstr[0] ""
"Одна свинья ушла на рынок."
msgstr[1] ""
"%d свиньи пошли на рынок."
msgstr[2] "На рынок вышли %d поросят."
msgstr[3] "%d поросят вышли на рынок."
`

type TestSuite struct {
	suite.Suite
}

func TestGogettext(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (t *TestSuite) TestPO2JSON() {
	poJSON, err := po2json(testPOFileContents)
	t.NoError(err)

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	t.NoError(enc.Encode(poJSON))
}
