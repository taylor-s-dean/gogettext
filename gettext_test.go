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
		"#This is a message with a # sign.": {
           "translation": "#This is a translation with a # sign."
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

func (t *TestSuite) TestNewMessageCatalogFromFile_Valid() {
	mc, err := NewMessageCatalogFromFile(poFilePath)
	t.NoError(err)
	t.NotNil(mc)
}

func (t *TestSuite) TestNewMessageCatalogFromFile_FileNotFound() {
	mc, err := NewMessageCatalogFromFile("./not-a-real-file.po")
	t.Error(err)
	t.Nil(mc)
}

func (t *TestSuite) TestNewMessageCatalogFromString_Valid() {
	fileContents, err := ioutil.ReadFile(poFilePath)
	t.NoError(err)
	mc, err := NewMessageCatalogFromString(string(fileContents))
	t.NoError(err)
	t.NotNil(mc)
}

func (t *TestSuite) TestNewMessageCatalogFromString_InvalidString() {
	mc, err := NewMessageCatalogFromString(`
msgid ""
msgid ""
`)
	t.Error(err)
	t.Nil(mc)
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

func (t *TestSuite) TestNewMessageCatalogFromBytes_Valid() {
	fileContents, err := ioutil.ReadFile(poFilePath)
	t.NoError(err)
	mc, err := NewMessageCatalogFromBytes(fileContents)
	t.NoError(err)
	t.NotNil(mc)
}

func (t *TestSuite) TestNewMessageCatalogFromString_InvalidBytes() {
	mc, err := NewMessageCatalogFromBytes([]byte(`
msgid ""
msgid ""
`))
	t.Error(err)
	t.Nil(mc)
}

func (t *TestSuite) TestNewMessageCatalogFromBytes_InvalidPluralForms() {
	mc, err := NewMessageCatalogFromBytes([]byte(`
msgid ""
msgstr ""
"Plural-Forms: nplurals=3; plural=(n!==1 ? 1 : 0);\n"
`))
	t.Error(err)
	t.Nil(mc)
}

func (t *TestSuite) TestMessageCatalog_GetMessages_Valid() {
	messages, err := t.mc.GetMessages()
	t.NoError(err)
	t.NotNil(messages)
	t.True(reflect.DeepEqual(t.messages, messages))
}

func (t *TestSuite) TestMessageCatalog_setPluralForms_Valid() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`
{
	"": {
		"": {
			"Plural-Forms": "nplurals=2; plural=(n==1 || n==11 ? 0 : 1);"
		}
	}
}`), &mc.messages)
	t.NoError(err)
	t.NotNil(mc)
	err = mc.setPluralForms()
	t.NoError(err)
	t.Equal("(n==1 || n==11 ? 0 : 1)", mc.pluralForms)
}

func (t *TestSuite) TestMessageCatalog_setPluralForms_NilMessageCatalog() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = nil
	err = mc.setPluralForms()
	t.EqualError(err, ErrorNilMessageCatalog.Error())
}

func (t *TestSuite) TestMessageCatalog_setPluralForms_NoPluralForms() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"": {"": {"": ""}}}`), &mc.messages)
	t.NoError(err)
	t.NotNil(mc)
	err = mc.setPluralForms()
	t.NoError(err)
	t.Equal(mc.pluralForms, defaultPluralForms)
}

func (t *TestSuite) TestMessageCatalog_setPluralForms_PluralsTypeAssertionFailed() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"": {"": {"Plural-Forms": 2}}}`), &mc.messages)
	t.NoError(err)
	t.NotNil(mc)
	err = mc.setPluralForms()
	t.EqualError(err, ErrorPluralsTypeAssertionFailed.Error())
}

func (t *TestSuite) TestMessageCatalog_setPluralForms_EmptyPluralFormsValue() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"": {"": {"Plural-Forms": ""}}}`), &mc.messages)
	t.NoError(err)
	t.NotNil(mc)
	err = mc.setPluralForms()
	t.Equal(mc.pluralForms, defaultPluralForms)
}

func (t *TestSuite) TestMessageCatalog_setPluralForms_InvalidPluralForms() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"": {"": {"Plural-Forms": "nplurals=2; plural=());"}}}`), &mc.messages)
	t.NoError(err)
	t.NotNil(mc)
	err = mc.setPluralForms()
	t.Error(err)
}
func (t *TestSuite) TestMessageCatalog_getMsgidMap_Valid() {
	msgidMap, err := t.mc.getMsgidMap("", "")
	t.NoError(err)
	t.NotNil(msgidMap)
	t.True(reflect.DeepEqual(t.messages[""].(map[string]interface{})[""].(map[string]interface{}), msgidMap))
}

