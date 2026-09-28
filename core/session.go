package llmrouter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type sessionKey struct{}

func WithSessionID(ctx context.Context, id string) context.Context {
	if id = strings.TrimSpace(id); id == "" {
		return ctx
	}
	return context.WithValue(ctx, sessionKey{}, id)
}

func SessionID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(sessionKey{}).(string)
	return id
}

func SessionUUID(ctx context.Context) string {
	id := SessionID(ctx)
	if id == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(id))
	raw := hex.EncodeToString(sum[:])
	return raw[0:8] + "-" + raw[8:12] + "-8" + raw[13:16] + "-a" + raw[17:20] + "-" + raw[20:32]
}
