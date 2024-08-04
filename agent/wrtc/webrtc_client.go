package wrtc

import (
	"encoding/json"
	"fmt"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/agent/orchestratorclient"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
)

func getOrCreatePeerConnection(serial string) (*webrtc.PeerConnection, error) {
	log.Info("creating peer connection")
	webrtcconn, err := connectWebRTC(serial)
	if err != nil {
		return nil, fmt.Errorf("could not establish webrtc connection to device: %v", err)
	}
	return webrtcconn, nil

}

func connectWebRTC(serial string) (*webrtc.PeerConnection, error) {
	log.Info("webRTC connection starting")
	peerConnectionEstablished := make(chan interface{})
	peerConnectionError := make(chan error)
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}
	/*defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			fmt.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}()
	*/
	// Create a datachannel with label 'data'
	dataChannel, err := peerConnection.CreateDataChannel("health", nil)
	if err != nil {
		panic(err)
	}
	dataChannel.Label()
	go func() {
		for {
			err := dataChannel.SendText("ping")
			if err != nil {
				log.Errorf("error sending ping: %v", err)
				return
			}
		}
	}()

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())
		if s == webrtc.PeerConnectionStateConnected {
			peerConnectionEstablished <- struct{}{}
		}
		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			peerConnectionError <- fmt.Errorf("peer connection failed")
		}
	})

	// Create an offer to send to the other process
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	// Note: this will start the gathering of ICE candidates
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		return nil, err
	}
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	log.Debug("wait gather")
	<-gatherComplete
	log.Debug("complete")

	// Send our offer to the HTTP server listening in the other process
	payload, err := json.Marshal(peerConnection.LocalDescription())
	if err != nil {
		return nil, err
	}
	log.Info("sending offer using rest")
	resp, err := orchestratorclient.OfferSDP(models.SDP{
		ID:     uuid.New(),
		SDP:    string(payload),
		Serial: serial,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send offer: %w", err)
	}

	//log.Println(resp)
	var answer webrtc.SessionDescription
	err = json.Unmarshal([]byte(resp), &answer)
	if err != nil {
		return nil, fmt.Errorf("failed parsing json SDP answer: %w", err)
	}
	log.Infof("answer received %+v", answer)
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		return nil, fmt.Errorf("failed setting remote description: %w", err)
	}
	select {
	case <-peerConnectionEstablished:
		log.Info("peer connection established")
		return peerConnection, nil
	case err := <-peerConnectionError:
		return nil, fmt.Errorf("peer connection failed: %w", err)
	}
}
