package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	klog "github.com/keeper-security/secrets-manager-go/core/logger"
)

type FieldSectionFlag byte

const (
	FieldSectionFields FieldSectionFlag = 1 << iota
	FieldSectionCustom
	FieldSectionBoth = FieldSectionFields | FieldSectionCustom
)

type FieldTokenFlag byte

const (
	FieldTokenType FieldTokenFlag = 1 << iota
	FieldTokenLabel
	FieldTokenBoth = FieldTokenType | FieldTokenLabel
)

type Record struct {
	RecordKeyBytes []byte
	Uid            string
	folderKeyBytes []byte
	folderUid      string
	innerFolderUid string
	Files          []*KeeperFile
	Revision       int64
	IsEditable     bool
	recordType     string
	RawJson        string
	RecordDict     map[string]interface{}
}

func (r *Record) FolderUid() string {
	return r.folderUid
}

func (r *Record) InnerFolderUid() string {
	return r.innerFolderUid
}

func (r *Record) Password() string {
	password := ""
	// password (if `login` type)
	if r.Type() == "login" {
		password = r.GetFieldValueByType("password")
	}
	return password
}

func (r *Record) SetPassword(password string) {
	if passwordFields := r.GetFieldsByType("password"); len(passwordFields) > 0 {
		passwordField := passwordFields[0]
		if vlist, ok := passwordField["value"].([]interface{}); ok && len(vlist) > 0 {
			if _, ok := vlist[0].(string); ok {
				vlist[0] = password
			} else {
				klog.Error("error changing password - expected string value")
			}
		} else {
			passwordField["value"] = []string{password}
		}
		r.update()
	} else {
		klog.Error("password field not found for UID: " + r.Uid)
	}
}

func (r *Record) SetFieldValueSingle(fieldType, value string) {
	if fields := r.GetFieldsByType(fieldType); len(fields) > 0 {
		field := fields[0]
		if vlist, ok := field["value"].([]interface{}); ok && len(vlist) > 0 {
			if _, ok := vlist[0].(string); ok {
				vlist[0] = value
			} else {
				klog.Error("error changing field value - expected string value")
			}
		} else {
			field["value"] = []string{value}
		}
		r.update()
	} else {
		klog.Error("field not found for UID: " + r.Uid)
	}
}

func (r *Record) SetCustomFieldValueSingle(fieldLabel, value string) {
	if fields := r.GetCustomFieldsByLabel(fieldLabel); len(fields) > 0 {
		field := fields[0]
		if vlist, ok := field["value"].([]interface{}); ok && len(vlist) > 0 {
			if _, ok := vlist[0].(string); ok {
				vlist[0] = value
			} else {
				klog.Error("error changing custom field value - expected string value")
			}
		} else {
			field["value"] = []string{value}
		}
		r.update()
	} else {
		klog.Error("custom field not found for UID: " + r.Uid)
	}
}

func (r *Record) Title() string {
	if recordTitle, ok := r.RecordDict["title"].(string); ok {
		return recordTitle
	}
	return ""
}

func (r *Record) SetTitle(title string) {
	if _, ok := r.RecordDict["title"]; ok {
		r.RecordDict["title"] = title
		r.update()
	}
}

func (r *Record) Type() string {
	if recordType, ok := r.RecordDict["type"].(string); ok {
		return recordType
	}
	return ""
}

func (r *Record) SetType(newType string) {
	klog.Error("Changing record type is not allowed!") // not implemented
}

func (r *Record) Notes() string {
	if recordNotes, ok := r.RecordDict["notes"].(string); ok {
		return recordNotes
	}
	return ""
}

func (r *Record) SetNotes(notes string) {
	if _, ok := r.RecordDict["notes"]; ok {
		r.RecordDict["notes"] = notes
		r.update()
	}
}

func (r *Record) GetFieldsBySection(fieldSectionType FieldSectionFlag) []interface{} {
	fields := []interface{}{}
	if fieldSectionType&FieldSectionFields == FieldSectionFields {
		if iFields, ok := r.RecordDict["fields"]; ok {
			if aFields, ok := iFields.([]interface{}); ok {
				fields = append(fields, aFields...)
			}
		}
	}

	if fieldSectionType&FieldSectionCustom == FieldSectionCustom {
		if iFields, ok := r.RecordDict["custom"]; ok {
			if aFields, ok := iFields.([]interface{}); ok {
				fields = append(fields, aFields...)
			}
		}
	}
	return fields
}

// GetFieldsByMask returns all fields from the corresponding field section (fields, custom or both)
// where fieldToken matches the FieldTokenFlag (type, label or both)
func (r *Record) GetFieldsByMask(fieldToken string, fieldTokenFlag FieldTokenFlag, fieldType FieldSectionFlag) []map[string]interface{} {
	result := []map[string]interface{}{}

	fields := r.GetFieldsBySection(fieldType)

	for i := range fields {
		if fmap, ok := fields[i].(map[string]interface{}); ok {
			val := map[string]interface{}{}
			if fieldTokenFlag&FieldTokenType == FieldTokenType {
				if fType, ok := fmap["type"].(string); ok && fType == fieldToken {
					val = fmap
				}
			}
			if len(val) == 0 && fieldTokenFlag&FieldTokenLabel == FieldTokenLabel {
				if fLabel, ok := fmap["label"].(string); ok && fLabel == fieldToken {
					val = fmap
				}
			}
			if len(val) > 0 {
				result = append(result, val)
			}
		}
	}

	return result
}

func (r *Record) GetFieldsByType(fieldType string) []map[string]interface{} {
	return r.GetFieldsByMask(fieldType, FieldTokenType, FieldSectionFields)
}

func (r *Record) GetFieldsByLabel(fieldLabel string) []map[string]interface{} {
	return r.GetFieldsByMask(fieldLabel, FieldTokenLabel, FieldSectionFields)
}

func (r *Record) GetCustomFieldsByType(fieldType string) []map[string]interface{} {
	return r.GetFieldsByMask(fieldType, FieldTokenType, FieldSectionCustom)
}

func (r *Record) GetCustomFieldsByLabel(fieldLabel string) []map[string]interface{} {
	return r.GetFieldsByMask(fieldLabel, FieldTokenLabel, FieldSectionCustom)
}

