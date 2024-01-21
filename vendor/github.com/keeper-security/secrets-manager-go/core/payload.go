package core

import (
	"encoding/json"
	"net/http"

	klog "github.com/keeper-security/secrets-manager-go/core/logger"
)

type Context struct {
	TransmissionKey TransmissionKey
	ClientId        []byte
	ClientKey       []byte
}

func NewContext(transmissionKey TransmissionKey, clientId []byte, clientKey []byte) *Context {
	return &Context{
		TransmissionKey: transmissionKey,
		ClientId:        clientId,
		ClientKey:       clientKey,
	}
}

type TransmissionKey struct {
	PublicKeyId  string
	Key          []byte
	EncryptedKey []byte
}

func NewTransmissionKey(publicKeyId string, key []byte, encryptedKey []byte) *TransmissionKey {
	return &TransmissionKey{
		PublicKeyId:  publicKeyId,
		Key:          key,
		EncryptedKey: encryptedKey,
	}
}

type GetPayload struct {
	ClientVersion    string   `json:"clientVersion"`
	ClientId         string   `json:"clientId"`
	PublicKey        string   `json:"publicKey,omitempty"`
	RequestedRecords []string `json:"requestedRecords"`
	RequestedFolders []string `json:"requestedFolders"`
}

func (p *GetPayload) GetPayloadToJson() (string, error) {
	if pb, err := json.Marshal(p); err == nil {
		return string(pb), nil
	} else {
		klog.Error("Error serializing GetPayload to JSON: " + err.Error())
		return "", err
	}
}

func (p *GetPayload) GetPayloadFromJson(jsonData string) {
	bytes := []byte(jsonData)
	res := GetPayload{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		*p = res
	} else {
		klog.Error("Error deserializing GetPayload from JSON: " + err.Error())
	}
}

type UpdateTransactionType string

const (
	TransactionTypeNone     UpdateTransactionType = ""
	TransactionTypeGeneral  UpdateTransactionType = "general"
	TransactionTypeRotation UpdateTransactionType = "rotation"
)

type UpdatePayload struct {
	ClientVersion   string                `json:"clientVersion"`
	ClientId        string                `json:"clientId"`
	RecordUid       string                `json:"recordUid"`
	Revision        int64                 `json:"revision"`
	Data            string                `json:"data"`
	TransactionType UpdateTransactionType `json:"transactionType,omitempty"`
}

func (p *UpdatePayload) UpdatePayloadToJson() (string, error) {
	if pb, err := json.Marshal(p); err == nil {
		return string(pb), nil
	} else {
		klog.Error("Error serializing UpdatePayload to JSON: " + err.Error())
		return "", err
	}
}

func (p *UpdatePayload) UpdatePayloadFromJson(jsonData string) {
	bytes := []byte(jsonData)
	res := UpdatePayload{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		*p = res
	} else {
		klog.Error("Error deserializing UpdatePayload from JSON: " + err.Error())
	}
}

type CompleteTransactionPayload struct {
	ClientVersion string `json:"clientVersion"`
	ClientId      string `json:"clientId"`
	RecordUid     string `json:"recordUid"`
}

func (p *CompleteTransactionPayload) CompleteTransactionPayloadToJson() (string, error) {
	if pb, err := json.Marshal(p); err == nil {
		return string(pb), nil
	} else {
		klog.Error("Error serializing CompleteTransactionPayload to JSON: " + err.Error())
		return "", err
	}
}

func (p *CompleteTransactionPayload) CompleteTransactionPayloadFromJson(jsonData string) {
	bytes := []byte(jsonData)
	res := CompleteTransactionPayload{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		*p = res
	} else {
		klog.Error("Error deserializing CompleteTransactionPayload from JSON: " + err.Error())
	}
}

type CreatePayload struct {
	ClientVersion string `json:"clientVersion"`
	ClientId      string `json:"clientId"`
	RecordUid     string `json:"recordUid"`
	RecordKey     string `json:"recordKey"`
	FolderUid     string `json:"folderUid"`
	FolderKey     string `json:"folderKey"`
	Data          string `json:"data"`
	SubFolderUid  string `json:"subFolderUid,omitempty"`
}

func (p *CreatePayload) CreatePayloadToJson() (string, error) {
	if pb, err := json.Marshal(p); err == nil {
		return string(pb), nil
	} else {
		klog.Error("Error serializing CreatePayload to JSON: " + err.Error())
		return "", err
	}
}

func (p *CreatePayload) CreatePayloadFromJson(jsonData string) {
	bytes := []byte(jsonData)
	res := CreatePayload{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		*p = res
	} else {
		klog.Error("Error deserializing CreatePayload from JSON: " + err.Error())
	}
}

type DeletePayload struct {
	ClientVersion string   `json:"clientVersion"`
	ClientId      string   `json:"clientId"`
	RecordUids    []string `json:"recordUids"`
}

func (p *DeletePayload) DeletePayloadToJson() (string, error) {
	if pb, err := json.Marshal(p); err == nil {
		return string(pb), nil
	} else {
		klog.Error("Error serializing DeletePayload to JSON: " + err.Error())
		return "", err
	}
}

