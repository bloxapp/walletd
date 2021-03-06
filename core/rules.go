package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog/log"
	"github.com/shibukawa/configdir"
)

// RulesResult represents the result of running a set of rules.
type RulesResult int

const (
	UNKNOWN RulesResult = iota
	APPROVED
	DENIED
	FAILED
)

// RuleDefinition defines a rule.
type RuleDefinition struct {
	Name    string `json:"name"`
	Request string `json:"request"`
	Account string `json:"account"`
	Script  string `json:"script"`
}

// InitRules initialises the rules from a configuration.
func InitRules(ctx context.Context, defs []*RuleDefinition) ([]*Rule, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "core.InitRules")
	defer span.Finish()

	rules := make([]*Rule, 0, len(defs))
	for _, def := range defs {
		rule, err := NewRule(def)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// Rule contains a ready-to-run rule script.
type Rule struct {
	name    string
	request string
	account string
	script  string
}

// Name returns the name for the rule.
func (r *Rule) Name() string {
	return r.name
}

// Script returns the script for the rule.
func (r *Rule) Script() string {
	return r.script
}

// Matches returns true if this rule matches the path.
func (r *Rule) Matches(request string, account string) bool {
	if r.request != request {
		return false
	}
	res, err := regexp.Match(r.account, []byte(account))
	if err != nil {
		log.Warn().Err(err).
			Str("request", request).
			Str("account", account).
			Str("rule", r.name).
			Str("ruleaccount", r.account).
			Msg("Match attempt failed")
		return false
	}
	return res
}

// NewRule creates a new rule from its definition.
func NewRule(def *RuleDefinition) (*Rule, error) {
	configDirs := configdir.New("wealdtech", "walletd")
	configPath := configDirs.QueryFolders(configdir.Global)[0].Path
	path := filepath.Join(configPath, "scripts", def.Script)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if def.Account == "" {
		def.Account = "^.*$"
	}
	if !strings.HasPrefix(def.Account, "^") {
		def.Account = fmt.Sprintf("^%s", def.Account)
	}
	if !strings.HasSuffix(def.Account, "$") {
		def.Account = fmt.Sprintf("%s$", def.Account)
	}
	return &Rule{
		name:    def.Name,
		request: def.Request,
		account: def.Account,
		script:  string(contents),
	}, nil
}
