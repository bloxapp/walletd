package static

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/wealdtech/walletd/core"
)

// StaticChecker checks against a static list.
type StaticChecker struct {
	access map[string][]*path
}

type path struct {
	wallet     *regexp.Regexp
	account    *regexp.Regexp
	operations []string
}

// New creates a new static checker.
func New(ctx context.Context, config *core.Permissions) (*StaticChecker, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "checker.static.New")
	defer span.Finish()

	if config == nil {
		return nil, errors.New("certificate info is required")
	}
	if config.Certs == nil {
		return nil, errors.New("certificates are required")
	}
	if len(config.Certs) == 0 {
		return nil, errors.New("certificate info empty")
	}

	access := make(map[string][]*path, len(config.Certs))
	for _, certificateInfo := range config.Certs {
		if certificateInfo.Name == "" {
			return nil, errors.New("certificate info requires a name")
		}
		if len(certificateInfo.Perms) == 0 {
			return nil, errors.New("certificate info requires at least one permission")
		}
		paths := make([]*path, len(certificateInfo.Perms))
		for i, permissions := range certificateInfo.Perms {
			if permissions.Path == "" {
				return nil, errors.New("permission path cannot be blank")
			}
			walletName, accountName, err := walletAndAccountNamesFromPath(permissions.Path)
			if err != nil {
				return nil, fmt.Errorf("invalid account path %s", permissions.Path)
			}
			if walletName == "" {
				return nil, errors.New("wallet cannot be blank")
			}
			walletRegex, err := regexify(walletName)
			if err != nil {
				return nil, fmt.Errorf("invalid wallet regex %s", walletName)
			}
			accountRegex, err := regexify(accountName)
			if err != nil {
				return nil, fmt.Errorf("invalid account regex %s", accountName)
			}
			paths[i] = &path{
				wallet:     walletRegex,
				account:    accountRegex,
				operations: permissions.Operations,
			}
		}
		access[certificateInfo.Name] = paths
	}
	return &StaticChecker{
		access: access,
	}, nil
}

// Check checks the client to see if the account is allowed.
func (c *StaticChecker) Check(ctx context.Context, client string, account string, operation string) bool {
	span, _ := opentracing.StartSpanFromContext(ctx, "checker.static.Check")
	defer span.Finish()

	if client == "" {
		log.Info().Msg("No client certificate name")
		return false
	}
	log := log.With().Str("client", client).Str("account", account).Logger()

	walletName, accountName, err := walletAndAccountNamesFromPath(account)
	if err != nil {
		log.Debug().Err(err).Msg("Invalid path")
		return false
	}
	if walletName == "" {
		log.Debug().Err(err).Msg("Missing wallet name")
		return false
	}
	if accountName == "" {
		log.Debug().Err(err).Msg("Missing account name")
		return false
	}

	paths, exists := c.access[client]
	if !exists {
		log.Debug().Msg("Unknown client")
		return false
	}

	for _, path := range paths {
		if path.wallet.Match([]byte(walletName)) && path.account.Match([]byte(accountName)) {
			for i := range path.operations {
				if path.operations[i] == "All" || path.operations[i] == operation {
					return true
				}
			}
		}
	}
	return false
}

// walletAndAccountNamesFromPath is a helper that breaks out a path's components.
func walletAndAccountNamesFromPath(path string) (string, string, error) {
	if len(path) == 0 {
		return "", "", errors.New("invalid account format")
	}
	index := strings.Index(path, "/")
	if index == -1 {
		// Just the wallet
		return path, "", nil
	}
	if index == len(path)-1 {
		// Trailing /
		return path[:index], "", nil
	}
	return path[:index], path[index+1:], nil
}

func regexify(name string) (*regexp.Regexp, error) {
	// Empty equates to all.
	if name == "" {
		name = ".*"
	}
	// Anchor if required.
	if !strings.HasPrefix(name, "^") {
		name = fmt.Sprintf("^%s", name)
	}
	if !strings.HasSuffix(name, "$") {
		name = fmt.Sprintf("%s$", name)
	}

	return regexp.Compile(name)

}