// GetFieldValueByType returns string value of the *first* field from fields[] that matches fieldType
func (r *Record) GetFieldValueByType(fieldType string) string {
	if fieldType == "" {
		return ""
	}

	values := []string{}
	if fields := r.GetFieldsByType(fieldType); len(fields) > 0 {
		if iValues, ok := fields[0]["value"].([]interface{}); ok {
			for i := range iValues {
				val := iValues[i]
				// JavaScript has no integers but only one number type, IEEE754 double precision float
				if fval, ok := val.(float64); ok && fval == float64(int(fval)) {
					val = int(fval) // convert to int
				}
				values = append(values, fmt.Sprintf("%v", val))
			}
		}
	}

	return strings.Join(values, ", ")
}

// GetFieldValueByLabel returns string value of the *first* field from fields[] that matches fieldLabel
func (r *Record) GetFieldValueByLabel(fieldLabel string) string {
	if fieldLabel == "" {
		return ""
	}

	values := []string{}
	if fields := r.GetFieldsByLabel(fieldLabel); len(fields) > 0 {
		if iValues, ok := fields[0]["value"].([]interface{}); ok {
			for i := range iValues {
				values = append(values, fmt.Sprintf("%v", iValues[i]))
			}
		}
	}

	return strings.Join(values, ", ")
}

func (r *Record) GetFieldValuesByType(fieldType string) []string {
	values := []string{}
	if fieldType == "" {
		return values
	}

	if fields := r.GetFieldsByType(fieldType); len(fields) > 0 {
		if iValues, ok := fields[0]["value"].([]interface{}); ok {
			for i := range iValues {
				values = append(values, fmt.Sprintf("%v", iValues[i]))
			}
		}
	}

	return values
}

func (r *Record) GetCustomFieldValues(label string, fieldType string) []string {
	values := []string{}
	if label == "" && fieldType == "" {
		return values
	}

	fields := []map[string]interface{}{}
	if fieldType != "" {
		if flds := r.GetCustomFieldsByType(fieldType); len(flds) > 0 {
			for _, fld := range flds {
				if iLabel, ok := fld["label"].(string); label == "" || (ok && label == iLabel) {
					fields = append(fields, fld)
				}
			}
		}
	} else if label != "" {
		if flds := r.GetCustomFieldsByLabel(label); len(flds) > 0 {
			for _, fld := range flds {
				if iType, ok := fld["type"].(string); fieldType == "" || (ok && fieldType == iType) {
					fields = append(fields, fld)
				}
			}
		}
	}

	for _, field := range fields {
		if iValues, ok := field["value"].([]interface{}); ok {
			for i := range iValues {
				values = append(values, fmt.Sprintf("%v", iValues[i]))
			}
		}
	}

	return values
}

// GetCustomFieldValueByType returns string value of the *first* field from custom[] that matches fieldType
func (r *Record) GetCustomFieldValueByType(fieldType string) string {
	if fieldType == "" {
		return ""
	}

	values := []string{}
	if fields := r.GetCustomFieldsByType(fieldType); len(fields) > 0 {
		if iValues, ok := fields[0]["value"].([]interface{}); ok {
			for i := range iValues {
				values = append(values, fmt.Sprintf("%v", iValues[i]))
			}
		}
	}

	result := ""
	if len(values) == 1 {
		result = values[0]
	} else if len(values) > 1 {
		result = strings.Join(values, ", ")
	}
	return result
}

// GetCustomFieldValueByLabel returns string value of the *first* field from custom[] that matches fieldLabel
func (r *Record) GetCustomFieldValueByLabel(fieldLabel string) string {
	if fieldLabel == "" {
		return ""
	}

	values := []string{}
	if fields := r.GetCustomFieldsByLabel(fieldLabel); len(fields) > 0 {
		if iValues, ok := fields[0]["value"].([]interface{}); ok {
			for i := range iValues {
				values = append(values, fmt.Sprintf("%v", iValues[i]))
			}
		}
	}

	result := ""
	if len(values) == 1 {
		result = values[0]
	} else if len(values) > 1 {
		result = strings.Join(values, ", ")
	}
	return result
}

func NewRecordFromRecordData(recordData *RecordCreate, folder *Folder) *Record {
	recordKey, err := GenerateRandomBytes(32)
	if err != nil {
		return nil
	}
	recordUid, err := GenerateRandomBytes(16)
	if err != nil {
		return nil
	}

	return &Record{
		RecordKeyBytes: recordKey,
		Uid:            BytesToUrlSafeStr(recordUid),
		folderKeyBytes: folder.key,
		folderUid:      folder.uid,
		recordType:     recordData.RecordType,
		RawJson:        recordData.ToJson(),
		RecordDict:     recordData.ToDict(),
	}
}

func NewRecordFromRecordDataWithUid(recordUid string, recordData *RecordCreate, folder *Folder) *Record {
	// recordUid must be a base64 url safe encoded string (UID binary length is 16 bytes)
	ruid := UrlSafeStrToBytes(recordUid)
	if len(ruid) != 16 {
		if newUid, err := GenerateRandomBytes(16); err == nil {
			ruid = newUid
		} else {
			return nil
		}
	}

	recordKey, err := GenerateRandomBytes(32)
	if err != nil {
		return nil
	}

	return &Record{
		RecordKeyBytes: recordKey,
		Uid:            BytesToUrlSafeStr(ruid),
		folderKeyBytes: folder.key,
		folderUid:      folder.uid,
		recordType:     recordData.RecordType,
		RawJson:        recordData.ToJson(),
		RecordDict:     recordData.ToDict(),
	}
}

