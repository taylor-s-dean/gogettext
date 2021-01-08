package gogettext

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	poFilePath = "testdata/test.po"
)

var messagesJSON = []byte(`
{
    "": {
        "": {
            "Content-Transfer-Encoding": "8bit",
            "Content-Type": "text/plain; charset=UTF-8",
            "Language": "ru",
            "MIME-Version": "1.0",
            "Plural-Forms": "nplurals=3; plural=(n%10==1 && n%100!=11 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2);"
        },
        "%d user likes this.": {
            "plurals": [
                "one",
                "few",
                "many",
                "other"
            ]
        },
        "One piggy went to the market.": {
            "translation": "Одна свинья ушла на рынок."
        }
    },
    "Button label": {
        "Log in": {
            "translation": "Войти"
        }
    },
    "Context with plural": {
        "One piggy went to the market.": {
            "plurals": [
                "Одна свинья ушла на рынок.",
                "%d свиньи пошли на рынок.",
                "На рынок вышли %d поросят.",
                "%d поросят вышли на рынок."
            ],
            "translation": "Одна свинья ушла на рынок."
        }
    },
    "Dialog title": {
        "Log in": {
            "translation": "Вход в систему"
        }
    },
    "This is some context about the string.": {
        "Accept language %{accept_language} was rejected": {
            "translation": "Принять языки %{accept_language} были отклонены"
        }
    }
}
`)

type TestSuite struct {
	suite.Suite
	mc       *MessageCatalog
	messages map[string]interface{}
}

func (t *TestSuite) SetupSuite() {
	var err error
	t.mc, err = NewMessageCatalogFromFile(poFilePath)
	t.NoError(err)
	t.NotNil(t.mc)

	err = json.Unmarshal(messagesJSON, &t.messages)
	t.NoError(err)
	t.NotNil(t.messages)
}

func TestGettext(t *testing.T) {
	suite.Run(t, new(TestSuite))
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

func (t *TestSuite) TestNewMessageCatalogFromString_InvalidPluralForms() {
	mc, err := NewMessageCatalogFromString(`
msgid ""
msgstr ""
"Plural-Forms: nplurals=3; plural=(n!==1 ? 1 : 0);\n"
`)
	t.Error(err)
	t.Nil(mc)
}

func (t *TestSuite) TestNewMessageCatalogFromBytes() {
	fileContents, err := ioutil.ReadFile(poFilePath)
	t.NoError(err)
	mc, err := NewMessageCatalogFromBytes(fileContents)
	t.NoError(err)
	t.NotNil(mc)
}

func (t *TestSuite) TestMessageCatalog_GetMessages() {
	messages, err := t.mc.GetMessages()
	t.NoError(err)
	t.NotNil(messages)
	t.True(reflect.DeepEqual(t.messages, messages))
}

func (t *TestSuite) TestMessageCatalog_Gettext_MsgidExists() {
	msgstr := t.mc.Gettext("One piggy went to the market.")
	t.EqualValues("Одна свинья ушла на рынок.", msgstr)
}

func (t *TestSuite) TestMessageCatalog_Gettext_MsgidMissing() {
	msgid := "This msgid doesn't exist."
	msgstr, err := t.mc.TryGettext(msgid)
	t.EqualError(err, ErrorMsgidNotFound.Error())
	t.EqualValues(msgid, msgstr)
}

func (t *TestSuite) TestMessageCatalog_NGettext_One() {
	msgid := "%d user likes this."
	msgstr := t.mc.NGettext(msgid, "plural", 1)
	t.EqualValues("one", msgstr)
}

func (t *TestSuite) TestMessageCatalog_NGettext_Few() {
	msgid := "%d user likes this."
	msgstr := t.mc.NGettext(msgid, "plural", 2)
	t.EqualValues("few", msgstr)
}

func (t *TestSuite) TestMessageCatalog_NGettext_Many() {
	msgid := "%d user likes this."
	msgstr := t.mc.NGettext(msgid, "plural", 5)
	t.EqualValues("many", msgstr)
}

func (t *TestSuite) TestMessageCatalog_PGettext_Valid() {
	msgstr := t.mc.PGettext("Button label", "Log in")
	t.EqualValues("Войти", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryPGettext_MissingMsgctxt() {
	msgid := "Log in"
	msgstr, err := t.mc.TryPGettext("Butt", msgid)
	t.EqualError(err, ErrorMsgctxtNotFound.Error())
	t.EqualValues(msgid, msgstr)
}
