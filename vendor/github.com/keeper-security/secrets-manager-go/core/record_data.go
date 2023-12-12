package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

type KeeperRecordData struct {
	Type   string              `json:"type,omitempty"`
	Title  string              `json:"title,omitempty"`
	Notes  string              `json:"notes,omitempty"`
	Fields []KeeperRecordField `json:"fields,omitempty"`
	Custom []KeeperRecordField `json:"custom,omitempty"`
}

type KeeperRecordField struct {
	Type  string `json:"type"`
	Label string `json:"label,omitempty"`
}

type Login struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// Login field constructor with the single value to eliminate the complexity of the passing List as a value
func NewLogin(value string) *Login {
	return &Login{
		KeeperRecordField: KeeperRecordField{Type: "login"},
		Value:             []string{value},
	}
}

type PasswordComplexity struct {
	Length    int `json:"length,omitempty"`
	Caps      int `json:"caps,omitempty"`
	Lowercase int `json:"lowercase,omitempty"`
	Digits    int `json:"digits,omitempty"`
	Special   int `json:"special,omitempty"`
}

type Password struct {
	KeeperRecordField
	Required          bool                `json:"required,omitempty"`
	PrivacyScreen     bool                `json:"privacyScreen,omitempty"`
	EnforceGeneration bool                `json:"enforceGeneration,omitempty"`
	Complexity        *PasswordComplexity `json:"complexity,omitempty"`
	Value             []string            `json:"value,omitempty"`
}

// Password field constructor with the single value to eliminate the complexity of the passing List as a value
func NewPassword(value string) *Password {
	return &Password{
		KeeperRecordField: KeeperRecordField{Type: "password"},
		Value:             []string{value},
	}
}

type Url struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// Url field constructor with the single value to eliminate the complexity of the passing List as a value
func NewUrl(value string) *Url {
	return &Url{
		KeeperRecordField: KeeperRecordField{Type: "url"},
		Value:             []string{value},
	}
}

type FileRef struct {
	KeeperRecordField
	Required bool     `json:"required,omitempty"`
	Value    []string `json:"value,omitempty"`
}

// FileRef field constructor with the single value to eliminate the complexity of the passing List as a value
func NewFileRef(value string) *FileRef {
	return &FileRef{
		KeeperRecordField: KeeperRecordField{Type: "fileRef"},
		Value:             []string{value},
	}
}

type OneTimeCode struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// OneTimeCode field constructor with the single value to eliminate the complexity of the passing List as a value
func NewOneTimeCode(value string) *OneTimeCode {
	return &OneTimeCode{
		KeeperRecordField: KeeperRecordField{Type: "oneTimeCode"},
		Value:             []string{value},
	}
}

type Name struct {
	First  string `json:"first,omitempty"`
	Middle string `json:"middle,omitempty"`
	Last   string `json:"last,omitempty"`
}

type Names struct {
	KeeperRecordField
	Required      bool   `json:"required,omitempty"`
	PrivacyScreen bool   `json:"privacyScreen,omitempty"`
	Value         []Name `json:"value,omitempty"`
}

// Names field constructor with the single value to eliminate the complexity of the passing List as a value
func NewNames(value Name) *Names {
	return &Names{
		KeeperRecordField: KeeperRecordField{Type: "name"},
		Value:             []Name{value},
	}
}

type BirthDate struct {
	KeeperRecordField
	Required      bool    `json:"required,omitempty"`
	PrivacyScreen bool    `json:"privacyScreen,omitempty"`
	Value         []int64 `json:"value,omitempty"`
}

// BirthDate field constructor with the single value to eliminate the complexity of the passing List as a value
func NewBirthDate(value int64) *BirthDate {
	return &BirthDate{
		KeeperRecordField: KeeperRecordField{Type: "birthDate"},
		Value:             []int64{value},
	}
}

type Date struct {
	KeeperRecordField
	Required      bool    `json:"required,omitempty"`
	PrivacyScreen bool    `json:"privacyScreen,omitempty"`
	Value         []int64 `json:"value,omitempty"`
}

// Date field constructor with the single value to eliminate the complexity of the passing List as a value
func NewDate(value int64) *Date {
	return &Date{
		KeeperRecordField: KeeperRecordField{Type: "date"},
		Value:             []int64{value},
	}
}

type ExpirationDate struct {
	KeeperRecordField
	Required      bool    `json:"required,omitempty"`
	PrivacyScreen bool    `json:"privacyScreen,omitempty"`
	Value         []int64 `json:"value,omitempty"`
}

