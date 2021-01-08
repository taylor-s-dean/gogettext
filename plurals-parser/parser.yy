
%{
package pluralsparser

import (
    "fmt"
    "log"
    "unicode/utf8"
    "strings"
    "strconv"
)
%}

%union {
    num uint64
    str string
}

%token <str> tokIDENTIFIER
%token <num> tokNUMBER
%token
    tokMOD
    tokTHEN
    tokELSE
    tokLT
    tokLE
    tokGT
    tokGE
    tokEQ
    tokNE
    tokAND
    tokOR
    tokLPAREN
    tokRPAREN
    tokINVALID
;

%type <num>
    unit
    expression
    if_statement
    multiplicative
    associative
    relational
    equality
    logical
;

%left tokOR
%left tokAND;
%left tokEQ tokNE;
%left tokLT tokLE tokGT tokGE;
%left tokMOD;
%left tokTHEN;
%right tokELSE;

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

multiplicative: expression tokMOD expression   { $$ = $1 % $3 };

associative: tokLPAREN if_statement tokRPAREN   { $$ = $2 };

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

expression:
  tokNUMBER
| tokIDENTIFIER  { $$ = yylex.(*yyLex).Variables[$1] }
| associative
| multiplicative
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
        case c == '%':
            return tokMOD
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
	add := func(b *strings.Builder, c rune) {
		if _, err := b.WriteRune(c); err != nil {
			log.Fatalf("WriteRune: %s", err)
		}
	}
	b := strings.Builder{}
	add(&b, c)
	L: for {
		switch {
		case isNumber[x.peek]:
		    c = x.next()
			add(&b, c)
		default:
			break L
		}
	}
    var err error
	yylval.num, err = strconv.ParseUint(b.String(), 10, 64)
	if err != nil {
		log.Printf("ERRtokOR: %s. Bad number %q", err, b.String())
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

func NewLexer(line []byte, n uint64) *yyLex {
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

func Evaluate(expression string, n uint64) (uint64, error) {
    yyErrorVerbose = true
    l := NewLexer([]byte(expression), n)
    yyParse(l)
    return l.Result, l.Err
}
