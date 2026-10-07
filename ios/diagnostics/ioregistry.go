package diagnostics

import ios "github.com/danielpaulus/go-ios/ios"

type ioregistryRequest struct {
	reqMap map[string]string
}

func newIORegistryRequest() *ioregistryRequest {
	return &ioregistryRequest{map[string]string{
		"Request": "IORegistry",
	}}
}

func (req *ioregistryRequest) addPlane(plane string) {
	req.reqMap["CurrentPlane"] = plane
}

func (req *ioregistryRequest) addName(name string) {
	req.reqMap["EntryName"] = name
}

func (req *ioregistryRequest) addClass(class string) {
	req.reqMap["EntryClass"] = class
}

func (req *ioregistryRequest) encoded() ([]byte, error) {
	bt, err := ios.PlistCodec{}.Encode(req.reqMap)
	if err != nil {
		return nil, err
	}
	return bt, nil
}

// IORegistry queries the device's ioregistry through the diagnostics relay.
// plane, name and class map to the CurrentPlane, EntryName and EntryClass keys
// of the request; empty values are omitted, so any subset can be passed.
// Note that newer iOS versions return no data for EntryName queries; prefer
// EntryClass there (e.g. class "IOPMPowerSource" for battery stats).
func (diagnosticsConn *Connection) IORegistry(plane string, name string, class string) (map[string]interface{}, error) {
	req := newIORegistryRequest()
	if plane != "" {
		req.addPlane(plane)
	}
	if name != "" {
		req.addName(name)
	}
	if class != "" {
		req.addClass(class)
	}
	encoded, err := req.encoded()
	if err != nil {
		return nil, err
	}
	err = diagnosticsConn.deviceConn.Send(encoded)
	if err != nil {
		return nil, err
	}
	response, err := diagnosticsConn.plistCodec.Decode(diagnosticsConn.deviceConn.Reader())
	if err != nil {
		return nil, err
	}
	return ios.ParsePlist(response)
}
