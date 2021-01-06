package gogettext

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/Knetic/govaluate"
	"github.com/taylor-s-dean/gogettext/po2json"
)

var (
	pluralFormsRegex = regexp.MustCompile(`nplurals\s*=\s*(\d+);\s*plural\s*=\s*([n0-9%!=&|?:><+() \-]+);`)
)

type MessageCatalog struct {
	messages    map[string]interface{}
	mutex       sync.RWMutex
	pluralForms *govaluate.EvaluableExpression
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
	mc.pluralForms, err = govaluate.NewEvaluableExpression("n==1 ? 0 : 1")
	if err != nil {
		return err
	}

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

	pluralForms, err := govaluate.NewEvaluableExpression(matches[2])
	if err != nil {
		return err
	}

	mc.pluralForms = pluralForms
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

	idxObj, err := mc.pluralForms.Evaluate(map[string]interface{}{"n": n})
	if err != nil {
		fmt.Println(err)
		fmt.Println("HERE 0")
		return fallbackMsgstr
	}

	idx, ok := idxObj.(int)
	if !ok {
		fmt.Println("HERE 1")
		return fallbackMsgstr
	}

	if mc.messages == nil {
		fmt.Println("HERE 2")
		return fallbackMsgstr
	}

	msgctxtObj, ok := mc.messages[""]
	if !ok {
		fmt.Println("HERE 3")
		return fallbackMsgstr
	}

	msgctxtMap, ok := msgctxtObj.(map[string]interface{})
	if !ok {
		fmt.Println("HERE 4")
		return fallbackMsgstr
	}

	msgidObj, ok := msgctxtMap[fallbackMsgstr]
	if !ok {
		fmt.Println("HERE 5")
		return fallbackMsgstr
	}

	msgidMap, ok := msgidObj.(map[string]interface{})
	if !ok {
		fmt.Println("HERE 6")
		return fallbackMsgstr
	}

	msgstrPluralsObj, ok := msgidMap["plurals"]
	if !ok {
		fmt.Println("HERE 7")
		return fallbackMsgstr
	}

	msgstrPluralsList, ok := msgstrPluralsObj.([]string)
	if !ok {
		fmt.Println("HERE 8")
		return fallbackMsgstr
	}

	if len(msgstrPluralsList) < idx+1 {
		fmt.Println("HERE 9")
		return fallbackMsgstr
	}

	fmt.Println("HERE 10")
	return msgstrPluralsList[idx]
}
