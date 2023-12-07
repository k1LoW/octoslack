package transformer

import (
	"bytes"
	"errors"
	"net/http"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
	"github.com/k1LoW/octoslack/config"
	"github.com/spf13/cast"
	"golang.org/x/exp/slog"
)

var (
	ErrNoneOfConditionsMet = errors.New("none of conditions met")
	ErrDropAction          = errors.New("met condition to drop")
	crRep                  = strings.NewReplacer("\r", "")
)

type Transformer struct {
	config *config.Config
}

func New(cfg *config.Config) *Transformer {
	return &Transformer{
		config: cfg,
	}
}

func (t *Transformer) Transform(req *http.Request) (*http.Request, error) {
	const slackHost = "hooks.slack.com"
	var body bytes.Buffer
	if t.config == nil || len(t.config.Requests) == 0 {
		return nil, ErrNoneOfConditionsMet
	}
	if _, err := body.ReadFrom(req.Body); err != nil {
		return nil, err
	}
	if err := req.Body.Close(); err != nil {
		return nil, err
	}
	if len(body.Bytes()) == 0 {
		return nil, errors.New("empty payload")
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(body.Bytes(), &payload); err != nil {
		return nil, err
	}
	env := map[string]interface{}{
		"github_event": req.Header.Get("X-GitHub-Event"),
		"method":       req.Method,
		"headers":      req.Header,
		"path":         req.URL.Path,
		"payload":      payload,
		// built-in funcs
		"quote":            quote,
		"quote_md":         quoteMarkdown,
		"shorten_lines":    shortenLines,
		"shorten_lines_md": shortenLinesMarkdown,
		"string":           cast.ToString,
	}
	for _, r := range t.config.Requests {
		tf, err := evalCond(r.Condition, env)
		if err != nil {
			slog.Error("Failed to eval condition", err)
			continue
		}
		if !tf {
			continue
		}
		switch r.Action {
		case config.ForwardAction:
			u := req.URL
			u.Host = slackHost
			u.Scheme = "https"
			nreq, err := http.NewRequest(req.Method, u.String(), bytes.NewReader(body.Bytes()))
			if err != nil {
				return nil, err
			}
			if req.Header.Get("Host") != "" {
				req.Header.Set("Host", slackHost)
			}
			nreq.Header = req.Header
			return nreq, nil
		case config.DropAction:
			return nil, ErrDropAction
		}
		// config.TransformAction
		v, err := evalExpand(r.Transform, env)
		if err != nil {
			return nil, err
		}
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		u := req.URL
		u.Host = slackHost
		u.Scheme = "https"
		req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}
	return nil, ErrNoneOfConditionsMet
}

func evalCond(cond string, env map[string]interface{}) (bool, error) {
	v, err := expr.Eval(cond, env)
	if err != nil {
		return false, err
	}
	switch vv := v.(type) {
	case bool:
		return vv, nil
	default:
		return false, nil
	}
}

func evalExpand(tmpl, env map[string]interface{}) (map[string]interface{}, error) {
	const (
		delimStart = "{{"
		delimEnd   = "}}"
	)
	// Expand using expand.ExprRepFn
	b, err := yaml.Marshal(tmpl)
	if err != nil {
		return nil, err
	}
	e, err := expand.ReplaceYAML(string(b), func(in string) (string, error) {
		repfn := expand.ExprRepFn(delimStart, delimEnd, env)
		out, err := repfn(in)
		if err != nil {
			return "", err
		}
		return crRep.Replace(out), nil
	}, true)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := yaml.Unmarshal([]byte(e), &out); err != nil {
		return nil, err
	}
	return out, nil
}
