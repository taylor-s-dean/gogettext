
%{
package pluralsparser

import (
    "fmt"
    "strconv"
    "strings"
    "unicode/utf8"

    "github.com/pkg/errors"
)
%}

%union {
    num uint64
    str string
}

%token <str> tokIDENTIFIER
%token <num> tokNUMBER
%token
    tokMOD          // %
    tokMULTIPLY     // *
    tokDIVIDE       // /
    tokADD          // +
    tokSUBTRACT     // -
    tokTHEN         // ?
    tokELSE         // :
    tokLT           // <
    tokLE           // <=
    tokGT           // >
    tokGE           // >=
    tokEQ           // ==
    tokNE           // !=
    tokAND          // &&
    tokOR           // ||
    tokLPAREN       // (
    tokRPAREN       // )
    tokINVALID
;

%type <num>
    unit
    expression
    if_statement
    multiplicative
    additive
    associative
    relational
    equality
    logical
;

%right tokTHEN tokELSE
%left tokOR
%left tokAND
%left tokEQ tokNE
%left tokLT tokLE tokGT tokGE
%left tokADD tokSUBTRACT
%left tokMOD tokMULTIPLY tokDIVIDE

%start unit;

%%

unit: if_statement
    {
        yylex.(*yyLex).Result = $1
    }
;

if_statement:
  expression
| expression tokTHEN if_statement tokELSE if_statement
    {
        if $1 != 0 {
            $$ = $3
        } else {
            $$ = $5
        }
    }
;

multiplicative:
  expression tokMOD expression      { $$ = $1 % $3 }
| expression tokMULTIPLY expression { $$ = $1 * $3 }
| expression tokDIVIDE expression   { $$ = $1 / $3 }

associative: tokLPAREN if_statement tokRPAREN   { $$ = $2 }

relational:
  expression tokLT expression
    {
        if $1 < $3 {
            $$ = 1
        } else {
            $$ = 0
        }
    }
| expression tokLE expression
    {
        if $1 <= $3 {
            $$ = 1
        } else {
            $$ = 0
        }
    }
| expression tokGT expression
    {
        if $1 > $3 {
            $$ = 1
        } else {
            $$ = 0
        }
    }
| expression tokGE expression
    {
        if $1 >= $3 {
            $$ = 1
        } else {
            $$ = 0
        }
    }
;

equality:
  expression tokEQ expression
    {
        if $1 == $3 {
            $$ = 1
        } else {
            $$ = 0
        }
    }
| expression tokNE expression
    {
        if $1 != $3 {
            $$ = 1
        } else {
            $$ = 0
        }
    }
;

logical:
  expression tokAND expression
    {
        if $1 != 0 && $3 != 0 {
            $$ = 1
        } else {
            $$ = 0
        }
    }
| expression tokOR expression
   {
        if $1 != 0 || $3 != 0 {
            $$ = 1
        } else {
            $$ = 0
        }
    }
;

additive:
  expression tokADD expression      { $$ = $1 + $3 }
| expression tokSUBTRACT expression { $$ = $1 - $3 }

expression:
  tokNUMBER
| tokIDENTIFIER  { $$ = yylex.(*yyLex).Variables[$1] }
| associative
| multiplicative
| additive
| relational
| equality
| logical
;
%%

const eof = 0

type yyLex struct {
	line      []byte
	peek      rune
    idx       int
    orig      []byte
    Result    uint64
    Variables map[string]uint64
    Err       error
}

var isNumber = map[rune]bool{
    '0': true,
    '1': true,
    '2': true,
    '3': true,
    '4': true,
    '5': true,
    '6': true,
    '7': true,
    '8': true,
    '9': true,
}

var isWhitespace = map[rune]bool{
    ' ':  true,
    '\t': true,
    '\n': true,
    '\r': true,
}

func (x *yyLex) Lex(yylval *yySymType) (res int) {
	for {
		c := x.next()
        p := x.peek
		switch {
		case c == eof:
			return eof
        case c == '0':
            yylval.num = 0
            return tokNUMBER
		case isNumber[c]:
			return x.num(c, yylval)
		case c == '>' && p != '=':
			return tokGT
		case c == '>' && p == '=':
            x.next()
			return tokGE
        case c == '<' && p != '=':
            return tokLT
        case c == '<' && p == '=':
            x.next()
            return tokLE
        case c == '&' && p == '&':
            x.next()
            return tokAND
        case c == '|' && p == '|':
            x.next()
            return tokOR
        case c == '!' && p == '=':
            x.next()
            return tokNE
        case c == '=' && p == '=':
            x.next()
            return tokEQ
        case c == '?':
            return tokTHEN
        case c == ':':
            return tokELSE
        case c == '(':
            return tokLPAREN
        case c == ')':
            return tokRPAREN
        case c == '+':
            return tokADD
        case c == '-':
            return tokSUBTRACT
        case c == '*':
            return tokMULTIPLY
        case c == '/':
            return tokDIVIDE
        case c == '%':
            return tokMOD
        case c == 'n':
            yylval.str = string(c)
            return tokIDENTIFIER
		case isWhitespace[c]:
		default:
            return tokINVALID
		}
	}
}

func (x *yyLex) num(c rune, yylval *yySymType) int {
	add := func(b *strings.Builder, c rune) error {
		if _, err := b.WriteRune(c); err != nil {
            return errors.Wrap(err, fmt.Sprintf("Failed to write rune %q", c))
		}
        return nil
	}

	b := strings.Builder{}
	if err := add(&b, c); err != nil {
        x.Err = err
        return eof
    }

	L: for {
		switch {
		case isNumber[x.peek]:
		    c = x.next()
            if err := add(&b, c); err != nil {
                x.Err = err
                return eof
            }
		default:
            break L
		}
	}

    var err error
	yylval.num, err = strconv.ParseUint(b.String(), 10, 64)
	if err != nil {
        x.Err = errors.Wrap(err, fmt.Sprintf("ERROR: %s. Bad number %q", err, b.String()))
		return eof
	}
	return tokNUMBER
}

func (x *yyLex) next() rune {
    x.idx++
    r := x.peek
    c, size := utf8.DecodeRune(x.line)
    x.line = x.line[size:]
    if size == 0 {
        c = eof
    }
    x.peek = c
	return r
}

func (x *yyLex) Error(s string) {
    ss := strings.Builder{}
    for i := 0; i < x.idx + 1; i++ {
        if x.idx == i {
            ss.WriteRune('^')
        } else {
            ss.WriteRune(' ')
        }
    }
    x.Err = fmt.Errorf("parse error: %s\n%s\n%s\n", s, x.orig, ss.String())
}

func newLexer(line []byte, n uint64) *yyLex {
    c, size := utf8.DecodeRune(line)
    return &yyLex{
        line:      line[size:],
        peek:      c,
        idx:       -1,
        orig:      line,
        Result:    0,
        Variables: map[string]uint64{"n": n},
        Err:       nil,
    }
}

// Evaluate compiles and evalutes the provided Plural-Format ternary string
// while supstituting the variable "n" with the provided value.
//
// Returns the resulting index into the plural array and an error if an
// error was encountered.
func Evaluate(expression string, n uint64) (uint64, error) {
    yyErrorVerbose = true
    l := newLexer([]byte(expression), n)
    yyParse(l)
    return l.Result, l.Err
}
