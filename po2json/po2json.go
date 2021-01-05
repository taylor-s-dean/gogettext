package po2json

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type translationKey struct {
	Msgctxt      strings.Builder
	Msgid        strings.Builder
	Msgstr       strings.Builder
	MsgidPlural  strings.Builder
	MsgstrPlural []*strings.Builder
}

type stateEnum int

const (
	stateUnspecified stateEnum = iota
	stateMsgctxt
	stateMsgid
	stateMsgstr
	stateMsgidPlural
	stateMsgstrPlural
)

var (
	stateStrings = map[stateEnum]string{
		stateUnspecified:  "unspecified",
		stateMsgctxt:      "msgctxt",
		stateMsgid:        "msgid",
		stateMsgstr:       "msgstr",
		stateMsgidPlural:  "msgid_plural",
		stateMsgstrPlural: "msgstr_plural",
	}

	regexComment        = regexp.MustCompile(`\s*#.*`)
	regexEmpty          = regexp.MustCompile(`^\s*$`)
	regexMsgctxt        = regexp.MustCompile(`msgctxt\s+"(.*)"`)
	regexMsgid          = regexp.MustCompile(`msgid\s+"(.*)"`)
	regexMsgstr         = regexp.MustCompile(`msgstr\s+"(.*)"`)
	regexMsgidPlural    = regexp.MustCompile(`msgid_plural\s+"(.*)"`)
	regexMsgstrPlural   = regexp.MustCompile(`msgstr\[\d+\]\s+"(.*)"`)
	regexString         = regexp.MustCompile(`"(.*)"`)
	regexHeaderKeyValue = regexp.MustCompile(`([a-zA-Z0-9-]+)\s*:\s*(.*?)\\n`)
)

type Loader struct {
	key        translationKey
	state      stateEnum
	nextStates map[stateEnum]bool
	poJSON     map[string]interface{}
}

func (l *Loader) init() {
	l.state = stateUnspecified
	l.nextStates = map[stateEnum]bool{stateMsgctxt: true, stateMsgid: true}
	l.poJSON = map[string]interface{}{}
}

func (l *Loader) Load(fileContents string) (map[string]interface{}, error) {
	l.init()

	for _, line := range strings.Split(fileContents, "\n") {
		// Ignore the line if it is a comment.
		// We expect the next line to be anything.
		if regexComment.MatchString(line) {
			continue
		}

		// If this is an empty line, then we expect the next
		// non-empty non-comment line to be msgctxt or msgid.
		if regexEmpty.MatchString(line) {
			if err := l.addKeyToJson(); err != nil {
				return nil, err
			}

			l.key = translationKey{}
			l.state = stateUnspecified
			l.nextStates = map[stateEnum]bool{stateMsgctxt: true, stateMsgid: true}
			continue
		}

		// If this is a msgctxt line, then:
		// 1) msgctxt must be a valid state.
		// 2) We expect the next line to be either a string or a msgid.
		if submatch := regexMsgctxt.FindStringSubmatch(line); submatch != nil {
			l.state = stateMsgctxt
			if err := l.expectState(); err != nil {
				return nil, err
			}

			l.nextStates = map[stateEnum]bool{stateMsgid: true}
			l.key.Msgctxt.WriteString(submatch[1])
			continue
		}

		// If this is a msgid line, then:
		// 1) msgid must be a valid state.
		// 2) We expect the next line to be either a string, msgstr, or msgid_plural.
		if submatch := regexMsgid.FindStringSubmatch(line); submatch != nil {
			l.state = stateMsgid
			if err := l.expectState(); err != nil {
				return nil, err
			}

			l.nextStates = map[stateEnum]bool{stateMsgidPlural: true, stateMsgstr: true}
			l.key.Msgid.WriteString(submatch[1])
			continue
		}

		// If this is a msgstr line, then:
		// 1) msgstr must be a valid state.
		// 2) We expect the next line to be either a string or blank.
		if submatch := regexMsgstr.FindStringSubmatch(line); submatch != nil {
			l.state = stateMsgstr
			if err := l.expectState(); err != nil {
				return nil, err
			}

			l.nextStates = map[stateEnum]bool{stateMsgidPlural: true, stateMsgstr: true}
			l.key.Msgstr.WriteString(submatch[1])
			continue
		}

		// If this is a msgid_plural line, then:
		// 1) msgid_plural must be a valid state.
		// 2) We expect the next line to be either a string or msgstr_plural.
		if submatch := regexMsgidPlural.FindStringSubmatch(line); submatch != nil {
			l.state = stateMsgidPlural
			if err := l.expectState(); err != nil {
				return nil, err
			}

			l.nextStates = map[stateEnum]bool{stateMsgstrPlural: true}
			l.key.MsgidPlural.WriteString(submatch[1])
			continue
		}

		// If this is a msgstr_plural line, then:
		// 1) msgstr_plural must be a valid state.
		// 2) We expect the next line to be either a string, msgstr_plural, or blank.
		if submatch := regexMsgstrPlural.FindStringSubmatch(line); submatch != nil {
			l.state = stateMsgstrPlural
			if err := l.expectState(); err != nil {
				return nil, err
			}

			l.nextStates = map[stateEnum]bool{stateMsgstrPlural: true}
			plural := strings.Builder{}
			plural.WriteString(submatch[1])
			l.key.MsgstrPlural = append(l.key.MsgstrPlural, &plural)
			continue
		}

		// If this is a string continuation, then:
		// 1) Append the string to the existing string as determined by the
		// current_state.
		if submatch := regexString.FindStringSubmatch(line); submatch != nil {
			switch l.state {
			case stateMsgctxt:
				l.key.Msgctxt.WriteString(submatch[1])
			case stateMsgid:
				l.key.Msgid.WriteString(submatch[1])
			case stateMsgstr:
				l.key.Msgstr.WriteString(submatch[1])
			case stateMsgidPlural:
				l.key.MsgidPlural.WriteString(submatch[1])
			case stateMsgstrPlural:
				l.key.MsgstrPlural[len(l.key.MsgstrPlural)-1].WriteString(submatch[1])
			case stateUnspecified:
				return nil, errors.New("Encountered invalid state. Please ensure the input file is in a valid .po format.")
			}
			continue
		}
	}

	err := l.addKeyToJson()
	if err != nil {
		return nil, err
	}

	return l.poJSON, nil
}