func (t *TestSuite) TestMessageCatalog_getMsgidMap_NilMessageCatalog() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = nil
	msgidMap, err := mc.getMsgidMap("", "")
	t.EqualError(err, ErrorNilMessageCatalog.Error())
	t.Nil(msgidMap)
}

func (t *TestSuite) TestMessageCatalog_getMsgidMap_MsgctxtNotFound() {
	msgidMap, err := t.mc.getMsgidMap("bob", "")
	t.EqualError(err, ErrorMsgctxtNotFound.Error())
	t.Nil(msgidMap)
}

func (t *TestSuite) TestMessageCatalog_getMsgidMap_MsgidNotFound() {
	msgidMap, err := t.mc.getMsgidMap("", "bob")
	t.EqualError(err, ErrorMsgidNotFound.Error())
	t.Nil(msgidMap)
}

func (t *TestSuite) TestMessageCatalog_getMsgidMap_MsgctxtTypeAssertionFailed() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"":""}`), &mc.messages)
	t.NoError(err)
	msgidMap, err := mc.getMsgidMap("", "")
	t.EqualError(err, ErrorMsgctxtTypeAssertionFailed.Error())
	t.Nil(msgidMap)
}

func (t *TestSuite) TestMessageCatalog_getMsgidMap_MsgidTypeAssertionFailed() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"":{"":""}}`), &mc.messages)
	t.NoError(err)
	msgidMap, err := mc.getMsgidMap("", "")
	t.EqualError(err, ErrorMsgidTypeAssertionFailed.Error())
	t.Nil(msgidMap)
}

