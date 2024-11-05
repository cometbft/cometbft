package log_test

import (
	"bytes"
	stderr "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/cometbft/cometbft/libs/log"
)

func TestTracingLogger(t *testing.T) {
	var buf bytes.Buffer

	logger := log.NewJSONLoggerNoTS(&buf)

	logger1 := log.NewTracingLogger(logger)
	err1 := errors.New("courage is grace under pressure")
	err2 := errors.New("it does not matter how slowly you go, so long as you do not stop")
	logger1.With("err1", err1).Info("foo", "err2", err2)

	want := strings.ReplaceAll(
		strings.ReplaceAll(
			`{"level":"INFO","msg":"foo","err1":"`+
				fmt.Sprintf("%+v", err1)+
				`","err2":"`+
				fmt.Sprintf("%+v", err2)+
				`"}`,
			"\t", "",
		), "\n", "")
	have := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(buf.String()), "\\n", ""), "\\t", "")
	if want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}

	buf.Reset()

	logger.With(
		"err1", stderr.New("opportunities don't happen. You create them"),
	).Info(
		"foo", "err2", stderr.New("once you choose hope, anything's possible"),
	)

	want = `{"level":"INFO","msg":"foo","err1":"opportunities don't happen. You create them","err2":"once you choose hope, anything's possible"}`
	have = strings.TrimSpace(buf.String())
	if want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}

	buf.Reset()

	logger.With("user", "Sam").With("context", "value").Info("foo", "bar", "baz")

	want = `{"level":"INFO","msg":"foo","user":"Sam","context":"value","bar":"baz"}`
	have = strings.TrimSpace(buf.String())
	if want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}
}
