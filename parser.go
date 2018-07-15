package rmcomment

import (
	"bytes"
	"io"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/read"
)

const (
	OriContent = 1
	OriLine    = 2
	OriMLine   = 3
	OriQuote   = 4
)

type commentLexer struct {
	content *contentToken
	line    *lineCommentToken
	mline   *mlineCommentToken
	quote   *quoteToken
}

var lexer = NewCommentLexer()

func NewCommentLexer() *commentLexer {
	return &commentLexer{
		quote: &quoteToken{},
	}
}

func (lexer *commentLexer) Parse(src *parse.Source, precedence int) interface{} {
	return parse.Parse(src, lexer, precedence)
}

func (lexer *commentLexer) PrefixToken(src *parse.Source) parse.PrefixToken {
	if src.Error() == io.EOF {
		return nil
	}
	switch src.Peek1() {
	case '#':
		return lexer.line
	case '/':
		switch string(src.PeekN(2)) {
		case "//":
			return lexer.line
		case "/*":
			return lexer.mline
		}
	case '"':
		lexer.quote.c = '"'
		return lexer.quote
	case '\'':
		lexer.quote.c = '\''
		return lexer.quote
	}
	return lexer.content
}

func (lexer *commentLexer) InfixToken(src *parse.Source) (parse.InfixToken, int) {
	if src.Error() == io.EOF {
		return nil, 0
	}
	switch src.Peek1() {
	case '#':
		return lexer.line, OriLine
	case '/':
		switch string(src.PeekN(2)) {
		case "//":
			return lexer.line, OriLine
		case "/*":
			return lexer.mline, OriMLine
		}
	case '"':
		lexer.quote.c = '"'
		return lexer.quote, OriQuote
	case '\'':
		lexer.quote.c = '\''
		return lexer.quote, OriQuote
	}
	return lexer.content, OriContent
}

type contentToken struct {
}

func (token *contentToken) readOneContent(src *parse.Source, old []byte) []byte {
	for {
		data := AnyExcept(src, []byte{'#', '/', '"', '\''})
		old = append(old, data...)
		if src.Error() == io.EOF {
			break
		}
		if src.Peek1() == '#' || src.Peek1() == '"' || src.Peek1() == '\'' {
			break
		}
		switch string(src.PeekN(2)) {
		case "//":
			return old
		case "/*":
			return old
		}
		old = append(old, src.Read1())
	}
	return old
}

func (token *contentToken) PrefixParse(src *parse.Source) interface{} {
	//fmt.Println("content prefix")
	return token.readOneContent(src, []byte{})
}

func (token *contentToken) InfixParse(src *parse.Source, left interface{}) interface{} {
	//fmt.Println("content infix")
	res := token.readOneContent(src, []byte{})
	right := lexer.Parse(src, OriContent)
	return Combine(left, res, right)
}

type lineCommentToken struct {
}

func (token *lineCommentToken) readOneLine(src *parse.Source) {
	read.AnyExcept1(src, '\n')
	// keep the '\n' for headers

	//if src.Error() == io.EOF {
	//    return
	//}
	//src.Read1()
}

func (token *lineCommentToken) PrefixParse(src *parse.Source) interface{} {
	//fmt.Println("line prefix")
	token.readOneLine(src)
	return nil
}

func (token *lineCommentToken) InfixParse(src *parse.Source, left interface{}) interface{} {
	//fmt.Println("line infix")
	token.readOneLine(src)
	right := lexer.Parse(src, OriLine)
	return Combine(left, nil, right)
}

type mlineCommentToken struct {
}

func (token *mlineCommentToken) readOneMLine(src *parse.Source) {
	for {
		read.AnyExcept1(src, '*')
		if src.Error() == io.EOF {
			return
		}
		src.Read1()
		if src.Peek1() == '/' {
			src.Read1()
			return
		}
	}
}

func (token *mlineCommentToken) PrefixParse(src *parse.Source) interface{} {
	//fmt.Println("mline prefix")
	src.ReadN(2)
	token.readOneMLine(src)
	return nil
}

func (token *mlineCommentToken) InfixParse(src *parse.Source, left interface{}) interface{} {
	//fmt.Println("mline infix")
	token.readOneMLine(src)
	right := lexer.Parse(src, OriLine)
	return Combine(left, nil, right)
}

type quoteToken struct {
	c byte
}

func (token *quoteToken) readOneQuote(src *parse.Source, old []byte) []byte {
	old = append(old, src.Read1())
	for {
		r := read.AnyExcept1(src, token.c)
		old = append(old, r...)
		if len(r) > 0 && r[len(r)-1] == '\\' {
			continue
		}
		if src.Error() == io.EOF {
			break
		}
		old = append(old, src.Read1())
		break
	}
	return old
}

func (token *quoteToken) PrefixParse(src *parse.Source) interface{} {
	//fmt.Println("quote prefix")
	return token.readOneQuote(src, []byte{})
}

func (token *quoteToken) InfixParse(src *parse.Source, left interface{}) interface{} {
	//fmt.Println("quote infix")
	if src.Error() == io.EOF {
		return left
	}
	res := token.readOneQuote(src, []byte{})
	right := lexer.Parse(src, OriLine)
	return Combine(left, res, right)
}

func Combine(left interface{}, mid []byte, right interface{}) []byte {
	var res []byte
	if left != nil {
		res = append(res, left.([]byte)...)
	}
	if mid != nil {
		res = append(res, mid...)
	}
	if right != nil {
		res = append(res, right.([]byte)...)
	}
	return res
}

func AnyExcept(src *parse.Source, barr []byte) []byte {
	space := src.ClaimSpace()
	for src.Error() == nil {
		buf := src.Peek()
		for i := 0; i < len(buf); i++ {
			b := buf[i]
			if bytes.Contains(barr, []byte{b}) {
				space = append(space, buf[:i]...)
				src.ConsumeN(i)
				return space
			}
		}
		space = append(space, buf...)
		src.Consume()
	}
	return space
}