func NewRecordFromJson(recordDict map[string]interface{}, secretKey []byte, folderUid string) *Record {
	record := Record{}

	// if folderUid is present then secretKey is the folder key
	// if folderUid is empty then record is directly shared to the app and secretKey is the app key
	if strings.TrimSpace(folderUid) != "" {
		record.folderUid = folderUid
		record.folderKeyBytes = secretKey
	}

	if uid, ok := recordDict["recordUid"]; ok {
		record.Uid = strings.TrimSpace(uid.(string))
	}
	if innerFolderUid, ok := recordDict["innerFolderUid"]; ok {
		if ifuid, ok := innerFolderUid.(string); ok {
			record.innerFolderUid = strings.TrimSpace(ifuid)
		}
	}
	if revision, ok := recordDict["revision"].(float64); ok {
		record.Revision = int64(revision)
	}
	if isEditable, ok := recordDict["isEditable"].(bool); ok {
		record.IsEditable = isEditable
	}

	recordKeyEncryptedStr := ""
	if recKey, ok := recordDict["recordKey"]; ok {
		recordKeyEncryptedStr = strings.TrimSpace(recKey.(string))
	}

	if recordKeyEncryptedStr != "" {
		//Folder Share
		recordKeyEncryptedBytes := Base64ToBytes(recordKeyEncryptedStr)
		if recordKeyBytes, err := Decrypt(recordKeyEncryptedBytes, secretKey); err == nil {
			record.RecordKeyBytes = recordKeyBytes
		} else {
			klog.Error("error decrypting record key: " + err.Error() + " - Record UID: " + record.Uid)
		}
	} else {
		//Single Record Share
		record.RecordKeyBytes = secretKey
	}

	if recordEncryptedData, ok := recordDict["data"]; ok && len(record.RecordKeyBytes) > 0 {
		strRecordEncryptedData := recordEncryptedData.(string)
		if recordDataJson, err := DecryptRecord(Base64ToBytes(strRecordEncryptedData), record.RecordKeyBytes); err == nil {
			record.RawJson = recordDataJson
			record.RecordDict = JsonToDict(record.RawJson)
		} else {
			klog.Error("error decrypting record data: " + err.Error())
		}
	}

	if recordType, ok := record.RecordDict["type"]; ok {
		record.recordType = recordType.(string)
	}

	// files
	if recordFiles, ok := recordDict["files"]; ok {
		if rfSlice, ok := recordFiles.([]interface{}); ok {
			for i := range rfSlice {
				if rfMap, ok := rfSlice[i].(map[string]interface{}); ok {
					if file := NewKeeperFileFromJson(rfMap, record.RecordKeyBytes); file != nil {
						record.Files = append(record.Files, file)
					}
				}
			}
		}
	}

	return &record
}

// FindFileByTitle finds the first file with matching title
func (r *Record) FindFileByTitle(title string) *KeeperFile {
	for i := range r.Files {
		if r.Files[i].Title == title {
			return r.Files[i]
		}
	}
	return nil
}

// FindFileByName finds the first file with matching filename
func (r *Record) FindFileByFilename(filename string) *KeeperFile {
	for i := range r.Files {
		if r.Files[i].Name == filename {
			return r.Files[i]
		}
	}
	return nil
}

// FindFile finds the first file with matching file UID, name or title
func (r *Record) FindFile(name string) *KeeperFile {
	for i := range r.Files {
		if r.Files[i].Uid == name || r.Files[i].Name == name || r.Files[i].Title == name {
			return r.Files[i]
		}
	}
	return nil
}

// FindFiles finds all files with matching file UID, name or title
func (r *Record) FindFiles(name string) []*KeeperFile {
	result := []*KeeperFile{}
	for i := range r.Files {
		if r.Files[i].Uid == name || r.Files[i].Name == name || r.Files[i].Title == name {
			result = append(result, r.Files[i])
		}
	}
	return result
}

func (r *Record) DownloadFileByTitle(title string, path string) bool {
	if foundFile := r.FindFileByTitle(title); foundFile != nil {
		return foundFile.SaveFile(path, false)
	}
	return false
}

func (r *Record) DownloadFile(fileUid string, path string) bool {
	for i := range r.Files {
		if r.Files[i].Uid == fileUid {
			return r.Files[i].SaveFile(path, false)
		}
	}
	return false
}

func (r *Record) ToString() string {
	return fmt.Sprintf("[Record: UID=%s, revision=%d, editable=%t, type: %s, title: %s, files count: %d]", r.Uid, r.Revision, r.IsEditable, r.recordType, r.Title(), len(r.Files))
}

func (r *Record) update() {
	// Record class works directly on fields in recordDict here we only update the raw JSON
	r.RawJson = DictToJson(r.RecordDict)
}

func (r *Record) value(values []interface{}, single bool) []interface{} {
	if len(values) == 0 {
		return []interface{}{}
	}
	if single {
		return []interface{}{values[0]}
	}
	return values
}

func (r *Record) fieldSearch(fields []interface{}, fieldKey string) map[string]interface{} {
	// This is a generic field search that returns the field
	// It will work for for both standard and custom fields.
	// It returns the field as a map[string]interface{}.

	foundItem := map[string]interface{}{}
	if len(fields) == 0 {
		return foundItem
	}

	// First check in the field_key matches any labels. Label matching is case sensitive.
	for _, item := range fields {
		if iValue, ok := item.(map[string]interface{}); ok {
			if iLabel, found := iValue["label"]; found {
				if sLabel, ok := iLabel.(string); ok && strings.EqualFold(sLabel, fieldKey) {
					foundItem = iValue
					break
				}
			}
		}
	}
	// If the label was not found, check the field type. Field type is case insensitive.
	if len(foundItem) == 0 {
		for _, item := range fields {
			if iValue, ok := item.(map[string]interface{}); ok {
				if iType, found := iValue["type"]; found {
					if sType, ok := iType.(string); ok && strings.EqualFold(sType, fieldKey) {
						foundItem = iValue
						break
					}
				}
			}
		}
	}
	return foundItem
}

func (r *Record) getStandardField(fieldType string) map[string]interface{} {
	if iFields, found := r.RecordDict["fields"]; found {
		if sFields, ok := iFields.([]interface{}); ok {
			return r.fieldSearch(sFields, fieldType)
		}
	}
	return map[string]interface{}{}
}

func (r *Record) GetStandardFieldValue(fieldType string, single bool) ([]interface{}, error) {
	field := r.getStandardField(fieldType)
	if len(field) == 0 {
		return nil, fmt.Errorf("cannot find standard field %s in record", fieldType)
	}
	sValue := []interface{}{}
	if iValue, found := field["value"]; found {
		if sVal, ok := iValue.([]interface{}); ok {
			sValue = sVal
		}
	}
	return r.value(sValue, single), nil
}

func (r *Record) SetStandardFieldValue(fieldType string, value interface{}) error {
	field := r.getStandardField(fieldType)
	if len(field) == 0 {
		return fmt.Errorf("cannot find standard field %s in record", fieldType)
	}
	if _, ok := value.([]interface{}); !ok {
		value = []interface{}{value}
	}
	field["value"] = value
	r.update()
	return nil
}

func (r *Record) FieldExists(section, name string) bool {
	result := false
	if section != "fields" && section != "custom" {
		return result
	}

	if rfi, found := r.RecordDict[section]; found && rfi != nil {
		if rfsi, ok := rfi.([]interface{}); ok && len(rfsi) > 0 {
			for _, v := range rfsi {
				if fmap, ok := v.(map[string]interface{}); ok {
					if ftype, found := fmap["type"]; found && name == fmt.Sprint(ftype) {
						result = true
						break
					}
				}
			}
		}
	}

	return result
}