func (p *DeletePayload) DeletePayloadFromJson(jsonData string) {
	bytes := []byte(jsonData)
	res := DeletePayload{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		*p = res
	} else {
		klog.Error("Error deserializing DeletePayload from JSON: " + err.Error())
	}
}

type CreateFolderPayload struct {
	ClientVersion   string `json:"clientVersion"`
	ClientId        string `json:"clientId"`
	FolderUid       string `json:"folderUid"`
	SharedFolderUid string `json:"sharedFolderUid"`
	SharedFolderKey string `json:"sharedFolderKey"`
	Data            string `json:"data"`
	ParentUid       string `json:"parentUid"`
}

func (p *CreateFolderPayload) CreateFolderPayloadToJson() (string, error) {
	if pb, err := json.Marshal(p); err == nil {
		return string(pb), nil
	} else {
		klog.Error("Error serializing CreateFolderPayload to JSON: " + err.Error())
		return "", err
	}
}

func (p *CreateFolderPayload) CreateFolderPayloadFromJson(jsonData string) {
	bytes := []byte(jsonData)
	res := CreateFolderPayload{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		*p = res
	} else {
		klog.Error("Error deserializing CreateFolderPayload from JSON: " + err.Error())
	}
}

type UpdateFolderPayload struct {
	ClientVersion string `json:"clientVersion"`
	ClientId      string `json:"clientId"`
	FolderUid     string `json:"folderUid"`
	Data          string `json:"data"`
}

func (p *UpdateFolderPayload) UpdateFolderPayloadToJson() (string, error) {
	if pb, err := json.Marshal(p); err == nil {
		return string(pb), nil
	} else {
		klog.Error("Error serializing UpdateFolderPayload to JSON: " + err.Error())
		return "", err
	}
}

func (p *UpdateFolderPayload) UpdateFolderPayloadFromJson(jsonData string) {
	bytes := []byte(jsonData)
	res := UpdateFolderPayload{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		*p = res
	} else {
		klog.Error("Error deserializing UpdateFolderPayload from JSON: " + err.Error())
	}
}

type DeleteFolderPayload struct {
	ClientVersion string   `json:"clientVersion"`
	ClientId      string   `json:"clientId"`
	FolderUids    []string `json:"folderUids"`
	ForceDeletion bool     `json:"forceDeletion"`
}

func (p *DeleteFolderPayload) DeleteFolderPayloadToJson() (string, error) {
	if pb, err := json.Marshal(p); err == nil {
		return string(pb), nil
	} else {
		klog.Error("Error serializing DeleteFolderPayload to JSON: " + err.Error())
		return "", err
	}
}

func (p *DeleteFolderPayload) DeleteFolderPayloadFromJson(jsonData string) {
	bytes := []byte(jsonData)
	res := DeleteFolderPayload{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		*p = res
	} else {
		klog.Error("Error deserializing DeleteFolderPayload from JSON: " + err.Error())
	}
}

type FileUploadPayload struct {
	ClientVersion   string `json:"clientVersion"`
	ClientId        string `json:"clientId"`
	FileRecordUid   string `json:"fileRecordUid"`
	FileRecordKey   string `json:"fileRecordKey"`
	FileRecordData  string `json:"fileRecordData"`
	OwnerRecordUid  string `json:"ownerRecordUid"`
	OwnerRecordData string `json:"ownerRecordData"`
	LinkKey         string `json:"linkKey"`
	FileSize        int    `json:"fileSize"`
}

func (p *FileUploadPayload) FileUploadPayloadToJson() (string, error) {
	if pb, err := json.Marshal(p); err == nil {
		return string(pb), nil
	} else {
		klog.Error("Error serializing FileUploadPayload to JSON: " + err.Error())
		return "", err
	}
}

func FileUploadPayloadFromJson(jsonData string) *FileUploadPayload {
	bytes := []byte(jsonData)
	res := FileUploadPayload{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		return &res
	} else {
		klog.Error("Error deserializing FileUploadPayload from JSON: " + err.Error())
		return nil
	}
}

type EncryptedPayload struct {
	EncryptedPayload []byte
	Signature        []byte
}

func NewEncryptedPayload(encryptedPayload []byte, signature []byte) *EncryptedPayload {
	return &EncryptedPayload{
		EncryptedPayload: encryptedPayload,
		Signature:        signature,
	}
}

type KsmHttpResponse struct {
	StatusCode   int
	Data         []byte
	HttpResponse *http.Response
}

func NewKsmHttpResponse(statusCode int, data []byte, httpResponse *http.Response) *KsmHttpResponse {
	return &KsmHttpResponse{
		StatusCode:   statusCode,
		Data:         data,
		HttpResponse: httpResponse,
	}
}

type QueryOptions struct {
	RecordsFilter []string
	FoldersFilter []string
}

type CreateOptions struct {
	FolderUid    string
	SubFolderUid string
}
