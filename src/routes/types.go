package routes

type message struct {
	UserName string
	Content  string
	IsOwn    bool
}

// pages
type roomPage struct {
	RoomName  string
	Sidebar   sidebar
	WsUrl     string
	Messages  []message
	Done      bool
	NextStart int
}

// responses
type messageResponse struct {
	RoomName    string
	Selected    bool
	ChatContent chatContent
}

type postResponse struct {
	RoomName string
}

// components
type sidebar []roomButton

type roomButton struct {
	RoomName string
	Selected bool
}

type chatContent struct {
	RoomName  string
	Messages  []message
	Done      bool
	NextStart int
}

type roomResponse struct {
	RoomId    int64
	Done      bool
	NextStart int
	Messages  []message
}

type indexPage struct {
	RoomNames []string
}
