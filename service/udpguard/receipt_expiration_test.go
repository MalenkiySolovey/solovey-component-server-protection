package udpguard

import (
	"testing"

	protectionrepository "github.com/MalenkiySolovey/solovey-component-server-protection/service/durable"
)

func TestReceiptExpiredReceiptHasExplicitUDPContractCode(t *testing.T) {
	if code := ErrorCode(udpReceiptError(protectionrepository.ErrIdempotencyKeyExpired)); code != CodeIdempotencyExpired {
		t.Fatalf("expired receipt code=%q", code)
	}
}
