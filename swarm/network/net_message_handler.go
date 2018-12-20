package network

func ComposeChallengeMessage(messageId string, challenge string) *NetMessage {

	message := new(NetMessage)
	message.MessageId = messageId
	message.MessageType = NetMessageType_Challenge

	challengeMessage := new(ChallengeMessage)
	challengeMessage.Challenge = challenge


	message.Data = &NetMessage_ChallengeMessage{ChallengeMessage: challengeMessage}

	return message
}

func ComposeEventMessage(messageId string, eventData []byte) *NetMessage {

	message := new(NetMessage)
	message.MessageId = messageId
	message.MessageType = NetMessageType_Event

	eventMessage := new(EventMessage)
	eventMessage.Data = eventData


	message.Data = &NetMessage_EventMessage{EventMessage: eventMessage}

	return message

}