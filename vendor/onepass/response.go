package onepass

import (
	"encoding/json"
	"errors"
)

type AuthResponse struct {
	Action  string          	`json:"action"`
	Version string          	`json:"version"`
	Payload AuthResponsePayload 	`json:"payload"`
}

type AuthResponsePayload struct {
	Alg 	string 	`json:"alg"`
	Code 	string 	`json:"code"`
	Method 	string 	`json:"method"`
	M3 	string 	`json:"m3"`
}

type Response struct {
	Action  string          `json:"action"`
	Version string          `json:"version"`
	Payload ResponsePayload `json:"payload"`
}

type ResponsePayload struct {
	Item          ItemResponsePayload    `json:"item"`
	Options       map[string]interface{} `json:"options"`
	OpenInTabMode string                 `json:"openInTabMode"`
	IV 	string 	`json:"iv"`
	Hmac 	string 	`json:"hmac"`
	Data 	string 	`json:"data"`
	Alg 	string 	`json:"alg"`
}

type ResponseData struct {
	NakedDomains	[]string 		`json:"nakedDomains"`
	OpenInTabMode 	string			`json:"openInTabMode"`
	Url 		string 			`json:"url"`
	ItemUUID 	string 			`json:"itemUUID"`
	Context 	string 			`json:"context"`
	Script 		[][]string		`json:"script"`
}

type ResponseContext struct {
	ItemUUID 	string 		`json:"itemUUID"`
	ProfileUUID 	string 		`json:"profileUUID"`
	UUID 		string 		`json:"uuid"`
}

type ItemResponsePayload struct {
	Uuid           string                 `json:"uuid"`
	NakedDomains   []string               `json:"nakedDomains"`
	Overview       map[string]interface{} `json:"overview"`
	SecureContents SecureContents         `json:"secureContents"`
}

type SecureContents struct {
	HtmlForm map[string]interface{} `json:"htmlForm"`
	Fields   []map[string]string    `json:"fields"`
}

func LoadAuthResponse(rawResponseStr string) (*AuthResponse, error) {
	rawResponseBytes := []byte(rawResponseStr)
	var response AuthResponse

	if err := json.Unmarshal(rawResponseBytes, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func LoadResponseData(rawResponseStr string) (*ResponseData, error) {
	rawResponseBytes := []byte(rawResponseStr)
	var response ResponseData

	idx := len(rawResponseBytes) - 1
	for idx >= 0 {
		if (rawResponseBytes[idx] == 125) {
			break
		}
		rawResponseBytes = rawResponseBytes[:len(rawResponseBytes) - 1]
		idx--
	}
	if err := json.Unmarshal(rawResponseBytes, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func LoadResponse(rawResponseStr string) (*Response, error) {
	rawResponseBytes := []byte(rawResponseStr)
	var response Response

	if err := json.Unmarshal(rawResponseBytes, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func LoadContext(context string) (*ResponseContext, error) {
	rawContext := []byte(context)
	var response ResponseContext

	if err := json.Unmarshal(rawContext, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (response *ResponseData) GetPassword() (string, error) {
	return response.Script[1][2], nil
	return "", errors.New("No password found in the response")
}
