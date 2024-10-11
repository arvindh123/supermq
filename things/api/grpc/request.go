package grpc

type authorizeReq struct {
	ThingID    string
	ThingKey   string
	ChannelID  string
	Permission string
}
