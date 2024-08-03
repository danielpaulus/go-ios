package wrtc

import (
	"encoding/json"
	"fmt"

	"sync"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/syslog"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
)

func CreateSDPAnswer(sdps []models.SDP) ([]models.SDP, error) {
	var sdpAnswers []models.SDP
	for _, sdp := range sdps {
		sdpAnswer, err := generateSDPAnswer(sdp)
		if err != nil {
			log.WithField("sdpid", sdp.ID).WithField("err", err).Errorf("failed creating answer for sdp record")
			continue
		}
		sdpAnswers = append(sdpAnswers, sdpAnswer)
	}
	return sdpAnswers, nil
}

func generateSDPAnswer(sdpModel models.SDP) (models.SDP, error) {

	var candidatesMux sync.Mutex
	pendingCandidates := make([]*webrtc.ICECandidate, 0)
	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

	// Prepare the configuration
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
		if err := peerConnection.Close(); err != nil {
			fmt.Printf("cannot close peerConnection: %v\n", err)
		}
	}()
	*/

	// When an ICE candidate is available send to the other Pion instance
	// the other Pion instance will add this candidate by calling AddICECandidate
	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		candidatesMux.Lock()
		defer candidatesMux.Unlock()

		desc := peerConnection.RemoteDescription()
		if desc == nil {
			pendingCandidates = append(pendingCandidates, c)
		}
	})

	sdp := webrtc.SessionDescription{}
	err = json.Unmarshal([]byte(sdpModel.SDP), &sdp)
	if err != nil {
		panic(err)
	}

	if err := peerConnection.SetRemoteDescription(sdp); err != nil {
		panic(err)
	}

	// Create an answer to send to the other process
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	log.Info("wait gather")
	<-gatherComplete
	log.Info("complete")
	fmt.Println((*peerConnection.LocalDescription()))

	// Send our answer to the HTTP server listening in the other process
	payload, err := json.Marshal(peerConnection.LocalDescription())
	if err != nil {
		panic(err)
	}
	responseSdpModel := models.SDP{
		ID:  sdpModel.ID,
		SDP: string(payload),
	}
	/*	candidatesMux.Lock()
		for _, c := range pendingCandidates {
			onICECandidateErr := signalCandidate(c)
			if onICECandidateErr != nil {
				panic(onICECandidateErr)
			}
		}
		candidatesMux.Unlock()
	*/
	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection state has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
			//os.Exit(0)
		}
	})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s\n", d.Label())

		// Register channel opening handling
		d.OnOpen(func() {
			fmt.Printf("Data channel '%s' opened.", d.Label())
			err := d.SendText("ok")
			if err != nil {
				log.Errorf("failed sending datachannel message: %v", err)
			}
		})
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Info("received message")
			var data map[string]string
			json.Unmarshal(msg.Data, &data)
			log.Infof("received data: %+v", data)
			if data["cmd"] == "syslog" {
				log.Info("launching syslog")
				udid := data["serial"]
				de := ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: udid}}
				go func() {
					log.Info("start pushing logs to remote")
					syslogConnection, err := syslog.New(de)
					if err != nil {
						log.Errorf("failed creating syslog connection: %v", err)
						return
					}
					for {
						s, err := syslogConnection.ReadLogMessage()
						if err != nil {
							log.Errorf("error reading syslog: %v", err)
							return
						}
						d.Send([]byte(s))
					}
				}()
			}
		})

	})

	return responseSdpModel, nil
}
