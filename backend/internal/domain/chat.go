package domain

// ChatSource identifies where an OPSBot answer came from (CU-09 CA-1).
type ChatSource struct {
	Type  string
	ID    string
	Title string
}

// ChatAnswer is OPSBot's reply.
type ChatAnswer struct {
	Answer  string
	Sources []ChatSource
}

// ChatQuestion is a message sent to OPSBot on behalf of an authenticated user.
type ChatQuestion struct {
	SessionID string
	Message   string
	UserID    string
	UserToken string
}
