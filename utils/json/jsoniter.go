package json

import "github.com/json-iterator/go"

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	Marshal = json.Marshal
	MarshalToString = json.MarshalToString
	Unmarshal = json.Unmarshal
	MarshalIndent = json.MarshalIndent
	NewDecoder = json.NewDecoder
	NewEncoder = json.NewEncoder
)
