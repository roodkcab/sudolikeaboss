package onepass

import (
	"encoding/json"
	"github.com/ravenac95/sudolikeaboss/websocketclient"
	"fmt"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"crypto/aes"
	"strings"
	"time"
	"strconv"
)

type Command struct {
	Action   string  `json:"action"`
	Number   int     `json:"number"`
	Version  string  `json:"version"`
	BundleId string  `json:"bundleId"`
	Payload  Payload `json:"payload"`
}

type Payload struct {
	Version      	string            	`json:"version,omitempty"`
	Capabilities 	[]string          	`json:"capabilities,omitempty"`

	ExtId  		string          	`json:"extId,omitempty"`
	Method 		string			`json:"method,omitempty"`
	Secret 		string 			`json:"secret,omitempty"`
	CC 		string 			`json:"cc,omitempty"`
	M4 		string 			`json:"M4,omitempty"`

	Url          	string            	`json:"url"`
	Title 		string 			`json:"title"`
	TabUrl		string 			`json:"tabUrl"`
	Options      	map[string]string 	`json:"options"`
	Context		string			`json:"context"`

	Alg		string 			`json:"alg"`
	Iv		string			`json:"iv"`
	Hmac		string			`json:"hmac"`
	Data		string                  `json:"data"`

	DocumentUUID		string				`json:"documentUUID"`
	DocumentURL		string 				`json:"documentUrl"`
	CollectedTimestamp	string 				`json:"collectedTimestamp"`
	Forms 			map[string]map[string]string 	`json:"forms"`
	Fields 			[]map[string]string		`json:"fields"`
}

type WebsocketClient interface {
	Connect() error
	Receive(v interface{}) error
	Send(v interface{}) error
}

type Configuration struct {
	WebsocketUri      string `json:"websocketUri"`
	WebsocketProtocol string `json:"websocketProtocol"`
	WebsocketOrigin   string `json:"websocketOrigin"`
	DefaultHost       string `json:"defaultHost"`
}

type OnePasswordClient struct {
	DefaultHost     string
	websocketClient WebsocketClient
	number          int
	encryptor 	encrypt
}

var method string
var extId = "7EA0C2BA4DAC1ACA796A4F26076E9936"
var codec = Codec{}

func NewClientWithConfig(configuration *Configuration) (*OnePasswordClient, error) {
	return NewClient(configuration.WebsocketUri, configuration.WebsocketProtocol, configuration.WebsocketOrigin, configuration.DefaultHost)
}

func NewClient(websocketUri string, websocketProtocol string, websocketOrigin string, defaultHost string) (*OnePasswordClient, error) {
	websocketClient := websocketclient.NewClient(websocketUri, websocketProtocol, websocketOrigin)

	return NewCustomClient(websocketClient, defaultHost)
}

