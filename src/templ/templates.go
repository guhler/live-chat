package templ

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func renderFunc[D any](templateName string) func(echo.Context, int, D) error {
	return func(c echo.Context, status int, data D) error {
		return c.Render(status, templateName, data)
	}
}

var (
	RenderRoomsPage    = renderFunc[RoomsPage]("rooms.html")
	RenderRoomPage     = renderFunc[RoomPage]("room.html")
	RenderSwitchRoom   = renderFunc[SwitchRoom]("room/switch-room")
	RenderMessageList  = renderFunc[MessageList]("room/message-list")
	RenderNewRoomError = renderFunc[string]("sidebar/new-room-error")
	RenderRoomBtn      = func(c echo.Context, status int, data RoomButton) error {
		if strings.HasSuffix(data.CurrentUrl, "/rooms") {
			return c.Render(status, "rooms/li", data)
		}
		return c.Render(status, "sidebar/room-btn", data)
	}
)
