package transformer

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
	"github.com/k1LoW/octoslack/config"
)

var (
	ErrNoneOfConditionsMet = errors.New("none of conditions met")
	nlRep                  = strings.NewReplacer("\r\n", "\n")
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
	var buf bytes.Buffer
	if t.config == nil || len(t.config.Requests) == 0 {
		return nil, ErrNoneOfConditionsMet
	}
	if _, err := buf.ReadFrom(req.Body); err != nil {
		return nil, err
	}
	if err := req.Body.Close(); err != nil {
		return nil, err
	}
	if len(buf.Bytes()) == 0 {
		return nil, errors.New("empty payload")
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		return nil, err
	}
	env := map[string]interface{}{
		"github_event": req.Header.Get("X-GitHub-Event"),
		"method":       req.Method,
		"headers":      req.Header,
		"payload":      payload,
		"quote":        quote,
	}
	for _, e := range t.config.Requests {
		tf, err := evalCond(e.Condition, env)
		if err != nil {
			return nil, err
		}
		if !tf {
			continue
		}
		v, err := evalExpand(e.Transform, env)
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
		req.Header.Add("Content-Type", "application/json")
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
		return nlRep.Replace(out), nil
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

func quote(v interface{}) string {
	lines := strings.Split(v.(string), "\n")
	quoted := []string{}
	for _, l := range lines {
		ql := fmt.Sprintf("> %s", l)
		if ql == "> " {
			ql = ">"
		}
		quoted = append(quoted, ql)
	}
	return strings.Join(quoted, "\n")
}
