package network

import (
	"../common/log"
	"github.com/decred/dcrd/dcrec/secp256k1"
	"math/rand"
	"strconv"
)

func HandleMessage(context *NetContext, message *NetMessage) {

	if context == nil {
		log.W("[network] handling message failed: context is nil.")
		return
	}

	if message == nil {
		log.W("[network] handling message failed: message is nil.")
		return
	}

	log.I2("[network] handling message. context=[%s] message type=[%d] message id=[%s]", context.index, message.MessageType, message.MessageId)

	if message.MessageType == NetMessageType_Challenge {

		log.I("[network] challenge message detected.")
		challengeMessage := message.GetChallengeMessage()
		HandleChallengeMessage(context, challengeMessage)

	} else if message.MessageType == NetMessageType_ChallengeResponse {

		log.I("[network] challenge response message detected.")
		challengeResponseMessage := message.GetChallengeResponseMessage()
		HandleChallengeResponseMessage(context,challengeResponseMessage)

	} else if message.MessageType == NetMessageType_Event {

		log.I("[network] event message detected.")
		eventMessage := message.GetEventMessage()

		context.contextManager.netProcessor.GetEventHandler().HandleEventData(context, eventMessage.Data)
	}
}

func ComposeChallengeMessage(challenge []byte) *NetMessage {

	message := new(NetMessage)
	message.MessageId = GenerateRandomMessageId()
	message.MessageType = NetMessageType_Challenge

	challengeMessage := new(ChallengeMessage)
	challengeMessage.Challenge = challenge

	message.Data = &NetMessage_ChallengeMessage{ChallengeMessage: challengeMessage}

	return message
}

func HandleChallengeMessage(context *NetContext, message *ChallengeMessage) {

	if context == nil || message == nil {
		return
	}

	if context.contextManager == nil || context.contextManager.netProcessor == nil ||
		context.contextManager.netProcessor.topology == nil {
		return
	}

	selfDevice := context.contextManager.netProcessor.topology.Self()

	if selfDevice == nil || selfDevice.PrivateKey == nil {
		return
	}

	privateKey, _ := secp256k1.PrivKeyFromBytes(selfDevice.PrivateKey)
	plainTextBytes, err := secp256k1.Decrypt(privateKey, message.Challenge)

	log.D("[network] handling challenge message. challenge len=", len(message.Challenge))
	log.D("[network] decrypted plain text:", string(plainTextBytes))

	if err != nil {
		return
	}

	responseMessage := ComposeChallengeResponseMessage(message.Challenge, string(plainTextBytes))
	err = context.SendMessage(responseMessage)

	if err == nil {
		context.status = NetConextStatus_Ready
	} else {
		context.status = NetConextStatus_Closed
	}
}

func ComposeChallengeResponseMessage(challenge []byte, plainText string) *NetMessage {

	message := new(NetMessage)
	message.MessageId = GenerateRandomMessageId()
	message.MessageType = NetMessageType_ChallengeResponse

	challengeResponseMessage := new(ChallengeResponseMessage)
	challengeResponseMessage.Challenge = challenge
	challengeResponseMessage.PlainText = plainText

	message.Data = &NetMessage_ChallengeResponseMessage{ChallengeResponseMessage: challengeResponseMessage}

	return message
}

func HandleChallengeResponseMessage(context *NetContext, message *ChallengeResponseMessage) {

	if context == nil || message == nil {
		return
	}

	plainText := context.GetMetadata("challenge_plain_text")
	log.I("[net] handling challenge response.")

	if plainText == message.PlainText {
		log.I("[network] challenge response text is matching. set status to ready")
		context.status = NetConextStatus_Ready
	} else {
		log.I("[network] challenge response text is not matching. set status to closed. got text=", message.PlainText, "text in metadata=", plainText)
		context.status = NetConextStatus_Closed
	}
}

func ComposeEventMessage(eventData []byte) *NetMessage {

	message := new(NetMessage)
	message.MessageId = GenerateRandomMessageId()
	message.MessageType = NetMessageType_Event

	eventMessage := new(EventMessage)
	eventMessage.Data = eventData


	message.Data = &NetMessage_EventMessage{EventMessage: eventMessage}

	return message

}

func GenerateRandomMessageId() string {

	return strconv.FormatUint(rand.Uint64(), 10)
}