// ExpirationDate field constructor with the single value to eliminate the complexity of the passing List as a value
func NewExpirationDate(value int64) *ExpirationDate {
	return &ExpirationDate{
		KeeperRecordField: KeeperRecordField{Type: "expirationDate"},
		Value:             []int64{value},
	}
}

type Text struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// Text field constructor with the single value to eliminate the complexity of the passing List as a value
func NewText(value string) *Text {
	return &Text{
		KeeperRecordField: KeeperRecordField{Type: "text"},
		Value:             []string{value},
	}
}

type SecurityQuestion struct {
	Question string `json:"question,omitempty"`
	Answer   string `json:"answer,omitempty"`
}

type SecurityQuestions struct {
	KeeperRecordField
	Required      bool               `json:"required,omitempty"`
	PrivacyScreen bool               `json:"privacyScreen,omitempty"`
	Value         []SecurityQuestion `json:"value,omitempty"`
}

// SecurityQuestions field constructor with the single value to eliminate the complexity of the passing List as a value
func NewSecurityQuestions(value SecurityQuestion) *SecurityQuestions {
	return &SecurityQuestions{
		KeeperRecordField: KeeperRecordField{Type: "securityQuestion"},
		Value:             []SecurityQuestion{value},
	}
}

type Multiline struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// Multiline field constructor with the single value to eliminate the complexity of the passing List as a value
func NewMultiline(value string) *Multiline {
	return &Multiline{
		KeeperRecordField: KeeperRecordField{Type: "multiline"},
		Value:             []string{value},
	}
}

type Email struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// Email field constructor with the single value to eliminate the complexity of the passing List as a value
func NewEmail(value string) *Email {
	return &Email{
		KeeperRecordField: KeeperRecordField{Type: "email"},
		Value:             []string{value},
	}
}

type CardRef struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// CardRef field constructor with the single value to eliminate the complexity of the passing List as a value
func NewCardRef(value string) *CardRef {
	return &CardRef{
		KeeperRecordField: KeeperRecordField{Type: "cardRef"},
		Value:             []string{value},
	}
}

type AddressRef struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// AddressRef field constructor with the single value to eliminate the complexity of the passing List as a value
func NewAddressRef(value string) *AddressRef {
	return &AddressRef{
		KeeperRecordField: KeeperRecordField{Type: "addressRef"},
		Value:             []string{value},
	}
}

type PinCode struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// PinCode field constructor with the single value to eliminate the complexity of the passing List as a value
func NewPinCode(value string) *PinCode {
	return &PinCode{
		KeeperRecordField: KeeperRecordField{Type: "pinCode"},
		Value:             []string{value},
	}
}

type Phone struct {
	Region string `json:"region,omitempty"` // Region code. Ex. US
	Number string `json:"number,omitempty"` // Phone number. Ex. 510-222-5555
	Ext    string `json:"ext,omitempty"`    // Extension number. Ex. 9987
	Type   string `json:"type,omitempty"`   // Phone number type. Ex. Mobile
}

type Phones struct {
	KeeperRecordField
	Required      bool    `json:"required,omitempty"`
	PrivacyScreen bool    `json:"privacyScreen,omitempty"`
	Value         []Phone `json:"value,omitempty"`
}

// Phones field constructor with the single value to eliminate the complexity of the passing List as a value
func NewPhones(value Phone) *Phones {
	return &Phones{
		KeeperRecordField: KeeperRecordField{Type: "phone"},
		Value:             []Phone{value},
	}
}

type Secret struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// Secret field constructor with the single value to eliminate the complexity of the passing List as a value
func NewSecret(value string) *Secret {
	return &Secret{
		KeeperRecordField: KeeperRecordField{Type: "secret"},
		Value:             []string{value},
	}
}

type SecureNote struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// SecureNote field constructor with the single value to eliminate the complexity of the passing List as a value
func NewSecureNote(value string) *SecureNote {
	return &SecureNote{
		KeeperRecordField: KeeperRecordField{Type: "note"},
		Value:             []string{value},
	}
}

type AccountNumber struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// AccountNumber field constructor with the single value to eliminate the complexity of the passing List as a value
func NewAccountNumber(value string) *AccountNumber {
	return &AccountNumber{
		KeeperRecordField: KeeperRecordField{Type: "accountNumber"},
		Value:             []string{value},
	}
}

