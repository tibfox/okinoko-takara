package main

import (
	"okinoko_lottery/sdk"
	"strconv"
	"strings"
)

// unwrapPayload trims quotes and whitespace, aborting if the payload is empty.
func unwrapPayload(payload *string, errMsg string) string {
	if payload == nil {
		sdk.Abort(errMsg)
	}
	raw := strings.TrimSpace(*payload)
	if raw == "" {
		sdk.Abort(errMsg)
	}
	if len(raw) >= 2 {
		first := raw[0]
		last := raw[len(raw)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			if unquoted, err := strconv.Unquote(raw); err == nil {
				return unquoted
			}
			raw = strings.TrimSpace(raw[1 : len(raw)-1])
			if raw == "" {
				sdk.Abort(errMsg)
			}
		}
	}
	return raw
}
