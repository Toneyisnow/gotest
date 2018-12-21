package dag

type PayloadHandler interface {

	OnPayloadSubmitted(data PayloadData)

	OnPayloadAccepted(data PayloadData)

	OnPayloadRejected(data PayloadData)
}