func (r *Record) RemoveField(section, name string, removeAll bool) int {
	removed := 0
	if section != "fields" && section != "custom" {
		return removed
	}

	if rfi, found := r.RecordDict[section]; found && rfi != nil {
		if rfsi, ok := rfi.([]interface{}); ok && len(rfsi) > 0 {
			ix := []int{}
			for i, v := range rfsi {
				if fmap, ok := v.(map[string]interface{}); ok {
					if ftype, found := fmap["type"]; found && name == fmt.Sprint(ftype) {
						ix = append(ix, i)
					}
				}
			}
			if len(ix) > 1 && !removeAll {
				ix = ix[:1]
			}
			if len(ix) > 0 {
				// work on a copy since slices do not support in-place operations
				rfsic := make([]interface{}, len(rfsi))
				copy(rfsic, rfsi)

				sort.Sort(sort.Reverse(sort.IntSlice(ix)))
				for _, i := range ix {
					removed++
					rfsic = append(rfsic[:i], rfsic[i+1:]...)
				}
				r.RecordDict[section] = rfsic
			}
		}
	}

	return removed
}

func (r *Record) InsertField(section string, field interface{}) error {
	if section != "fields" && section != "custom" {
		return fmt.Errorf("unknown field section '%s'", section)
	}
	if !IsFieldClass(field) {
		return fmt.Errorf("field is not a vaild field class")
	}

	if rfi, found := r.RecordDict[section]; !found || rfi == nil {
		r.RecordDict[section] = []interface{}{}
	}
	if rfi, found := r.RecordDict[section]; found && rfi != nil {
		if rfsi, ok := rfi.([]interface{}); ok {
			// work on a copy since slices do not support in-place operations
			rfsic := make([]interface{}, len(rfsi))
			copy(rfsic, rfsi)
			if fmap, err := structToMap(field); err == nil {
				rfsic = append(rfsic, fmap)
			} else {
				return fmt.Errorf("error converting field %v - Error: %s", field, err.Error())
			}
			r.RecordDict[section] = rfsic
		} else {
			return fmt.Errorf("section '%s' is not in the expected format - expected []interface{}", section)
		}
	} else {
		return fmt.Errorf("section '%s' not found", section)
	}

	return nil
}

func (r *Record) UpdateField(section string, field interface{}) error {
	if section != "fields" && section != "custom" {
		return fmt.Errorf("unknown field section '%s'", section)
	}
	if !IsFieldClass(field) {
		return fmt.Errorf("field is not a vaild field class")
	}

	fieldMap, err := structToMap(field)
	if err != nil {
		return fmt.Errorf("error converting field %v - Error: %s", field, err.Error())
	}

	fieldType := ""
	if fType, found := fieldMap["type"]; found {
		fieldType = strings.TrimSpace(fmt.Sprint(fType))
	}
	if fieldType == "" {
		return fmt.Errorf("error - missing field type in field: %v", field)
	}

	if rfi, found := r.RecordDict[section]; !found || rfi == nil {
		r.RecordDict[section] = []interface{}{}
	}
	if rfi, found := r.RecordDict[section]; found && rfi != nil {
		if rfsi, ok := rfi.([]interface{}); ok {
			for _, v := range rfsi {
				if fmap, ok := v.(map[string]interface{}); ok {
					if ftype, found := fmap["type"]; found && fieldType == fmt.Sprint(ftype) {
						for key := range fmap {
							delete(fmap, key)
						}
						for key, val := range fieldMap {
							fmap[key] = val
						}
						return nil
					}
				}
			}
		} else {
			return fmt.Errorf("section '%s' is not in the expected format - expected []interface{}", section)
		}
	} else {
		return fmt.Errorf("section '%s' not found", section)
	}

	return fmt.Errorf("field type '%s' not found", fieldType)
}

func (r *Record) getCustomField(fieldType string) map[string]interface{} {
	if iFields, found := r.RecordDict["custom"]; found {
		if sFields, ok := iFields.([]interface{}); ok {
			return r.fieldSearch(sFields, fieldType)
		}
	}
	return map[string]interface{}{}
}

func (r *Record) GetCustomFieldValue(fieldType string, single bool) ([]interface{}, error) {
	field := r.getCustomField(fieldType)
	if len(field) == 0 {
		return nil, fmt.Errorf("cannot find custom field %s in record", fieldType)
	}
	sValue := []interface{}{}
	if iValue, found := field["value"]; found {
		if sVal, ok := iValue.([]interface{}); ok {
			sValue = sVal
		}
	}
	return r.value(sValue, single), nil
}

func (r *Record) SetCustomFieldValue(fieldType string, value interface{}) error {
	field := r.getCustomField(fieldType)
	if len(field) == 0 {
		return fmt.Errorf("cannot find custom field %s in record", fieldType)
	}
	if _, ok := value.([]interface{}); !ok {
		value = []interface{}{value}
	}
	field["value"] = value
	r.update()
	return nil
}

// AddCustomField adds new custom field to the record
// The new field must satisfy the IsFieldClass function
func (r *Record) AddCustomField(field interface{}) error {
	if !IsFieldClass(field) {
		return fmt.Errorf("cannot add custom field - unknown field type for %v ", field)
	}

	var iCustom interface{} = []interface{}{}
	if iFields, found := r.RecordDict["custom"]; found {
		iCustom = iFields
	} else {
		r.RecordDict["custom"] = iCustom
	}

	if sCustom, ok := iCustom.([]interface{}); ok {
		if fmap := ObjToDict(field); fmap != nil {
			sCustom = append(sCustom, fmap)
			r.RecordDict["custom"] = sCustom
			r.update()
			return nil
		} else {
			return fmt.Errorf("cannot add custom field - error converting to JSON, field: %v ", field)
		}
	} else {
		return fmt.Errorf("cannot add custom field - custom[] is not the expected array type, custom: %v ", iCustom)
	}
}

func (r *Record) CanClone() bool {
	if strings.TrimSpace(r.folderUid) != "" && len(r.folderKeyBytes) > 0 {
		return true
	} else {
		return false
	}
}

