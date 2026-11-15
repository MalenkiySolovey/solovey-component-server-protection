package localproxy

import (
	"testing"

	protectionrepository "github.com/MalenkiySolovey/solovey-component-server-protection/service/durable"
)

func TestReceiptExpiredReceiptHasExplicitLocalProxyContractCode(t *testing.T) {
	if code := ErrorCode(localProxyReceiptError(protectionrepository.ErrIdempotencyKeyExpired)); code != CodeIdempotencyExpired {
		t.Fatalf("expired receipt code=%q", code)
	}
}
