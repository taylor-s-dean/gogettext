package po2json

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Key struct {
	Msgctxt      strings.Builder
	Msgid        strings.Builder
	Msgstr       strings.Builder
	MsgidPlural  strings.Builder
	MsgstrPlural []*strings.Builder
}

type State int

const (
	stateUnspecified State = iota
	stateMsgctxt
	stateMsgid
	stateMsgstr
	stateMsgidPlural
	stateMsgstrPlural
)

var (
	stateStrings = map[State]string{
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
	currentKey      Key
	currentState    State
	validNextStates map[State]bool
	poJSON          map[string]interface{}
}

func (l *Loader) init() {
	l.currentState = stateUnspecified
	l.validNextStates = map[State]bool{stateMsgctxt: true, stateMsgid: true}
	l.poJSON = map[string]interface{}{}
}

func (l *Loader) Load(fileContents string) (map[string]interface{}, error) {
	l.init()

	for _, line := range strings.Split(fileContents, "\n") {
		fmt.Println("LINE::" + line)
		// Ignore the line if it is a comment.
		// We expect the next line to be anything.
		if regexComment.MatchString(line) {
			fmt.Println("STATE::COMMENT")
			continue
		}

		// If this is an empty line, then we expect the next
		// non-empty non-comment line to be msgctxt or msgid.
		if regexEmpty.MatchString(line) {
			fmt.Println("STATE::EMPTY")
			if err := l.addKeyToJson(); err != nil {
				return nil, err
			}
			l.currentKey = Key{}
			l.currentState = stateUnspecified
			l.validNextStates = map[State]bool{stateMsgctxt: true, stateMsgid: true}
			continue
		}

		// If this is a msgctxt line, then:
		// 1) msgctxt must be a valid state.
		// 2) We expect the next line to be either a string or a msgid.
		if submatch := regexMsgctxt.FindStringSubmatch(line); submatch != nil {
			fmt.Println("STATE::MSGCTXT")
			l.currentState = stateMsgctxt
			if err := l.expectState(); err != nil {
				return nil, err
			}

			l.validNextStates = map[State]bool{stateMsgid: true}
			l.currentKey.Msgctxt.WriteString(submatch[1])
			continue
		}

		// If this is a msgid line, then:
		// 1) msgid must be a valid state.
		// 2) We expect the next line to be either a string, msgstr, or msgid_plural.
		if submatch := regexMsgid.FindStringSubmatch(line); submatch != nil {
			fmt.Println("STATE::MSGID")
			l.currentState = stateMsgid
			if err := l.expectState(); err != nil {
				return nil, err
			}

			l.validNextStates = map[State]bool{stateMsgidPlural: true, stateMsgstr: true}
			l.currentKey.Msgid.WriteString(submatch[1])
			continue
		}

		// If this is a msgstr line, then:
		// 1) msgstr must be a valid state.
		// 2) We expect the next line to be either a string or blank.
		if submatch := regexMsgstr.FindStringSubmatch(line); submatch != nil {
			fmt.Println("STATE::MSGSTR")
			l.currentState = stateMsgstr
			if err := l.expectState(); err != nil {
				return nil, err
			}

			l.validNextStates = map[State]bool{stateMsgidPlural: true, stateMsgstr: true}
			l.currentKey.Msgstr.WriteString(submatch[1])
			continue
		}

		// If this is a msgid_plural line, then:
		// 1) msgid_plural must be a valid state.
		// 2) We expect the next line to be either a string or msgstr_plural.
		if submatch := regexMsgidPlural.FindStringSubmatch(line); submatch != nil {
			fmt.Println("STATE::MSGID_PLURAL")
			l.currentState = stateMsgidPlural
			if err := l.expectState(); err != nil {
				return nil, err
			}

			l.validNextStates = map[State]bool{stateMsgstrPlural: true}
			l.currentKey.MsgidPlural.WriteString(submatch[1])
			continue
		}

		// If this is a msgstr_plural line, then:
		// 1) msgstr_plural must be a valid state.
		// 2) We expect the next line to be either a string, msgstr_plural, or blank.
		if submatch := regexMsgstrPlural.FindStringSubmatch(line); submatch != nil {
			fmt.Println("STATE::MSGSTR_PLURAL")
			l.currentState = stateMsgstrPlural
			if err := l.expectState(); err != nil {
				return nil, err
			}

			l.validNextStates = map[State]bool{stateMsgstrPlural: true}
			plural := strings.Builder{}
			plural.WriteString(submatch[1])
			l.currentKey.MsgstrPlural = append(l.currentKey.MsgstrPlural, &plural)
			continue
		}

		// If this is a string continuation, then:
		// 1) Append the string to the existing string as determined by the
		// current_state.
		if submatch := regexString.FindStringSubmatch(line); submatch != nil {
			fmt.Println("STRING")
			switch l.currentState {
			case stateMsgctxt:
				l.currentKey.Msgctxt.WriteString(submatch[1])
			case stateMsgid:
				l.currentKey.Msgid.WriteString(submatch[1])
			case stateMsgstr:
				l.currentKey.Msgstr.WriteString(submatch[1])
			case stateMsgidPlural:
				l.currentKey.MsgidPlural.WriteString(submatch[1])
			case stateMsgstrPlural:
				l.currentKey.MsgstrPlural[len(l.currentKey.MsgstrPlural)-1].WriteString(submatch[1])
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
	msgctxt := l.currentKey.Msgctxt.String()
	msgid := l.currentKey.Msgid.String()
	msgstr := l.currentKey.Msgstr.String()
	msgstrPlural := []string{}
	for _, plural := range l.currentKey.MsgstrPlural {
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
			fmt.Println("HEADER")
			for _, submatch := range regexHeaderKeyValue.FindAllStringSubmatch(msgstr, -1) {
				key := submatch[1]
				if _, ok := msgidObj[key]; ok {
					return fmt.Errorf("Invalid .po file. Found duplicate header key (%s).", key)
				}
				msgidObj[key] = submatch[2]
			}
		} else {
			fmt.Println("TRANSLATION")
			if _, ok := msgidObj["translation"]; ok {
				return fmt.Errorf("Invalid .po file. Found duplicate msgstr for msgid (%s).", msgid)
			}
			msgidObj["translation"] = msgstr
		}
	}
	if len(msgstrPlural) > 0 {
		if _, ok := msgidObj["plurals"]; !ok {
			msgidObj["plurals"] = &[]string{}
		}
		pluralsObj := msgidObj["plurals"].(*[]string)
		*pluralsObj = append(*pluralsObj, msgstrPlural...)
	}

	return nil
}

func (l *Loader) expectState() error {
	if !l.validNextStates[l.currentState] {
		return errors.New(fmt.Sprintf("Invalid .po file. Found %s, expected one of %s.", stateStrings[l.currentState], l.printValidNextStates()))
	}
	return nil
}

func (l *Loader) printValidNextStates() string {
	ss := strings.Builder{}
	ss.WriteRune('{')
	for state := range l.validNextStates {
		ss.WriteString(stateStrings[state])
	}
	ss.WriteRune('}')
	return ss.String()
}
