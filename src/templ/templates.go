package templ

import (
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
	RenderRoomBtn      = renderFunc[RoomButton]("sidebar/room-btn")
	RenderNewRoomError = renderFunc[string]("sidebar/new-room-error")
)
