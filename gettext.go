//go:generate goyacc -o ./plurals-parser/parser.go ./plurals-parser/parser.yy
package gogettext

import (
	"encoding/json"
	"regexp"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/taylor-s-dean/gogettext/plurals-parser"
	"github.com/taylor-s-dean/gogettext/po2json"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrorMsgctxtNotFound                = Error("message context not found")
	ErrorMsgidNotFound                  = Error("message identifier not found")
	ErrorTranslationNotFound            = Error("translation not found")
	ErrorPluralNotFound                 = Error("plurals not found")
	ErrorNilMessageCatalog              = Error("message catalog is nil")
	ErrorMsgctxtTypeAssertionFailed     = Error("message context type assertion failed")
	ErrorMsgidTypeAssertionFailed       = Error("message context type assertion failed")
	ErrorPluralsTypeAssertionFailed     = Error("message context type assertion failed")
	ErrorTranslationTypeAssertionFailed = Error("message context type assertion failed")
	ErrorPluralIndexOutOfBounds         = Error("plural index out of bounds")
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
		return nil, errors.Wrap(err, "failed to load .po file")
	}

	if err := mc.setPluralForms(); err != nil {
		return nil, errors.Wrap(err, "failed to set plural forms")
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
		return nil, errors.Wrap(err, "failed to load .po file")
	}

	if err := mc.setPluralForms(); err != nil {
		return nil, errors.Wrap(err, "failed to set plural forms")
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
		return nil, errors.Wrap(err, "failed to load .po file")
	}

	if err := mc.setPluralForms(); err != nil {
		return nil, errors.Wrap(err, "failed to set plural forms")
	}

	return mc, nil
}

func (mc *MessageCatalog) GetMessages() (map[string]interface{}, error) {
	mc.mutex.RLock()
	messagesBytes, err := json.Marshal(mc.messages)
	mc.mutex.RUnlock()

	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal message catalog")
	}

	messages := &map[string]interface{}{}
	if err := json.Unmarshal(messagesBytes, messages); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal message catalog")
	}

	return *messages, nil
}

func (mc *MessageCatalog) setPluralForms() error {
	var err error
	mc.pluralForms = "n==1 ? 0 : 1"

	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	msgidMap, err := mc.getMsgidMap("", "")
	if err != nil {
		return err
	}

	pluralFormsObj, ok := msgidMap["Plural-Forms"]
	if !ok {
		return nil
	}

	pluralFormsStr, ok := pluralFormsObj.(string)
	if !ok {
		return ErrorTranslationTypeAssertionFailed
	}

	matches := pluralFormsRegex.FindStringSubmatch(pluralFormsStr)
	if matches == nil {
		return nil
	}

	mc.pluralForms = matches[2]
	if _, err := pluralsparser.Evaluate(mc.pluralForms, 0); err != nil {
		return err
	}

	mc.nPlurals, err = strconv.ParseInt(matches[1], 10, 64)

	if err != nil {
		return err
	}

	return nil
}

func (mc *MessageCatalog) getMsgidMap(msgctxt string, msgid string) (map[string]interface{}, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if mc.messages == nil {
		return nil, ErrorNilMessageCatalog
	}

	msgctxtObj, ok := mc.messages[msgctxt]
	if !ok {
		return nil, ErrorMsgctxtNotFound
	}

	msgctxtMap, ok := msgctxtObj.(map[string]interface{})
	if !ok {
		return nil, ErrorMsgctxtTypeAssertionFailed
	}

	msgidObj, ok := msgctxtMap[msgid]
	if !ok {
		return nil, ErrorMsgidNotFound
	}

	msgidMap, ok := msgidObj.(map[string]interface{})
	if !ok {
		return nil, ErrorMsgidTypeAssertionFailed
	}

	return msgidMap, nil
}

// Gettext returns the msgstr associated with the msgid.
// This method returns the msgid if the corresponding msgstr cannot be found.
func (mc *MessageCatalog) Gettext(msgid string) string {
	msgstr, _ := mc.TryGettext(msgid)
	return msgstr
}

