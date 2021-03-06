package signer

import (
	"github.com/wealdtech/walletd/services/autounlocker"
	"github.com/wealdtech/walletd/services/checker"
	"github.com/wealdtech/walletd/services/fetcher"
	"github.com/wealdtech/walletd/services/ruler"
)

// Handler is the signer handler.
type Handler struct {
	checker      checker.Service
	fetcher      fetcher.Service
	ruler        ruler.Service
	autounlocker autounlocker.Service
}

// New creates a new signer handler.
func New(autounlocker autounlocker.Service, checker checker.Service, fetcher fetcher.Service, ruler ruler.Service) *Handler {
	return &Handler{
		autounlocker: autounlocker,
		checker:      checker,
		fetcher:      fetcher,
		ruler:        ruler,
	}
}
