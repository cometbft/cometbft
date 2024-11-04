package log_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cometbft/cometbft/libs/log"
)

func TestVariousLevels(t *testing.T) {
	testCases := []struct {
		name    string
		allowed log.Option
		want    string
	}{
		{
			"AllowAll",
			log.AllowAll(),
			strings.Join([]string{
				`{"level":"DEBUG","msg":"here","this is":"debug log"}`,
				`{"level":"INFO","msg":"here","this is":"info log"}`,
				`{"level":"WARN","msg":"here","this is":"warn log"}`,
				`{"level":"ERROR","msg":"here","this is":"error log"}`,
			}, "\n"),
		},
		{
			"AllowError",
			log.AllowError(),
			strings.Join([]string{
				`{"level":"ERROR","msg":"here","this is":"error log"}`,
			}, "\n"),
		},
		{
			"AllowInfo",
			log.AllowInfo(),
			strings.Join([]string{
				`{"level":"INFO","msg":"here","this is":"info log"}`,
				`{"level":"WARN","msg":"here","this is":"warn log"}`,
				`{"level":"ERROR","msg":"here","this is":"error log"}`,
			}, "\n"),
		},
		{
			"AllowDebug",
			log.AllowDebug(),
			strings.Join([]string{
				`{"level":"DEBUG","msg":"here","this is":"debug log"}`,
				`{"level":"INFO","msg":"here","this is":"info log"}`,
				`{"level":"WARN","msg":"here","this is":"warn log"}`,
				`{"level":"ERROR","msg":"here","this is":"error log"}`,
			}, "\n"),
		},
		{
			"AllowNone",
			log.AllowNone(),
			``,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := log.NewFilter(log.NewJSONLoggerNoTS(&buf), tc.allowed)

			logger.Debug("here", "this is", "debug log")
			logger.Info("here", "this is", "info log")
			logger.Warn("here", "this is", "warn log")
			logger.Error("here", "this is", "error log")

			if want, have := tc.want, strings.TrimSpace(buf.String()); want != have {
				t.Errorf("\nwant:\n%s\nhave:\n%s", want, have)
			}
		})
	}
}

func TestLevelContext(t *testing.T) {
	var buf bytes.Buffer

	logger := log.NewJSONLoggerNoTS(&buf)
	logger = log.NewFilter(logger, log.AllowError())
	logger = logger.With("context", "value")

	logger.Error("foo", "bar", "baz")

	want := `{"level":"ERROR","msg":"foo","context":"value","bar":"baz"}`
	have := strings.TrimSpace(buf.String())
	if want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}

	buf.Reset()
	logger.Info("foo", "bar", "baz")
	if want, have := ``, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}
}

func TestVariousAllowWith(t *testing.T) {
	var buf bytes.Buffer

	logger := log.NewJSONLoggerNoTS(&buf)

	logger1 := log.NewFilter(logger, log.AllowError(), log.AllowInfoWith("context", "value"))
	logger1.With("context", "value").Info("foo", "bar", "baz")

	want := `{"level":"INFO","msg":"foo","context":"value","bar":"baz"}`
	have := strings.TrimSpace(buf.String())
	if want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}

	buf.Reset()

	logger2 := log.NewFilter(
		logger,
		log.AllowError(),
		log.AllowInfoWith("context", "value"),
		log.AllowNoneWith("user", "Sam"),
	)

	logger2.With("context", "value", "user", "Sam").Info("foo", "bar", "baz")
	if want, have := ``, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}

	buf.Reset()

	logger3 := log.NewFilter(
		logger,
		log.AllowError(),
		log.AllowInfoWith("context", "value"),
		log.AllowNoneWith("user", "Sam"),
	)

	logger3.With("user", "Sam").With("context", "value").Info("foo", "bar", "baz")

	want = `{"level":"INFO","msg":"foo","user":"Sam","context":"value","bar":"baz"}`
	have = strings.TrimSpace(buf.String())
	if want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}
}
