package rmcomment

import (
	"io/ioutil"
	"os"

	"github.com/modern-go/parse"
)

func StringRm(content string) string {
	src := parse.NewSourceString(content)
	ans := parse.Parse(src, NewCommentLexer(), 0)
	if ans != nil {
		return string(ans.([]byte))
	}
	return ""
}

func FileRm(f *os.File) (string, error) {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return StringRm(string(data)), nil
}

func PathRm(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	return FileRm(f)
}
