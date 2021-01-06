package gogettext

import (
	"encoding/json"
	"sync"

	"github.com/taylor-s-dean/gogettext/po2json"
)

/*
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
*/

type MessageCatalog struct {
	messages map[string]interface{}
	mutex    sync.RWMutex
}

func NewMessageCatalogFromFile(filePath string) (*MessageCatalog, error) {
	mc := &MessageCatalog{}
	var err error

	mc.mutex.Lock()
	mc.messages, err = po2json.LoadFile(filePath)
	mc.mutex.Unlock()

	if err != nil {
		return nil, err
	}

	return mc, nil
}

func NewMessageCatalogFromString(fileContents string) (*MessageCatalog, error) {
	mc := &MessageCatalog{}
	var err error

	mc.mutex.Lock()
	mc.messages, err = po2json.LoadString(fileContents)
	mc.mutex.Unlock()

	if err != nil {
		return nil, err
	}

	return mc, nil
}

func NewMessageCatalogFromBytes(fileContents []byte) (*MessageCatalog, error) {
	mc := &MessageCatalog{}
	var err error

	mc.mutex.Lock()
	mc.messages, err = po2json.LoadBytes(fileContents)
	mc.mutex.Unlock()

	if err != nil {
		return nil, err
	}

	return mc, nil
}

func (mc *MessageCatalog) GetMessages() (map[string]interface{}, error) {
	mc.mutex.RLock()
	messagesBytes, err := json.Marshal(mc.messages)
	mc.mutex.RUnlock()

	if err != nil {
		return nil, err
	}

	messages := &map[string]interface{}{}
	if err := json.Unmarshal(messagesBytes, messages); err != nil {
		return nil, err
	}

	return *messages, nil
}
