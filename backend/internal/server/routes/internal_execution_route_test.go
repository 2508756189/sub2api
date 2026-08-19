package routes

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestInternalExecutionRouteDisabledByDefault(t *testing.T) {
	router := newGatewayRoutesTestRouter()
	for _, route := range router.Routes() {
		require.NotEqual(t, "/internal/v1/model-executions/chat/completions", route.Path)
	}
}

func TestInternalExecutionRouteRegisteredOnlyWhenConfigured(t *testing.T) {
	router := newGatewayRoutesTestRouterWithConfig(&config.Config{
		Gateway: config.GatewayConfig{
			MaxBodySize:       1024 * 1024,
			TextMaxBodySize:   1024 * 1024,
			InternalExecution: config.GatewayInternalExecutionConfig{Enabled: true, Token: "dsh-secret"},
		},
	})

	found := false
	for _, route := range router.Routes() {
		if route.Method == http.MethodPost && route.Path == "/internal/v1/model-executions/chat/completions" {
			found = true
			break
		}
	}
	require.True(t, found)
}