func NewCustomClient(websocketClient WebsocketClient, defaultHost string) (*OnePasswordClient, error) {
	client := OnePasswordClient{
		websocketClient: websocketClient,
		DefaultHost:     defaultHost,
	}

	err := client.Connect()

	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (client *OnePasswordClient) Connect() error {
	err := client.websocketClient.Connect()

	return err
}

func (client *OnePasswordClient) createCommand(action string, payload Payload) *Command {
	command := Command{
		Action:   action,
		Number:   client.number,
		Version:  "4",
		BundleId: "com.sudolikeaboss.sudolikeaboss",
		Payload:  payload,
	}

	// Increment the number (it's a 1password thing that I saw whilst listening
	// to their commands
	client.number += 1
	return &command
}

func (client *OnePasswordClient) SendCommand(command *Command) (*Response, error) {
	jsonStr, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}
	err = client.websocketClient.Send(jsonStr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var rawResponseStr string
	err = client.websocketClient.Receive(&rawResponseStr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	response, err := LoadResponse(rawResponseStr)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (client *OnePasswordClient) SendHelloCommand() (*AuthResponse, error) {
	payload := Payload {
		Version:      "4.6.2.90",
		Capabilities: []string{"NextGenFillItem", "Synapse", "AddB5", "auth", "auth-sma-hmac256", "aead-cbchmac-256"},
		ExtId: extId,
	}
	command := client.createCommand("hello", payload)

	response, err := client.SendAuthCommand(command)
	if err != nil {
		return nil, err
	}
	method = response.Payload.Method
	if (response.Action == "authNew") {
		response, err = client.SendAuthRegisterCommand()
		if err != nil {
			return nil, err
		}
	}
	response, err = client.SendAuthBeginCommand()
	if err != nil {
		return nil, err
	}
	response, err = client.SendAuthContinueCommand(response)
	if err != nil {
		return nil, err
	}
	return response, nil
}


func (client *OnePasswordClient) SendAuthCommand(command *Command) (*AuthResponse, error) {
	jsonStr, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}
	err = client.websocketClient.Send(jsonStr)
	if err != nil {
		return nil, err
	}
	var rawResponseStr string
	err = client.websocketClient.Receive(&rawResponseStr)
	if err != nil {
		return nil, err
	}
	response, err := LoadAuthResponse(rawResponseStr)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (client *OnePasswordClient) extId() (string) {
	return "6EA0C2BA4DAC1ACA796A4F26076E9936"
}

func (client *OnePasswordClient) SendAuthRegisterCommand() (*AuthResponse, error) {
	/*
	L = a.alg;
        c = a.code;
        a = crypto.getRandomValues(new Uint8Array(16));
        a = sjcl.codec.bytes.toBits(a);
        a = sjcl.codec.base64.fromBits(a, true, true).toLowerCase();
	 */
	payload := Payload {
		Secret:		"123456",
		Method: 	method,
		ExtId:  	extId,
	}
	command := client.createCommand("authRegister", payload)
	response, err := client.SendAuthCommand(command)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (client *OnePasswordClient) SendAuthBeginCommand() (*AuthResponse, error) {
	//cc = Uint8Array[16]
	cc := codec.fromBits(codec.generateRandomBytesArray(16), true, true)
	payload := Payload {
		Method: 	method,
		ExtId:  	extId,
		CC: 		cc, //sjcl.codec.base64.fromBits(cc, !0, !0)
	}
	command := client.createCommand("authBegin", payload)
	response, err := client.SendAuthCommand(command)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (client *OnePasswordClient) SendAuthContinueCommand(authResponse *AuthResponse) (*AuthResponse, error) {
	/*
	        d = sjcl.codec.base64.toBits(response.cs, !0);
	        u = sha256([d, cc])
	        m3 == u
	        m4 = r.A(hmac(base64_to_bit(m3), secret))
	 */
	m3 := codec.toBits(authResponse.Payload.M3, true)
	key, err := hex.DecodeString("d76df8e7")
	if err != nil {
		fmt.Println(err)
	}
	sig := hmac.New(sha256.New, key)
	sig.Write(m3)
	m4 := sig.Sum(nil)

	payload := Payload {
		Method: 	method,
		ExtId:  	extId,
		M4:		codec.fromBits(m4, true, true),
	}
	command := client.createCommand("authVerify", payload)
	response, err := client.SendAuthCommand(command)
	if err != nil {
		return nil, err
	}

	encryptionBits := []byte("encryption")
	hmacBits := []byte("hmac")

	a := append(m3, m4...)
	a = append(a, encryptionBits...)
	sig.Reset()
	sig.Write(a)
	a = sig.Sum(nil)

	c := append(m4, m3...)
	c = append(c, hmacBits...)
	sig.Reset()
	sig.Write(c)
	c = sig.Sum(nil)

	cipher, err := aes.NewCipher(a)
	client.encryptor = encrypt {
		Ya: cipher,
		d: hmac.New(sha256.New, c),
	}

	return response, nil
}

func (client *OnePasswordClient) SendShowPopupCommand() (*ResponseData , error) {
	popUpPayload := Payload {
		Url:     client.DefaultHost,
		Options: map[string]string{"source": "toolbar-button"},
	}
	popUpPayloadStr, _ := json.Marshal(popUpPayload)
	command := client.createCommand("showPopup", client.encryptor.encryptPayload(string(popUpPayloadStr)))
	response, _ := client.SendCommand(command)

	if (response.Action == "collectDocuments") {
		responseData := client.encryptor.decryptPayload(response.Payload.Data, response.Payload.IV, response.Payload.Hmac)
		decryptPayload, _ := LoadResponseData(strings.Trim(responseData, "\f"))
		collectDocumentResultsPayload := Payload {
			CollectedTimestamp: strconv.FormatInt(time.Now().UTC().UnixNano() / 1000000 - 5000, 10),
			Context: decryptPayload.Context,
			DocumentUUID: "",
			DocumentURL: "",
			Url: "",
			Title: "",
			TabUrl: "",
			Forms: map[string]map[string]string{
			},
			Fields: []map[string]string {
				{"opid":"__0","elementNumber":"0","visible":"true","viewable":"true","htmlID":"email","htmlName":"","title":"","userEdited":"false","label-tag":"Email","label-right":"","label-left":"Email","placeholder":"","type":"email","value":"","disabled":"false","readonly":"false","onepasswordFieldType":"email","onepasswordDesignation":"username"},
				{"opid":"__1","elementNumber":"2","visible":"true","viewable":"true","htmlID":"master-password","htmlName":"","title":"","userEdited":"false","label-tag":"Master Password","label-right":"","label-left":"Master Password","placeholder":"","type":"password","value":"","disabled":"false","readonly":"false","onepasswordFieldType":"password","onepasswordDesignation":"password"},
			},
		}
		collectDocumentResults, _ := json.Marshal(collectDocumentResultsPayload)
		command = client.createCommand("collectDocumentResults", client.encryptor.encryptPayload(string(collectDocumentResults)))
		response, err := client.SendCommand(command)
		if (err != nil) {
			fmt.Println(err)
			return nil, err
		}
		responseData = client.encryptor.decryptPayload(response.Payload.Data, response.Payload.IV, response.Payload.Hmac)
		decryptPayload, err = LoadResponseData(strings.Trim(responseData, "\a"))
		if (err != nil) {
			fmt.Println(err)
			return nil, err
		}
		return decryptPayload, nil
	}
	return nil, nil
}

/*func (client *OnePasswordClient) SendFillItemCommand(data *ResponseData) (*Response, error) {
	payload := Payload {
		Context: data.Context,
		Details: map[string]string{"numberOfFieldsFilled": "0", "numberOfPasswordsFilled": "0"},
	}
	command := client.createCommand("fillItemResults", payload)
	response, err := client.SendCommand(command)
	if err != nil {
		return nil, err
	}
	return response, nil
}*/
