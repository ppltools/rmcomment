package rmcomment

import (
	"context"
	"testing"

	"github.com/modern-go/parse"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
)

func TestCommentLexer_Parse(t *testing.T) {
	t.Run("overflow", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h")
		must.Equal([]byte{'h'}, must.Call(src.PeekN, 2)[0].([]byte))
		src = parse.NewSourceString("hh")
		must.Equal([]byte{'h', 'h'}, must.Call(src.PeekN, 2)[0].([]byte))
	}))

	t.Run("no quote", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString(`# ok
name ok// hello
/*
 * dk
 */
my name is //lilei!`)
		parsed := parse.Parse(src, NewCommentLexer(), 0)
		must.Equal(`
name ok

my name is `, string(parsed.([]byte)))

	}))

	t.Run("quote", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString(`/* ok
name ok // hello
*
 * dk
 */
my name is//lilei!
/*
 * test
 */
aaa // d`)
		parsed := parse.Parse(src, NewCommentLexer(), 0)
		must.Equal(`
my name is

aaa `, string(parsed.([]byte)))
	}))
}