// TryGettext returns the msgstr associated with the msgid.
// This method returns msgid and an error if the corresponding msgstr cannot be
// found.
func (mc *MessageCatalog) TryGettext(msgid string) (string, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	msgidMap, err := mc.getMsgidMap("", msgid)
	if err != nil {
		return msgid, err
	}

	msgstrObj, ok := msgidMap["translation"]
	if !ok {
		return msgid, ErrorTranslationNotFound
	}

	msgstrStr, ok := msgstrObj.(string)
	if !ok {
		return msgid, ErrorTranslationTypeAssertionFailed
	}

	return msgstrStr, nil
}

// NGettext returns the msgid_plural associate with the given msgid and
// quantity.
// The specific plural returned is determined by evaluating the Plural-Forms
// header given the specified quantity.
// The Plural-Forms falls back to `plural=(n==1 ? 0 : 1)` in the event that
// no Plural-Forms header was provided in the .po file OR evaluation of the
// plural failed.
// In the case of plural evaluation failure or failure to find the associated
// msgstr, msgidSingular is returned if quantity == 1, otherwise
// msgidPlural is returned.
func (mc *MessageCatalog) NGettext(msgidSingular string, msgidPlural string, quantity int) string {
	msgstr, _ := mc.TryNGettext(msgidSingular, msgidPlural, quantity)
	return msgstr
}

// TryNGettext returns the msgid_plural associate with the given msgid and
// quantity.
// The specific plural returned is determined by evaluating the Plural-Forms
// header given the specified quantity.
// The Plural-Forms falls back to `plural=(n==1 ? 0 : 1)` in the event that
// no Plural-Forms header was provided in the .po file OR evaluation of the
// plural failed.
// In the case of plural evaluation failure or failure to find the associated
// msgstr, msgidSingular is returned if quantity == 1, otherwise
// msgidPlural is returned. An error is also returned in these cases.
func (mc *MessageCatalog) TryNGettext(msgidSingular string, msgidPlural string, quantity int) (string, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	fallbackMsgstr := msgidSingular
	if quantity != 1 {
		fallbackMsgstr = msgidPlural
	}

	idxUint, err := pluralsparser.Evaluate(mc.pluralForms, uint64(quantity))
	if err != nil {
		return fallbackMsgstr, err
	}

	msgidMap, err := mc.getMsgidMap("", msgidSingular)
	if err != nil {
		return fallbackMsgstr, err
	}

	msgstrPluralsObj, ok := msgidMap["plurals"]
	if !ok {
		return fallbackMsgstr, ErrorPluralNotFound
	}

	msgstrPluralsList, ok := msgstrPluralsObj.([]string)
	if !ok {
		return fallbackMsgstr, ErrorPluralsTypeAssertionFailed
	}

	idx := int(idxUint)
	if idx >= len(msgstrPluralsList) {
		return fallbackMsgstr, ErrorPluralIndexOutOfBounds
	}

	return msgstrPluralsList[idx], nil
}

// PGettext returns the Particular msgstr associated with the msgctxt
// and msgid.
// This method will return the msgid if no corresponding msgstr can be
// found.
func (mc *MessageCatalog) PGettext(msgctxt string, msgid string) string {
	msgstr, _ := mc.TryPGettext(msgctxt, msgid)
	return msgstr
}

// TryPGettext returns the Particular msgstr associated with the msgctxt
// and msgid.
// This method will return the msgid and an error if no corresponding msgstr
// can be found.
func (mc *MessageCatalog) TryPGettext(msgctxt string, msgid string) (string, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	msgidMap, err := mc.getMsgidMap(msgctxt, msgid)
	if err != nil {
		return msgid, err
	}

	msgstrObj, ok := msgidMap["translation"]
	if !ok {
		return msgid, ErrorTranslationNotFound
	}

	msgstrStr, ok := msgstrObj.(string)
	if !ok {
		return msgid, ErrorTranslationTypeAssertionFailed
	}

	return msgstrStr, nil
}
