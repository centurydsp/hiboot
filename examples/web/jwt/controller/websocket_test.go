package controller

import (
	"github.com/stretchr/testify/assert"
	"hidevops.io/hiboot/pkg/app/web"
	"hidevops.io/hiboot/pkg/starter/websocket"
	"net/http"
	"testing"
)

func TestWebSocketController(t *testing.T) {
	mockController := newWebsocketController(func(handler websocket.Handler, conn *websocket.Connection) {
		// For controller's unit testing, do nothing
	})

	testApp := web.NewTestApp(mockController).Run(t)
	assert.NotEqual(t, nil, testApp)

	token := genJwtToken(1000)

	testApp.Get("/websocket").
		WithHeader("Authorization", token).
		Expect().Status(http.StatusOK)

	testApp.Get("/websocket/status").
		WithHeader("Authorization", token).
		Expect().Status(http.StatusOK)
}
