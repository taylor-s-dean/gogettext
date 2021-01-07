//go:generate goyacc -o ./plurals-parser/parser.go ./plurals-parser/parser.yy
package gogettext

import (
	"encoding/json"
	"regexp"
	"strconv"
	"sync"

	"github.com/taylor-s-dean/gogettext/plurals-parser"
	"github.com/taylor-s-dean/gogettext/po2json"
)

var (
	pluralFormsRegex = regexp.MustCompile(`nplurals\s*=\s*(\d+);\s*plural\s*=\s*([n0-9%!=&|?:><+() \-]+);`)
)

type MessageCatalog struct {
	messages    map[string]interface{}
	mutex       sync.RWMutex
	pluralForms string
	nPlurals    int64
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

	if err := mc.setPluralForms(); err != nil {
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

	if err := mc.setPluralForms(); err != nil {
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

	if err := mc.setPluralForms(); err != nil {
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

func (mc *MessageCatalog) setPluralForms() error {
	var err error
	mc.pluralForms = "n==1 ? 0 : 1"

	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if mc.messages == nil {
		return nil
	}

	msgctxtObj, ok := mc.messages[""]
	if !ok {
		return nil
	}

	msgctxtMap, ok := msgctxtObj.(map[string]interface{})
	if !ok {
		return nil
	}

	msgidObj, ok := msgctxtMap[""]
	if !ok {
		return nil
	}

	msgidMap, ok := msgidObj.(map[string]interface{})
	if !ok {
		return nil
	}

	pluralFormsObj, ok := msgidMap["Plural-Forms"]
	if !ok {
		return nil
	}

	pluralFormsStr, ok := pluralFormsObj.(string)
	if !ok {
		return nil
	}

	matches := pluralFormsRegex.FindStringSubmatch(pluralFormsStr)
	if matches == nil {
		return nil
	}

	mc.pluralForms = matches[2]
	mc.nPlurals, err = strconv.ParseInt(matches[1], 10, 64)

	if err != nil {
		return err
	}

	return nil
}

func (mc *MessageCatalog) Gettext(msgid string) string {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if mc.messages == nil {
		return msgid
	}

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

func (mc *MessageCatalog) NGettext(msgidSingular string, msgidPlural string, n int) string {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	fallbackMsgstr := msgidSingular
	if n != 1 {
		fallbackMsgstr = msgidPlural
	}

	idxUint, err := pluralsparser.Evaluate(mc.pluralForms, uint64(n))
	if err != nil {
		return fallbackMsgstr
	}

	idx := int(idxUint)
	if mc.messages == nil {
		return fallbackMsgstr
	}

	msgctxtObj, ok := mc.messages[""]
	if !ok {
		return fallbackMsgstr
	}

	msgctxtMap, ok := msgctxtObj.(map[string]interface{})
	if !ok {
		return fallbackMsgstr
	}

	msgidObj, ok := msgctxtMap[msgidSingular]
	if !ok {
		return fallbackMsgstr
	}

	msgidMap, ok := msgidObj.(map[string]interface{})
	if !ok {
		return fallbackMsgstr
	}

	msgstrPluralsObj, ok := msgidMap["plurals"]
	if !ok {
		return fallbackMsgstr
	}

	msgstrPluralsList, ok := msgstrPluralsObj.([]string)
	if !ok {
		return fallbackMsgstr
	}

	if len(msgstrPluralsList) < idx+1 {
		return fallbackMsgstr
	}

	return msgstrPluralsList[idx]
}
