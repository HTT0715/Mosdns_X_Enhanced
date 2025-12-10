/*
 * Copyright (C) 2020-2022, IrineSistiana
 *
 * This file is part of mosdns.
 *
 * mosdns is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * mosdns is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package doh_whitelist

import (
	"context"
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"go.uber.org/zap"

	"github.com/pmkol/mosdns-x/coremain"
	"github.com/pmkol/mosdns-x/pkg/dnsutils"
	"github.com/pmkol/mosdns-x/pkg/executable_seq"
	"github.com/pmkol/mosdns-x/pkg/matcher/netlist"
	"github.com/pmkol/mosdns-x/pkg/query_context"
)

const PluginType = "doh_whitelist"

func init() {
	coremain.RegNewPluginFunc(PluginType, Init, func() interface{} { return new(Args) })
}

var _ coremain.ExecutablePlugin = (*whitelist)(nil)

type Args struct {
	Whitelist    []string `yaml:"whitelist"`     // IP addresses or CIDR ranges
	PathList     []string `yaml:"path_list"`     // Allowed URL paths (e.g., /dns-query/token123)
	RCode        int      `yaml:"rcode"`         // Response code when client is not in whitelist, default is REFUSED
	RequireBoth  bool     `yaml:"require_both"`  // If true, both IP and path must match; if false, either one matches (default: false)
}

type whitelist struct {
	*coremain.BP
	ipMatcher   *netlist.MatcherGroup
	pathList    map[string]struct{} // Set of allowed paths
	rcode       int
	requireBoth bool
}

func Init(bp *coremain.BP, args interface{}) (p coremain.Plugin, err error) {
	return newWhitelist(bp, args.(*Args))
}

func newWhitelist(bp *coremain.BP, args *Args) (*whitelist, error) {
	rcode := args.RCode
	if rcode == 0 {
		rcode = dns.RcodeRefused
	}

	var ipMatcher *netlist.MatcherGroup
	if len(args.Whitelist) > 0 {
		// Load IP whitelist from configuration
		mg, err := netlist.BatchLoadProvider(args.Whitelist, bp.M().GetDataManager())
		if err != nil {
			return nil, fmt.Errorf("failed to load IP whitelist: %w", err)
		}
		ipMatcher = mg
		bp.L().Info("doh IP whitelist loaded", zap.Int("count", mg.Len()))
	}

	// Load path whitelist
	pathList := make(map[string]struct{})
	for _, path := range args.PathList {
		// Normalize path: remove trailing slash, ensure leading slash
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		path = strings.TrimSuffix(path, "/")
		if path == "" {
			path = "/"
		}
		pathList[path] = struct{}{}
	}
	if len(pathList) > 0 {
		bp.L().Info("doh path whitelist loaded", zap.Int("count", len(pathList)))
	}

	// Check if at least one whitelist is configured
	if ipMatcher == nil && len(pathList) == 0 {
		return nil, fmt.Errorf("at least one of 'whitelist' or 'path_list' must be configured")
	}

	return &whitelist{
		BP:          bp,
		ipMatcher:   ipMatcher,
		pathList:    pathList,
		rcode:       rcode,
		requireBoth: args.RequireBoth,
	}, nil
}

// Exec checks if the DoH client is in the whitelist (IP or path).
// If the request is not DoH, it passes through.
// If the client is not in whitelist, it returns a refused response.
func (w *whitelist) Exec(ctx context.Context, qCtx *query_context.Context, next executable_seq.ExecutableChainNode) error {
	// Check if this is a DoH request
	protocol := qCtx.ReqMeta().GetProtocol()
	isDoH := protocol == query_context.ProtocolHTTPS ||
		protocol == query_context.ProtocolH2 ||
		protocol == query_context.ProtocolH3

	// If not DoH, pass through
	if !isDoH {
		return executable_seq.ExecChainNode(ctx, qCtx, next)
	}

	// Check IP whitelist
	ipMatched := false
	if w.ipMatcher != nil {
		clientAddr := qCtx.ReqMeta().GetClientAddr()
		if clientAddr.IsValid() {
			matched, err := w.ipMatcher.Match(clientAddr)
			if err != nil {
				w.L().Warn("failed to match client address", zap.Stringer("addr", clientAddr), zap.Error(err))
			} else {
				ipMatched = matched
			}
		}
	} else {
		// If no IP whitelist configured, consider IP check as passed (when requireBoth is false)
		ipMatched = !w.requireBoth
	}

	// Check path whitelist
	pathMatched := false
	if len(w.pathList) > 0 {
		requestPath := qCtx.ReqMeta().GetPath()
		// Normalize path for comparison
		requestPath = strings.TrimSuffix(requestPath, "/")
		if requestPath == "" {
			requestPath = "/"
		}
		_, pathMatched = w.pathList[requestPath]
	} else {
		// If no path whitelist configured, consider path check as passed (when requireBoth is false)
		pathMatched = !w.requireBoth
	}

	// Determine if request should be allowed
	allowed := false
	if w.requireBoth {
		// Both IP and path must match
		allowed = ipMatched && pathMatched
	} else {
		// Either IP or path matches
		allowed = ipMatched || pathMatched
	}

	if !allowed {
		// Request is not allowed, reject
		r := dnsutils.GenEmptyReply(qCtx.Q(), w.rcode)
		qCtx.SetResponse(r)
		return nil
	}

	// Request is allowed, continue
	return executable_seq.ExecChainNode(ctx, qCtx, next)
}

func (w *whitelist) Close() error {
	if w.ipMatcher != nil {
		return w.ipMatcher.Close()
	}
	return nil
}