func (t *TestSuite) TestMessageCatalog_Gettext_Valid() {
	msgstr := t.mc.Gettext("One piggy went to the market.")
	t.Equal("Одна свинья ушла на рынок.", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryGettext_MsgidNotFound() {
	msgid := "This msgid doesn't exist."
	msgstr, err := t.mc.TryGettext(msgid)
	t.EqualError(err, ErrorMsgidNotFound.Error())
	t.Equal(msgid, msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryGettext_TranslationNotFound() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"":{"":{"":""}}}`), &mc.messages)
	t.NoError(err)
	msgstr, err := mc.TryGettext("")
	t.EqualError(err, ErrorTranslationNotFound.Error())
	t.Equal("", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryGettext_TranslationTypeAssertionFailed() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"":{"":{"translation":2}}}`), &mc.messages)
	t.NoError(err)
	msgstr, err := mc.TryGettext("")
	t.EqualError(err, ErrorTranslationTypeAssertionFailed.Error())
	t.Equal("", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryGettext_WithStringEscapes() {
	mc, err := NewMessageCatalogFromString(`
msgid "test\"with quotes\"\nand a newline"
msgstr "This is a \"quoted\" string with a\nnewline."
`)
	t.NoError(err)
	msgstr, err := mc.TryGettext("test\"with quotes\"\nand a newline")
	t.NoError(err)
	t.Equal("This is a \"quoted\" string with a\nnewline.", msgstr)
}

func (t *TestSuite) TestMessageCatalog_NGettext_Valid() {
	msgid := "%d user likes this."
	msgstr := t.mc.NGettext(msgid, "plural", 2)
	t.Equal("few", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryNGettext_InvalidPluralForms() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"":{"":{"Plural-Forms": "nplurals=1; plural=(();"}}}`), &mc.messages)
	t.NoError(err)
	msgstr, err := mc.TryNGettext("singular", "plural", 1)
	t.Error(err)
	t.Equal("singular", msgstr)
	msgstr, err = mc.TryNGettext("singular", "plural", 2)
	t.Error(err)
	t.Equal("plural", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryNGettext_PluralNotFound() {
	msgid := "One piggy went to the market."
	msgstr, err := t.mc.TryNGettext(msgid, "plural", 1)
	t.EqualError(err, ErrorPluralNotFound.Error())
	t.Equal(msgid, msgstr)
	msgstr, err = t.mc.TryNGettext(msgid, "plural", 2)
	t.EqualError(err, ErrorPluralNotFound.Error())
	t.Equal("plural", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryNGettext_PluralsTypeAssertionFailed() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"":{"singular":{"plurals":1}}}`), &mc.messages)
	t.NoError(err)
	msgstr, err := mc.TryNGettext("singular", "plural", 1)
	t.EqualError(err, ErrorPluralsTypeAssertionFailed.Error())
	t.Equal("singular", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryNGettext_PluralIndexOutOfBounds() {
	mc, err := NewMessageCatalogFromString(`
msgid ""
msgstr ""
"Plural-Forms: nplurals=2; plural=(n==1 ? 0 : 1)"

msgid "singular"
msgid_plural "plural"
msgstr[0] "zero"
`)
	t.NoError(err)
	t.NotNil(mc)
	msgstr, err := mc.TryNGettext("singular", "plural", 1)
	t.NoError(err)
	t.Equal("zero", msgstr)
	msgstr, err = mc.TryNGettext("singular", "plural", 2)
	t.EqualError(err, ErrorPluralsIndexOutOfBounds.Error())
	t.Equal("plural", msgstr)
}

func (t *TestSuite) TestMessageCatalog_PGettext_Valid() {
	msgstr := t.mc.PGettext("Button label", "Log in")
	t.Equal("Войти", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryPGettext_MsgctxtNotFound() {
	msgid := "Log in"
	msgstr, err := t.mc.TryPGettext("Butt", msgid)
	t.EqualError(err, ErrorMsgctxtNotFound.Error())
	t.Equal(msgid, msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryPGettext_TranslationNotFound() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"":{"test":{}}}`), &mc.messages)
	t.NoError(err)
	msgstr, err := mc.TryPGettext("", "test")
	t.EqualError(err, ErrorTranslationNotFound.Error())
	t.Equal("test", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryPGettext_TranslationTypeAssertionFailed() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{}
	err = json.Unmarshal([]byte(`{"":{"test":{"translation":1}}}`), &mc.messages)
	t.NoError(err)
	msgstr, err := mc.TryPGettext("", "test")
	t.EqualError(err, ErrorTranslationTypeAssertionFailed.Error())
	t.Equal("test", msgstr)
}

func (t *TestSuite) TestMessageCatalog_NPGettext_Valid_One() {
	msgstr := t.mc.NPGettext("Context with plural", "One piggy went to the market.", "", 1)
	t.Equal("Одна свинья ушла на рынок.", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryNPGettext_Valid_One() {
	msgstr, err := t.mc.TryNPGettext("Context with plural", "One piggy went to the market.", "", 1)
	t.NoError(err)
	t.Equal("Одна свинья ушла на рынок.", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryNPGettext_Valid_Few() {
	msgstr, err := t.mc.TryNPGettext("Context with plural", "One piggy went to the market.", "", 2)
	t.NoError(err)
	t.Equal("%d свиньи пошли на рынок.", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryNPGettext_Valid_Many() {
	msgstr, err := t.mc.TryNPGettext("Context with plural", "One piggy went to the market.", "", 5)
	t.NoError(err)
	t.Equal("На рынок вышли %d поросят.", msgstr)
}

func (t *TestSuite) TestMessageCatalog_TryNPGettext_Singular_MsgctxtNotFound() {
	msgstr, err := t.mc.TryNPGettext("this doesnt exist", "singular", "plural", 1)
	t.EqualError(err, ErrorMsgctxtNotFound.Error())
	t.Equal(msgstr, "singular")
}

func (t *TestSuite) TestMessageCatalog_TryNPGettext_Plural_MsgidNotFound() {
	msgstr, err := t.mc.TryNPGettext("Context with plural", "singular", "plural", 2)
	t.EqualError(err, ErrorMsgidNotFound.Error())
	t.Equal(msgstr, "plural")
}

func (t *TestSuite) TestMessageCatalog_SearchMsgids_Valid() {
	mc, err := NewMessageCatalogFromBytes([]byte(`
msgid "braze.1234.name"
msgstr "name"

msgctxt "context"
msgid "braze.1234.address"
msgstr "address"

msgid "braze.1235.age"
msgstr "age"

msgctxt "more context"
msgid "braze.1235.place-of-birth"
msgstr "place of birth"
`))
	t.NoError(err)
	t.NotNil(mc)
	results, err := mc.SearchMsgids(`braze\.1234\.[a-zA-Z0-9_-]`)
	t.NoError(err)
	t.ElementsMatch(results, []SearchResults{
		{Msgctxt: "", Msgid: "braze.1234.name"},
		{Msgctxt: "context", Msgid: "braze.1234.address"},
	})
}

func (t *TestSuite) TestMessageCatalog_SearchMsgids_MsgctxtTypeAssertionFailed() {
	mc, err := NewMessageCatalogFromBytes([]byte(""))
	t.NoError(err)
	t.NotNil(mc)
	mc.messages = map[string]interface{}{"": ""}
	results, err := mc.SearchMsgids(`braze\.1234\.[a-zA-Z0-9_-]`)
	t.EqualError(err, ErrorMsgctxtTypeAssertionFailed.Error())
	t.Nil(results)
}

func (t *TestSuite) TestMessageCatalog_SearchMsgids_InvalidRegex() {
	results, err := t.mc.SearchMsgids(`****`)
	t.EqualError(err, "error parsing regexp: missing argument to repetition operator: `*`")
	t.Nil(results)
}
