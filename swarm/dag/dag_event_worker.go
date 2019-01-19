package dag

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"github.com/smartswarm/core/crypto/secp256k1"
	"github.com/smartswarm/go/log"
	"github.com/gogo/protobuf/proto"
)

type DagEventWorker struct {

	dagEngine *DagEngine

	isBusy     bool
	eventQueue chan *DagEvent

	stopSignal chan int
}

func NewDagEventWorker(queue chan *DagEvent, engine *DagEngine) *DagEventWorker {

	worker := new(DagEventWorker)
	worker.isBusy = false
	worker.stopSignal = make(chan int)
	worker.eventQueue = queue
	worker.dagEngine = engine

	return worker
}

func (this *DagEventWorker) DoEvent(event *DagEvent) {

	// Do the actual work here

	if event.GetEventType() == DagEventType_VertexesData {

		this.handleVertexesDataEvent(event.GetVertexesDataEvent())
	}
}

func (this *DagEventWorker) IsBusy() bool {
	return this.isBusy
}

func (this *DagEventWorker) Stop() {

	this.stopSignal <- 1
}

func (this *DagEventWorker) Start() {

	for {
		select {
			case event := <- this.eventQueue:
				this.isBusy = true
				this.DoEvent(event)
				this.isBusy = false
			case stop := <-this.stopSignal:
				if stop == 1 {
						return
				}
		}
	}
}

func (this *DagEventWorker) handleVertexesDataEvent(vertexesDataEvent *VertexesDataEvent) {

	if vertexesDataEvent == nil || vertexesDataEvent.MainVertex == nil {
		log.W("[dag] handleVertexesDataEvent error: vertexesDataEvent is not complete")
		return
	}

	// Deal with the related vertex first, and then finally deal with main Vertex
	if vertexesDataEvent.Vertexes != nil {
		for _, vertex := range vertexesDataEvent.Vertexes {

			if validated, err := this.validateVertex(vertex); validated == ProcessResult_Yes && err == nil {

				// If the checking is passed, push the vertex into next queue
				this.dagEngine.dagStorage.chanIncomingVertex.PushProto(vertex)
			}
		}
	}

	mainVertex := vertexesDataEvent.MainVertex
	if validated, err := this.validateVertex(mainVertex); validated == ProcessResult_Yes && err == nil {

		// If the checking is passed, push the vertex into next queue, and also decide whether to send it out
		this.dagEngine.dagStorage.chanIncomingVertex.PushProto(mainVertex)
		this.dagEngine.PushIncomingMainVertex(mainVertex)
	}
}

func (this *DagEventWorker) validateVertex(vertex *DagVertex) (result ProcessResult, err error) {

	if vertex == nil || vertex.Hash == nil || vertex.Signature == nil ||vertex.CreatorNodeId == 0 {
		return ProcessResult_No, nil
	}

	// Confirm the hash and signature is correct
	contentBytes, err := proto.Marshal(vertex.Content)
	if err != nil {
		log.W("[dag] fatal: proto Marshal content failed.")
		return ProcessResult_No, nil
	}

	calculatedHash := sha256.Sum256(contentBytes)
	if !bytes.Equal(vertex.Hash, calculatedHash[:]) {
		log.W("[dag] handleVertex: hash value is not correct")
		return ProcessResult_No, nil
	}

	peerNode := this.dagEngine.dagNodes.GetPeerNodeById(vertex.CreatorNodeId)
	if peerNode == nil {
		log.W("[dag] handleVertex: cannot find creator node")
		return ProcessResult_No, nil
	}

	if !secp256k1.VerifySignature(peerNode.Device.PublicKey, vertex.Hash, vertex.Signature[:64]) {

		publicKeyString := hex.EncodeToString(peerNode.Device.PublicKey)
		signatureString := hex.EncodeToString((vertex.Signature))
		log.W("[dag] handleVertex: signature is not matching. expected sig:[", signatureString, "] device public key:[", publicKeyString, "]")
		return ProcessResult_No, nil
	}

	return ProcessResult_Yes, nil
}