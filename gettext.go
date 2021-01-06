package gogettext

import (
	"encoding/json"
	"sync"

	"github.com/taylor-s-dean/gogettext/po2json"
)

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

func (mc *MessageCatalog) Gettext(msgid string) string {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	msgctxtObj, ok := mc.messages[""]
	if !ok {
		return msgid
	}

	msgctxtMap, ok := msgctxtObj.(map[string]interface{})
	if !ok {
		return msgid
	}

	msgidObj, ok := msgctxtMap[msgid]
	if !ok {
		return msgid
	}

	msgidMap, ok := msgidObj.(map[string]interface{})
	if !ok {
		return msgid
	}

	msgstrObj, ok := msgidMap["translation"]
	if !ok {
		return msgid
	}

	msgstrStr, ok := msgstrObj.(string)
	if !ok {
		return msgid
	}

	return msgstrStr
}