func findTemplateRecord(templateRecordUid string, records []*Record) (*Record, error) {
	var templateRecord *Record = nil

	for _, r := range records {
		if r.Uid == templateRecordUid {
			templateRecord = r
			if strings.TrimSpace(r.folderUid) != "" && len(r.folderKeyBytes) > 0 {
				break
			}
		}
	}

	if templateRecord == nil {
		return nil, fmt.Errorf("cannot find template record '%s' in record", templateRecordUid)
	}

	// Records shared directly to the application cannot be used as template records
	// only records in a shared folder (shared to the application) should be used as templates
	if strings.TrimSpace(templateRecord.folderUid) == "" || len(templateRecord.folderKeyBytes) == 0 {
		return nil, fmt.Errorf("found matching template record %s which is not in a shared folder", templateRecordUid)
	}

	return templateRecord, nil
}

// NewRecordClone returns a deep copy of the template object with new UID and RecordKeyBytes
// generates and uses new random UID if newRecordUid is empty
// returns error if template record is not found
func NewRecordClone(templateRecordUid string, records []*Record, newRecordUid string) (*Record, error) {
	templateRecord, err := findTemplateRecord(templateRecordUid, records)
	if err != nil {
		return nil, err
	}

	recordKeyBytes, _ := GetRandomBytes(32)
	folderKeyBytesCopy := make([]byte, len(templateRecord.folderKeyBytes))
	copy(folderKeyBytesCopy, templateRecord.folderKeyBytes)
	recordDictCopy := CopyableMap(templateRecord.RecordDict).DeepCopy()

	filesCopy := []*KeeperFile{}
	for _, f := range templateRecord.Files {
		filesCopy = append(filesCopy, f.DeepCopy())
	}

	recordUid := GenerateUid()
	if ruid := strings.TrimSpace(newRecordUid); ruid != "" {
		if numBytes := len(Base64ToBytes(ruid)); numBytes == 16 {
			recordUid = newRecordUid
		}
	}
	if newRecordUid != "" && recordUid != newRecordUid {
		klog.Warning("invalid new record UID provided:", newRecordUid, " - using autogenerated UID:", recordUid)
	}

	rec := &Record{
		RecordKeyBytes: recordKeyBytes,
		Uid:            recordUid,
		folderKeyBytes: folderKeyBytesCopy,
		folderUid:      templateRecord.folderUid,
		innerFolderUid: templateRecord.innerFolderUid,
		Files:          filesCopy,
		Revision:       templateRecord.Revision,
		IsEditable:     templateRecord.IsEditable,
		recordType:     templateRecord.recordType,
		RawJson:        templateRecord.RawJson,
		RecordDict:     recordDictCopy,
	}

	return rec, nil
}

// NewRecord returns a new empty record of the same type as template object but with new UID and RecordKeyBytes
// generates and uses new random UID if newRecordUid is empty
// returns error if template record is not found
func NewRecord(templateRecordUid string, records []*Record, newRecordUid string) (*Record, error) {
	templateRecord, err := findTemplateRecord(templateRecordUid, records)
	if err != nil {
		return nil, err
	}

	recordKeyBytes, _ := GetRandomBytes(32)
	folderKeyBytesCopy := make([]byte, len(templateRecord.folderKeyBytes))
	copy(folderKeyBytesCopy, templateRecord.folderKeyBytes)

	// copy and preserve known keys but clear all other values except record type
	// drop custom[] and any other unknown top-level keys
	recordDictCopy := CopyableMap(templateRecord.RecordDict).DeepCopy()
	for key, val := range recordDictCopy {
		switch key {
		case "type":
			continue
		case "title", "notes":
			recordDictCopy[key] = ""
		case "custom":
			delete(recordDictCopy, key)
		case "fields":
			if fslice, ok := val.([]interface{}); ok {
				for _, fs := range fslice {
					if fmap, ok := fs.(map[string]interface{}); ok {
						for fkey := range fmap {
							// preserve field type, clear label and value, drop everything else
							switch fkey {
							case "type":
								continue
							case "label":
								fmap[fkey] = ""
							case "value":
								fmap[fkey] = []interface{}{}
							default:
								klog.Warning("create new record - removing field type property", key)
								delete(fmap, key)
							}
						}
					} else {
						klog.Warning("create new record - fields type is not in the expected format and is removed")
					}
				}
			} else {
				klog.Warning("create new record - fields[] is not in the expected format and is replaced")
				recordDictCopy[key] = []interface{}{}
			}
		default:
			klog.Warning("create new record - removing unknown record type property", key)
			delete(recordDictCopy, key)
		}
	}
	rawJson := DictToJson(recordDictCopy)

	recordUid := GenerateUid()
	if ruid := strings.TrimSpace(newRecordUid); ruid != "" {
		if numBytes := len(Base64ToBytes(ruid)); numBytes == 16 {
			recordUid = newRecordUid
		}
	}
	if newRecordUid != "" && recordUid != newRecordUid {
		klog.Warning("invalid new record UID provided:", newRecordUid, " - using autogenerated UID:", recordUid)
	}

	rec := &Record{
		RecordKeyBytes: recordKeyBytes,
		Uid:            recordUid,
		folderKeyBytes: folderKeyBytesCopy,
		folderUid:      templateRecord.folderUid,
		innerFolderUid: templateRecord.innerFolderUid,
		Files:          []*KeeperFile{},
		Revision:       templateRecord.Revision,
		IsEditable:     templateRecord.IsEditable,
		recordType:     templateRecord.recordType,
		RawJson:        rawJson,
		RecordDict:     recordDictCopy,
	}

	return rec, nil
}

