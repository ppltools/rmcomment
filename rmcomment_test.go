package rmcomment

import (
	"context"
	"testing"

	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
)

func TestStringRm(t *testing.T) {
	t.Run("StringRm", test.Case(func(ctx context.Context) {
		ori := "hello, # world"
		must.Equal("hello, ", StringRm(ori))
	}))
	t.Run("PathRm", test.Case(func(ctx context.Context) {
		path := "rmcomment_test.go"
		_, err := PathRm(path)
		must.Equal(nil, err)
	}))
}