type PaymentCard struct {
	CardNumber         string `json:"cardNumber,omitempty"`
	CardExpirationDate string `json:"cardExpirationDate,omitempty"`
	CardSecurityCode   string `json:"cardSecurityCode,omitempty"`
}

type PaymentCards struct {
	KeeperRecordField
	Required      bool          `json:"required,omitempty"`
	PrivacyScreen bool          `json:"privacyScreen,omitempty"`
	Value         []PaymentCard `json:"value,omitempty"`
}

// PaymentCards field constructor with the single value to eliminate the complexity of the passing List as a value
func NewPaymentCards(value PaymentCard) *PaymentCards {
	return &PaymentCards{
		KeeperRecordField: KeeperRecordField{Type: "paymentCard"},
		Value:             []PaymentCard{value},
	}
}

type BankAccount struct {
	AccountType   string `json:"accountType,omitempty"`
	RoutingNumber string `json:"routingNumber,omitempty"`
	AccountNumber string `json:"accountNumber,omitempty"`
	OtherType     string `json:"otherType,omitempty"`
}

type BankAccounts struct {
	KeeperRecordField
	Required      bool          `json:"required,omitempty"`
	PrivacyScreen bool          `json:"privacyScreen,omitempty"`
	Value         []BankAccount `json:"value,omitempty"`
}

// BankAccounts field constructor with the single value to eliminate the complexity of the passing List as a value
func NewBankAccounts(value BankAccount) *BankAccounts {
	return &BankAccounts{
		KeeperRecordField: KeeperRecordField{Type: "bankAccount"},
		Value:             []BankAccount{value},
	}
}