// TODO: Check submatch lengths to ensure no index out of bounds
func (l *Loader) addKeyToJson() error {
	msgctxt := l.key.Msgctxt.String()
	msgid := l.key.Msgid.String()
	msgstr := l.key.Msgstr.String()
	msgstrPlural := []string{}
	for _, plural := range l.key.MsgstrPlural {
		msgstrPlural = append(msgstrPlural, plural.String())
	}

	if _, ok := l.poJSON[msgctxt]; !ok {
		l.poJSON[msgctxt] = map[string]interface{}{}
	}
	msgctxtObj := l.poJSON[msgctxt].(map[string]interface{})

	if _, ok := msgctxtObj[msgid]; !ok {
		msgctxtObj[msgid] = map[string]interface{}{}
	}
	msgidObj := msgctxtObj[msgid].(map[string]interface{})

	if len(msgstr) > 0 {
		if len(msgid) == 0 {
			for _, submatch := range regexHeaderKeyValue.FindAllStringSubmatch(msgstr, -1) {
				key := submatch[1]
				if _, ok := msgidObj[key]; ok {
					return fmt.Errorf("Invalid .po file. Found duplicate header key (%s).", key)
				}
				msgidObj[key] = submatch[2]
			}
		} else {
			if _, ok := msgidObj["translation"]; ok {
				return fmt.Errorf("Invalid .po file. Found duplicate msgstr for msgid (%s).", msgid)
			}
			msgidObj["translation"] = msgstr
		}
	}
	if len(msgstrPlural) > 0 {
		msgidObj["plurals"] = msgstrPlural
	}

	return nil
}

func (l *Loader) expectState() error {
	if !l.nextStates[l.state] {
		return errors.New(fmt.Sprintf("Invalid .po file. Found %s, expected one of %s.", stateStrings[l.state], l.printNextStates()))
	}
	return nil
}

func (l *Loader) printNextStates() string {
	ss := strings.Builder{}
	ss.WriteRune('{')
	for state := range l.nextStates {
		ss.WriteString(stateStrings[state])
	}
	ss.WriteRune('}')
	return ss.String()
}
