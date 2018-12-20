package network

const (

	DEFAULT_HEARTBEAT_INTERVAL = 30 * 1000
	DEFAULT_HEATBEAT_TOLERATION_COUNT = 3
	DEFAULT_MAX_INCOMING_CONNECTION = 24
	DEFAULT_MAX_OUTCOMING_CONNECTION = 24
	DEFAULT_MAX_MESSAGE_DATA_SIZE_IN_BYTE = 0	// Not used yet
)

type NetOption struct {

	_heartbeatInterval int64
	_heartbeatTolerationCount int32
	_maxIncomingConnection int32
	_maxOutcomingConnection int32
	_maxMessageDataSizeInByte int64	// Not used yet

	_needChallenge bool
}

func DefaultOption() *NetOption {

	option := new(NetOption)
	option._heartbeatInterval = DEFAULT_HEARTBEAT_INTERVAL
	option._heartbeatTolerationCount = DEFAULT_HEATBEAT_TOLERATION_COUNT
	option._maxIncomingConnection = DEFAULT_MAX_INCOMING_CONNECTION
	option._maxOutcomingConnection = DEFAULT_MAX_OUTCOMING_CONNECTION
	option._needChallenge = true

	return option
}