func (r *Record) Print() {
	fmt.Println("===")
	fmt.Println("Title: " + r.Title())
	fmt.Println("UID:   " + r.Uid)
	fmt.Println("Type:  " + r.Type())

	fmt.Println()
	fmt.Println("Fields")
	fmt.Println("------")
	skipFileds := map[string]struct{}{"fileRef": {}, "oneTimeCode": {}}
	if _fields, ok := r.RecordDict["fields"]; ok {
		if fields, ok := _fields.([]interface{}); ok {
			for i := range fields {
				if fmap, ok := fields[i].(map[string]interface{}); ok {
					ftype, _ := fmap["type"].(string)
					// flabel, _ := fmap["label"].(string)
					if _, found := skipFileds[ftype]; !found {
						fmt.Printf("%s : %v\n", ftype, fmap["value"]) // ", ".join(item["value"]
					}
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("Custom Fields")
	fmt.Println("------")
	if _fields, ok := r.RecordDict["custom"]; ok {
		if fields, ok := _fields.([]interface{}); ok {
			for i := range fields {
				if fmap, ok := fields[i].(map[string]interface{}); ok {
					ftype, _ := fmap["type"].(string)
					flabel, _ := fmap["label"].(string)
					fmt.Printf("%s (%s) : %v\n", ftype, flabel, fmap["value"]) // ", ".join(item["value"]
				}
			}
		}
	}
}

type KeeperFolder struct {
	FolderKey []byte
	FolderUid string
	ParentUid string
	Name      string
}

func NewKeeperFolder(folderMap map[string]interface{}, folderKey []byte) *KeeperFolder {
	folder := KeeperFolder{FolderKey: folderKey}
	if key, found := folderMap["folderUid"]; found {
		if val, ok := key.(string); ok {
			folder.FolderUid = val
		}
	}
	if key, found := folderMap["parent"]; found {
		if val, ok := key.(string); ok {
			folder.ParentUid = val
		}
	}
	if key, found := folderMap["data"]; found {
		if val, ok := key.(string); ok {
			if folderNameJson, err := DecryptAesCbc(UrlSafeStrToBytes(val), folderKey); err == nil {
				folderName := struct {
					Name string `json:"name"`
				}{}
				if err := json.Unmarshal(folderNameJson, &folderName); err == nil {
					folder.Name = folderName.Name
				} else {
					klog.Error("error parsing folder name: " + err.Error())
				}
			}
		}
	}

	return &folder
}

type Folder struct {
	key           []byte
	uid           string
	ParentUid     string
	Name          string
	data          map[string]interface{}
	folderRecords []map[string]interface{}
}

func NewFolderFromJson(folderDict map[string]interface{}, secretKey []byte) *Folder {
	folder := Folder{
		data: folderDict,
	}
	if uid, ok := folderDict["folderUid"]; ok {
		folder.uid = strings.TrimSpace(uid.(string))
		// only /get_folders retrieves parent and name/data
		if folderKeyEnc, ok := folderDict["folderKey"]; ok {
			if folderKey, err := Decrypt(Base64ToBytes(folderKeyEnc.(string)), secretKey); err == nil {
				folder.key = folderKey
				if folderRecords, ok := folderDict["records"]; ok {
					if iFolderRecords, ok := folderRecords.([]interface{}); ok {
						for i := range iFolderRecords {
							if folderRecord, ok := iFolderRecords[i].(map[string]interface{}); ok {
								folder.folderRecords = append(folder.folderRecords, folderRecord)
							}
						}
					} else {
						klog.Error("folder records JSON is in incorrect format")
					}
				}
			} else {
				klog.Error("error decrypting folder key: " + err.Error())
			}
		}
	} else {
		klog.Error("Not a folder")
		return nil
	}

	return &folder
}

func (f *Folder) Records() []*Record {
	records := []*Record{}
	if f.folderRecords != nil {
		for _, r := range f.folderRecords {
			if record := NewRecordFromJson(r, f.key, f.uid); record.Uid != "" {
				records = append(records, record)
			} else {
				klog.Error("error parsing folder record: ", r)
			}
		}
	}
	return records
}

type KeeperFile struct {
	FileKey  string
	metaDict map[string]interface{}

	FileData []byte

	Uid          string
	Type         string
	Title        string
	Name         string
	LastModified int
	Size         int

	F              map[string]interface{}
	RecordKeyBytes []byte
}

func NewKeeperFileFromJson(fileDict map[string]interface{}, recordKeyBytes []byte) *KeeperFile {
	f := &KeeperFile{
		F:              fileDict,
		RecordKeyBytes: recordKeyBytes,
	}

	// Set file metadata
	meta := f.GetMeta()

	if fuid, ok := fileDict["fileUid"].(string); ok {
		f.Uid = fuid
	}
	if recordType, ok := meta["type"].(string); ok {
		f.Type = recordType
	}
	if title, ok := meta["title"].(string); ok {
		f.Title = title
	}
	if name, ok := meta["name"].(string); ok {
		f.Name = name
	}
	if lastModified, ok := meta["lastModified"].(float64); ok {
		f.LastModified = int(lastModified)
	}
	if size, ok := meta["size"].(float64); ok {
		f.Size = int(size)
	}

	return f
}

func (f *KeeperFile) DeepCopy() *KeeperFile {
	return &KeeperFile{
		FileKey:        f.FileKey,
		metaDict:       CopyableMap(f.metaDict).DeepCopy(),
		FileData:       CloneByteSlice(f.FileData),
		Uid:            f.Uid,
		Type:           f.Type,
		Title:          f.Title,
		Name:           f.Name,
		LastModified:   f.LastModified,
		Size:           f.Size,
		F:              CopyableMap(f.F).DeepCopy(),
		RecordKeyBytes: CloneByteSlice(f.RecordKeyBytes),
	}
}

func (f *KeeperFile) DecryptFileKey() []byte {
	fileKeyEncryptedBase64 := f.F["fileKey"]
	fileKeyEncryptedBase64Str := fmt.Sprintf("%v", fileKeyEncryptedBase64)
	fileKeyEncrypted := Base64ToBytes(fileKeyEncryptedBase64Str)
	if fileKey, err := Decrypt(fileKeyEncrypted, f.RecordKeyBytes); err == nil {
		return fileKey
	} else {
		klog.Error("error decrypting file key " + fileKeyEncryptedBase64Str)
		return []byte{}
	}
}

func (f *KeeperFile) GetMeta() map[string]interface{} {
	// Returns file metadata dictionary (file name, title, size, type, etc.)
	if len(f.metaDict) == 0 {
		if data, ok := f.F["data"]; ok && data != nil {
			fileKey := f.DecryptFileKey()
			dataStr := fmt.Sprintf("%v", data)
			if metaJson, err := Decrypt(Base64ToBytes(dataStr), fileKey); err == nil {
				f.metaDict = JsonToDict(string(metaJson[:]))
			} else {
				klog.Error("error parsing file meta data " + dataStr)
			}
		}
	}
	return f.metaDict
}

func (f *KeeperFile) GetUrl() string {
	if url, ok := f.F["url"].(string); ok {
		return url
	}
	return ""
}

func (f *KeeperFile) GetFileData() []byte {
	// Return decrypted raw file data
	if len(f.FileData) == 0 { // cached if nothing
		fileKey := f.DecryptFileKey()
		if fileUrl, ok := f.F["url"]; ok && fileUrl != nil {
			fileUrlStr := fmt.Sprintf("%v", fileUrl)
			if rs, err := http.Get(fileUrlStr); err == nil {
				defer rs.Body.Close()
				if fileEncryptedData, err := ioutil.ReadAll(rs.Body); err == nil {
					if fileData, err := Decrypt(fileEncryptedData, fileKey); err == nil {
						f.FileData = fileData
					}
				}
			}
		}
	}
	return f.FileData
}

func (f *KeeperFile) SaveFile(path string, createFolders bool) bool {
	// Save decrypted file data to the provided path
	if createFolders {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			klog.Error("error creating folders " + err.Error())
		}
	}

	pathExists := false
	if absPath, err := filepath.Abs(path); err == nil {
		dirPath := filepath.Dir(absPath)
		if found, _ := PathExists(dirPath); found {
			pathExists = true
		}
	}

	if !pathExists {
		klog.Error("No such file or directory %s\nConsider using `SaveFile()` method with `createFolders=True` ", path)
		return false
	}

	fileData := f.GetFileData()
	if err := ioutil.WriteFile(path, fileData, 0644); err != nil {
		klog.Error("error savig file " + err.Error())
	}

	return true
}

func (f *KeeperFile) ToString() string {
	return fmt.Sprintf("[KeeperFile - name: %s, title: %s]", f.Name, f.Title)
}

type KeeperFileUpload struct {
	Name  string
	Title string
	Type  string
	Data  []byte
}

func GetFileForUpload(filePath, fileName, fileTitle, mimeType string) (*KeeperFileUpload, error) {
	// Helper method to get KeeperFileUpload struct from the file path
	if fileName == "" {
		fileName = path.Base(filePath)
	}
	if fileTitle == "" {
		fileTitle = fileName
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if fileDataBytes, err := ioutil.ReadFile(filePath); err == nil {
		return &KeeperFileUpload{
			Name:  fileName,
			Title: fileTitle,
			Type:  mimeType,
			Data:  fileDataBytes,
		}, nil
	} else {
		return nil, err
	}
}

type KeeperFileData struct {
	Title        string `json:"title,omitempty"`
	Name         string `json:"name,omitempty"`
	Type         string `json:"type,omitempty"`
	Size         int64  `json:"size,omitempty"`
	LastModified int64  `json:"lastModified,omitempty"`
}

type RecordField struct {
	Type     string
	Label    string
	Value    []interface{}
	Required bool
}

func NewRecordField(fieldType, label string, required bool, value interface{}) *RecordField {
	recordField := &RecordField{
		Type:     fieldType,
		Label:    label,
		Required: required,
	}
	if iValue, ok := value.([]interface{}); ok {
		recordField.Value = iValue
	} else if value == nil {
		recordField.Value = []interface{}{}
	} else {
		recordField.Value = []interface{}{value}
	}
	return recordField
}

type RecordCreate struct {
	RecordType string        `json:"type,omitempty"`
	Title      string        `json:"title,omitempty"`
	Notes      string        `json:"notes,omitempty"`
	Fields     []interface{} `json:"fields,omitempty"`
	Custom     []interface{} `json:"custom,omitempty"`
}

func NewRecordCreate(recordType, title string) *RecordCreate {
	return &RecordCreate{
		RecordType: recordType,
		Title:      title,
		Fields:     []interface{}{},
		Custom:     []interface{}{},
	}
}

func NewRecordCreateFromJson(recordJson string) *RecordCreate {
	// NB! this will silently ignore any unknown record and field attributes
	// NB! Do not serialize back to record or field for update - use only for record create
	rc := getRecordCreateFromJson(recordJson)
	if rc != nil {
		fields := []interface{}{}
		custom := []interface{}{}
		for _, fMap := range rc.Fields {
			if fld, err := convertToKeeperRecordField(fMap, false); err == nil {
				fields = append(fields, fld)
			} else {
				klog.Warning("skipped field definition due to conversion error(s) - " + err.Error())
			}
		}
		for _, fMap := range rc.Custom {
			if fld, err := convertToKeeperRecordField(fMap, false); err == nil {
				custom = append(custom, fld)
			} else {
				klog.Warning("skipped custom field definition due to conversion error(s) - " + err.Error())
			}
		}
		rc.Fields = fields
		rc.Custom = custom
	}
	return rc
}

func getRecordCreateFromJson(jsonData string) *RecordCreate {
	bytes := []byte(jsonData)
	res := RecordCreate{}

	if err := json.Unmarshal(bytes, &res); err != nil {
		klog.Error("Error deserializing RecordCreate from JSON: " + err.Error())
		return nil
	}
	return &res
}

func NewRecordCreateFromJsonDecoder(recordJson string, disallowUnknownFields bool) (*RecordCreate, error) {
	// NB! JSON mapping is controlled by disallowUnknownFields and may ignore any unknown record and field attributes
	// when disallowUnknownFields is true it is safe to serialize back to record or field for update but
	// when disallowUnknownFields is false avoid serializing RecordCreate back to JSON as may remove any extras
	rc, err := getRecordCreateFromJsonDecoder(recordJson, disallowUnknownFields)
	if err != nil {
		return nil, err
	}
	if rc != nil {
		fields := []interface{}{}
		custom := []interface{}{}
		for _, fMap := range rc.Fields {
			if fld, err := convertToKeeperRecordField(fMap, true); err == nil {
				fields = append(fields, fld)
			} else {
				return nil, err
			}
		}
		for _, fMap := range rc.Custom {
			if fld, err := convertToKeeperRecordField(fMap, true); err == nil {
				custom = append(custom, fld)
			} else {
				return nil, err
			}
		}
		rc.Fields = fields
		rc.Custom = custom
	}
	return rc, nil
}

func getRecordCreateFromJsonDecoder(jsonData string, disallowUnknownFields bool) (*RecordCreate, error) {
	// NB! Cannot validate RecordType because of custom types
	rc := RecordCreate{}

	if disallowUnknownFields {
		decoder := json.NewDecoder(strings.NewReader(jsonData))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&rc); err != nil {
			klog.Error("Error deserializing RecordCreate from strict JSON: " + err.Error())
			return nil, err
		}
	} else {
		if err := json.Unmarshal([]byte(jsonData), &rc); err != nil {
			klog.Error("Error deserializing RecordCreate from JSON: " + err.Error())
			return nil, err
		}
	}

	return &rc, nil
}

func convertToKeeperRecordField(fieldData interface{}, validate bool) (interface{}, error) {
	if fieldData == nil {
		return nil, errors.New("cannot convert empty field data")
	}
	fieldTypes := "|login|password|url|fileRef|oneTimeCode|name" +
		"|birthDate|date|expirationDate|text|securityQuestion|multiline|email|cardRef" +
		"|addressRef|pinCode|phone|secret|note|accountNumber|paymentCard|bankAccount" +
		"|keyPair|host|address|licenseNumber|recordRef|schedule|directoryType|databaseType" +
		"|pamHostname|pamResources|checkbox|script|passkey|"
	if fMap, ok := fieldData.(map[string]interface{}); ok {
		if fType, found := fMap["type"]; found {
			if sType, ok := fType.(string); ok && strings.Contains(fieldTypes, "|"+sType+"|") {
				return getKeeperRecordField(sType, fMap, validate)
			} else {
				return nil, fmt.Errorf("unknown field type %v", fMap)
			}
		} else {
			return nil, fmt.Errorf("field type missing in field data %v", fieldData)
		}
	} else {
		return nil, fmt.Errorf("expected format for field data %v", fieldData)
	}
}

func (r RecordCreate) ToDict() map[string]interface{} {
	recDict := map[string]interface{}{
		"type":   r.RecordType,
		"title":  r.Title,
		"fields": r.Fields,
	}
	if r.Notes != "" {
		recDict["notes"] = r.Notes
	}
	if len(r.Custom) > 0 {
		recDict["custom"] = r.Custom
	}
	return recDict
}

func (r RecordCreate) ToJson() string {
	return DictToJsonWithDefultIndent(r.ToDict())
}

func (r RecordCreate) getFieldsByType(field interface{}, single bool) []interface{} {
	result := []interface{}{}
	if field == nil {
		return result
	}

	fieldPtr := getFieldPtr(field)
	if fieldPtr == nil {
		return result
	}

	iType := reflect.TypeOf(fieldPtr)
	for i, f := range r.Fields {
		fptr := getFieldPtr(f)
		if fType := reflect.TypeOf(fptr); iType == fType {
			result = append(result, fptr)
			if reflect.TypeOf(f).Kind() != reflect.Ptr {
				r.Fields[i] = fptr
			}
			if single {
				return result
			}
		}
	}
	for i, f := range r.Custom {
		fptr := getFieldPtr(f)
		if fType := reflect.TypeOf(fptr); iType == fType {
			result = append(result, fptr)
			if reflect.TypeOf(f).Kind() != reflect.Ptr {
				r.Custom[i] = fptr
			}
			if single {
				return result
			}
		}
	}
	return result
}

// GetFieldsByType returns all fields of the same type as field param
// The search goes first through fields[] then custom[]
// Note: Method returns pointers so any value modifications are reflected directly in the record
func (r RecordCreate) GetFieldsByType(field interface{}) []interface{} {
	result := []interface{}{}
	if field == nil {
		return result
	}

	if records := r.getFieldsByType(field, false); records == nil {
		return result
	} else {
		return records
	}
}

// GetFieldByType returns first found field of the same type as field param
// The search goes first through fields[] then custom[]
// Note: Method returns a pointer so any value modifications are reflected directly in the record
func (r RecordCreate) GetFieldByType(field interface{}) interface{} {
	var result interface{}
	if field == nil {
		return result
	}

	if records := r.getFieldsByType(field, true); len(records) == 0 {
		return result
	} else {
		return records[0]
	}
}

// Return a pointer to the supplied struct via interface{}
func toFieldPtr(obj interface{}) interface{} {
	vp := reflect.New(reflect.TypeOf(obj))
	vp.Elem().Set(reflect.ValueOf(obj))
	return vp.Interface()
}

func getFieldPtr(field interface{}) interface{} {
	// already pointer type
	if fType := reflect.TypeOf(field); fType.Kind() == reflect.Ptr {
		return field
	}

	// struct - passed by value, get pointer
	switch field.(type) {
	case nil:
		return nil
	default:
		return toFieldPtr(field)
	}
}

// Application info
type AppData struct {
	Title   string `json:"title,omitempty"`
	AppType string `json:"type,omitempty"`
}

func NewAppData(title, appType string) *AppData {
	return &AppData{
		Title:   title,
		AppType: appType,
	}
}

// Server response contained details about the application and the records
// that were requested to be returned
type SecretsManagerResponse struct {
	AppData   AppData
	Folders   []*Folder
	Records   []*Record
	ExpiresOn int64
	Warnings  string
	JustBound bool
	// AppOwnerPublicKey string
	// EncryptedAppKey   string
}

// ExpiresOnStr retrieves string formatted expiration date
// if dateFormat is empty default format is used: "%Y-%m-%d %H:%M:%S"
func (r SecretsManagerResponse) ExpiresOnStr(dateFormat string) string {
	unixtimeSeconds := r.ExpiresOn / 1000
	return time.Unix(unixtimeSeconds, 0).Format("2006-01-02 15:04:05")
	// "2006-01-02 15:04:05" = "%Y-%m-%d %H:%M:%S"
	// RFC3339 = "2006-01-02T15:04:05Z07:00"
	// RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
}

type AddFileResponse struct {
	Url               string `json:"url"`
	Parameters        string `json:"parameters"`
	SuccessStatusCode int    `json:"successStatusCode"`
}

func AddFileResponseFromJson(jsonData string) (*AddFileResponse, error) {
	bytes := []byte(jsonData)
	res := AddFileResponse{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		return &res, nil
	} else {
		return nil, fmt.Errorf("Error deserializing AddFileResponse from JSON: " + err.Error())
	}
}

type DeleteSecretResponse struct {
	RecordUid    string `json:"recordUid"`
	ResponseCode string `json:"responseCode"`
	ErrorMessage string `json:"errorMessage"`
}

type DeleteSecretsResponse struct {
	Records []DeleteSecretResponse `json:"records"`
}

func DeleteSecretsResponseFromJson(jsonData string) (*DeleteSecretsResponse, error) {
	bytes := []byte(jsonData)
	res := DeleteSecretsResponse{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		return &res, nil
	} else {
		return nil, fmt.Errorf("Error deserializing DeleteSecretsResponse from JSON: " + err.Error())
	}
}

type DeleteFolderResponse struct {
	FolderUid    string `json:"folderUid"`
	ResponseCode string `json:"responseCode"`
	ErrorMessage string `json:"errorMessage"`
}

type DeleteFoldersResponse struct {
	Folders []DeleteFolderResponse `json:"folders"`
}

func DeleteFoldersResponseFromJson(jsonData string) (*DeleteFoldersResponse, error) {
	bytes := []byte(jsonData)
	res := DeleteFoldersResponse{}

	if err := json.Unmarshal(bytes, &res); err == nil {
		return &res, nil
	} else {
		return nil, fmt.Errorf("Error deserializing DeleteFoldersResponse from JSON: " + err.Error())
	}
}