type KeyPair struct {
	PublicKey  string `json:"publicKey,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
}

type KeyPairs struct {
	KeeperRecordField
	Required      bool      `json:"required,omitempty"`
	PrivacyScreen bool      `json:"privacyScreen,omitempty"`
	Value         []KeyPair `json:"value,omitempty"`
}

// KeyPairs field constructor with the single value to eliminate the complexity of the passing List as a value
func NewKeyPairs(value KeyPair) *KeyPairs {
	return &KeyPairs{
		KeeperRecordField: KeeperRecordField{Type: "keyPair"},
		Value:             []KeyPair{value},
	}
}

type Host struct {
	Hostname string `json:"hostName,omitempty"`
	Port     string `json:"port,omitempty"`
}

type Hosts struct {
	KeeperRecordField
	Required      bool   `json:"required,omitempty"`
	PrivacyScreen bool   `json:"privacyScreen,omitempty"`
	Value         []Host `json:"value,omitempty"`
}

// Hosts field constructor with the single value to eliminate the complexity of the passing List as a value
func NewHosts(value Host) *Hosts {
	return &Hosts{
		KeeperRecordField: KeeperRecordField{Type: "host"},
		Value:             []Host{value},
	}
}

type Address struct {
	Street1 string `json:"street1,omitempty"`
	Street2 string `json:"street2,omitempty"`
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	Country string `json:"country,omitempty"`
	Zip     string `json:"zip,omitempty"`
}

type Addresses struct {
	KeeperRecordField
	Required      bool      `json:"required,omitempty"`
	PrivacyScreen bool      `json:"privacyScreen,omitempty"`
	Value         []Address `json:"value,omitempty"`
}

// Addresses field constructor with the single value to eliminate the complexity of the passing List as a value
func NewAddresses(value Address) *Addresses {
	return &Addresses{
		KeeperRecordField: KeeperRecordField{Type: "address"},
		Value:             []Address{value},
	}
}

type LicenseNumber struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []string `json:"value,omitempty"`
}

// LicenseNumber field constructor with the single value to eliminate the complexity of the passing List as a value
func NewLicenseNumber(value string) *LicenseNumber {
	return &LicenseNumber{
		KeeperRecordField: KeeperRecordField{Type: "licenseNumber"},
		Value:             []string{value},
	}
}

type RecordRef struct {
	KeeperRecordField
	Required bool     `json:"required,omitempty"`
	Value    []string `json:"value,omitempty"`
}

// RecordRef field constructor with the single value to eliminate the complexity of the passing List as a value
func NewRecordRef(value string) *RecordRef {
	return &RecordRef{
		KeeperRecordField: KeeperRecordField{Type: "recordRef"},
		Value:             []string{value},
	}
}

type Schedule struct {
	Type          string `json:"type,omitempty"`
	UtcTime       string `json:"utcTime,omitempty"`
	Weekday       string `json:"weekday,omitempty"`
	IntervalCount int    `json:"intervalCount,omitempty"`
}

type Schedules struct {
	KeeperRecordField
	Required bool       `json:"required,omitempty"`
	Value    []Schedule `json:"value,omitempty"`
}

// Schedules field constructor with the single value to eliminate the complexity of the passing List as a value
func NewSchedules(value Schedule) *Schedules {
	return &Schedules{
		KeeperRecordField: KeeperRecordField{Type: "schedule"},
		Value:             []Schedule{value},
	}
}

type DirectoryType struct {
	KeeperRecordField
	Required bool     `json:"required,omitempty"`
	Value    []string `json:"value,omitempty"`
}

// DirectoryType field constructor with the single value to eliminate the complexity of the passing List as a value
func NewDirectoryType(value string) *DirectoryType {
	return &DirectoryType{
		KeeperRecordField: KeeperRecordField{Type: "directoryType"},
		Value:             []string{value},
	}
}

type DatabaseType struct {
	KeeperRecordField
	Required bool     `json:"required,omitempty"`
	Value    []string `json:"value,omitempty"`
}

// DatabaseType field constructor with the single value to eliminate the complexity of the passing List as a value
func NewDatabaseType(value string) *DatabaseType {
	return &DatabaseType{
		KeeperRecordField: KeeperRecordField{Type: "databaseType"},
		Value:             []string{value},
	}
}

type PamHostname struct {
	KeeperRecordField
	Required      bool   `json:"required,omitempty"`
	PrivacyScreen bool   `json:"privacyScreen,omitempty"`
	Value         []Host `json:"value,omitempty"`
}

// PamHostname field constructor with the single value to eliminate the complexity of the passing List as a value
func NewPamHostname(value Host) *PamHostname {
	return &PamHostname{
		KeeperRecordField: KeeperRecordField{Type: "pamHostname"},
		Value:             []Host{value},
	}
}

type PamResource struct {
	ControllerUid string   `json:"controllerUid,omitempty"`
	FolderUid     string   `json:"folderUid,omitempty"`
	ResourceRef   []string `json:"resourceRef,omitempty"`
}

type PamResources struct {
	KeeperRecordField
	Required bool          `json:"required,omitempty"`
	Value    []PamResource `json:"value,omitempty"`
}

// PamResources field constructor with the single value to eliminate the complexity of the passing List as a value
func NewPamResources(value PamResource) *PamResources {
	return &PamResources{
		KeeperRecordField: KeeperRecordField{Type: "pamResources"},
		Value:             []PamResource{value},
	}
}

type Checkbox struct {
	KeeperRecordField
	Required bool   `json:"required,omitempty"`
	Value    []bool `json:"value,omitempty"`
}

// Checkbox field constructor with the single value to eliminate the complexity of the passing List as a value
func NewCheckbox(value bool) *Checkbox {
	return &Checkbox{
		KeeperRecordField: KeeperRecordField{Type: "checkbox"},
		Value:             []bool{value},
	}
}

type Script struct {
	FileRef   string   `json:"fileRef,omitempty"`
	Command   string   `json:"command,omitempty"`
	RecordRef []string `json:"recordRef,omitempty"`
}

type Scripts struct {
	KeeperRecordField
	Required      bool     `json:"required,omitempty"`
	PrivacyScreen bool     `json:"privacyScreen,omitempty"`
	Value         []Script `json:"value,omitempty"`
}

// Scripts field constructor with the single value to eliminate the complexity of the passing List as a value
func NewScripts(value Script) *Scripts {
	return &Scripts{
		KeeperRecordField: KeeperRecordField{Type: "script"},
		Value:             []Script{value},
	}
}

type PasskeyPrivateKey struct {
	Crv    string   `json:"crv,omitempty"`
	D      string   `json:"d,omitempty"`
	Ext    bool     `json:"ext,omitempty"`
	KeyOps []string `json:"key_ops,omitempty"`
	Kty    string   `json:"kty,omitempty"`
	X      string   `json:"x,omitempty"`
	Y      int64    `json:"y,omitempty"`
}

type Passkey struct {
	PrivateKey   PasskeyPrivateKey `json:"privateKey,omitempty"`
	CredentialId string            `json:"credentialId,omitempty"`
	SignCount    int64             `json:"signCount,omitempty"`
	UserId       string            `json:"userId,omitempty"`
	RelyingParty string            `json:"relyingParty,omitempty"`
	Username     string            `json:"username,omitempty"`
	CreatedDate  int64             `json:"createdDate,omitempty"`
}

type Passkeys struct {
	KeeperRecordField
	Required bool      `json:"required,omitempty"`
	Value    []Passkey `json:"value,omitempty"`
}

// Passkeys field constructor with the single value to eliminate the complexity of the passing List as a value
func NewPasskeys(value Passkey) *Passkeys {
	return &Passkeys{
		KeeperRecordField: KeeperRecordField{Type: "passkey"},
		Value:             []Passkey{value},
	}
}

// getKeeperRecordField converts fieldData from generic interface{} to strongly typed interface{}
func getKeeperRecordField(fieldType string, fieldData map[string]interface{}, validate bool) (field interface{}, err error) {
	if jsonField := DictToJson(fieldData); strings.TrimSpace(jsonField) != "" {
		switch fieldType {
		case "login":
			field = &Login{}
		case "password":
			field = &Password{}
		case "url":
			field = &Url{}
		case "fileRef":
			field = &FileRef{}
		case "oneTimeCode":
			field = &OneTimeCode{}
		case "name":
			field = &Names{}
		case "birthDate":
			field = &BirthDate{}
		case "date":
			field = &Date{}
		case "expirationDate":
			field = &ExpirationDate{}
		case "text":
			field = &Text{}
		case "securityQuestion":
			field = &SecurityQuestions{}
		case "multiline":
			field = &Multiline{}
		case "email":
			field = &Email{}
		case "cardRef":
			field = &CardRef{}
		case "addressRef":
			field = &AddressRef{}
		case "pinCode":
			field = &PinCode{}
		case "phone":
			field = &Phones{}
		case "secret":
			field = &Secret{}
		case "note":
			field = &SecureNote{}
		case "accountNumber":
			field = &AccountNumber{}
		case "paymentCard":
			field = &PaymentCards{}
		case "bankAccount":
			field = &BankAccounts{}
		case "keyPair":
			field = &KeyPairs{}
		case "host":
			field = &Hosts{}
		case "address":
			field = &Addresses{}
		case "licenseNumber":
			field = &LicenseNumber{}
		case "recordRef":
			field = &RecordRef{}
		case "schedule":
			field = &Schedules{}
		case "directoryType":
			field = &DirectoryType{}
		case "databaseType":
			field = &DatabaseType{}
		case "pamHostname":
			field = &PamHostname{}
		case "pamResources":
			field = &PamResources{}
		case "checkbox":
			field = &Checkbox{}
		case "script":
			field = &Scripts{}
		case "passkey":
			field = &Passkeys{}
		default:
			return nil, fmt.Errorf("unable to convert unknown field type %v", fieldType)
		}

		if validate {
			decoder := json.NewDecoder(strings.NewReader(jsonField))
			decoder.DisallowUnknownFields()
			err = decoder.Decode(field)
		} else {
			err = json.Unmarshal([]byte(jsonField), field)
		}
		return
	} else {
		return nil, fmt.Errorf("unable to parse field from JSON '%v'", fieldData)
	}
}

func IsFieldClass(field interface{}) bool {
	switch field.(type) {
	case
		AccountNumber, *AccountNumber,
		AddressRef, *AddressRef,
		Addresses, *Addresses,
		BankAccounts, *BankAccounts,
		BirthDate, *BirthDate,
		CardRef, *CardRef,
		Checkbox, *Checkbox,
		DatabaseType, *DatabaseType,
		Date, *Date,
		DirectoryType, *DirectoryType,
		Email, *Email,
		ExpirationDate, *ExpirationDate,
		FileRef, *FileRef,
		Hosts, *Hosts,
		KeyPairs, *KeyPairs,
		LicenseNumber, *LicenseNumber,
		Login, *Login,
		Multiline, *Multiline,
		Names, *Names,
		OneTimeCode, *OneTimeCode,
		Password, *Password,
		PamHostname, *PamHostname,
		PamResources, *PamResources,
		PaymentCards, *PaymentCards,
		Phones, *Phones,
		PinCode, *PinCode,
		RecordRef, *RecordRef,
		Schedules, *Schedules,
		Secret, *Secret,
		SecureNote, *SecureNote,
		SecurityQuestions, *SecurityQuestions,
		Text, *Text,
		Url, *Url,
		Scripts, *Scripts,
		Passkeys, *Passkeys:
		return true
	}
	return false
}

func structToMap(data interface{}) (map[string]interface{}, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	mapData := make(map[string]interface{})
	err = json.Unmarshal(dataBytes, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}
