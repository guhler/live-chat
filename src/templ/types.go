package templ

// pages
type RoomsPage []struct {
	RoomName string
}

type RoomPage struct {
	RoomName  string
	Sidebar   Sidebar
	WsUrl     string
	Messages  []Message
	Done      bool
	NextStart int
}

// components
type Sidebar []RoomButton

type RoomButton struct {
	RoomName string
	Selected bool
}

type SwitchRoom struct {
	RoomName    string
	ChatContent MessageList
}

type MessageList struct {
	RoomName  string
	Messages  []Message
	Done      bool
	NextStart int
}

type Message struct {
	UserName string
	Content  string
	IsOwn    bool
}
