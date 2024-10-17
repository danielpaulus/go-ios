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
