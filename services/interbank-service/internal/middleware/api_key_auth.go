package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/interbank-service/internal/service"
)

// PeerContextKey is the gin context key under which the resolved peer's
// routing number is stored after successful authentication.
const PeerContextKey = "interbank.peer_routing_number"

// APIKeyAuth verifies the X-Api-Key header against the configured peer
// registry. On success it stashes the peer's routing number on the request
// context for downstream handlers; on failure it short-circuits with 401.
func APIKeyAuth(peers *service.PeerResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Api-Key")
		if key == "" {
			_ = c.Error(errors.UnauthorizedErr("missing X-Api-Key header"))
			c.Abort()
			return
		}

		peer, ok := peers.ByTheirAPIKey(key)
		if !ok {
			_ = c.Error(errors.UnauthorizedErr("unknown X-Api-Key"))
			c.Abort()
			return
		}

		c.Set(PeerContextKey, peer.RoutingNumber)
		c.Next()
	}
}
