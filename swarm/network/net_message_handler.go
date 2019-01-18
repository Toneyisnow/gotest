package network

import "../common/log"

func HandleMessage(context *NetContext, message *NetMessage) {

	if context == nil {
		log.W("[network] handling message failed: context is nil.")
		return
	}

	if message == nil {
		log.W("[network] handling message failed: message is nil.")
		return
	}



	log.I2("[network] handling message. context=[%s] message=[%s]", context.index, message.MessageId)
	if message.MessageType == NetMessageType_Event {

		log.I("[network] EventMessage detected.")
		eventMessage := message.GetEventMessage()

		context.contextManager.netProcessor.GetEventHandler().HandleEventData(context, eventMessage.Data)
	}
}

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