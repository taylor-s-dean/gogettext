package po2json

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
