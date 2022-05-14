package enpass

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// sqlcipher is necessary for sqlite crypto support
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// filename of the sqlite vault file
	vaultFileName = "vault.enpassdb"
	// contains info about your vault
	vaultInfoFileName = "vault.json"
)

// Vault : vault is the container object for vault-related operations
type Vault struct {
	// Logger : the logger instance
	logger logrus.Logger

	// vault.enpassdb : SQLCipher database
	databaseFilename string

	// vault.json
	vaultInfoFilename string

	// <uuid>.enpassattach : SQLCipher database files for attachments >1KB
	//attachments []string

	// pointer to our opened database
	db *sql.DB

	// vault.json : contains info about your vault for synchronizing
	vaultInfo VaultInfo
}

type VaultCredentials struct {
	KeyfilePath string
	Password    string
	DBKey       []byte
}

func (credentials *VaultCredentials) IsComplete() bool {
	return credentials.Password != "" || credentials.DBKey != nil
}

// NewVault : Create new instance of vault and load vault info
func NewVault(vaultPath string, logLevel logrus.Level) (*Vault, error) {
	v := Vault{logger: *logrus.New()}
	v.logger.SetLevel(logLevel)

	if vaultPath == "" {
		return nil, errors.New("empty vault path provided")
	}

	vaultPath, _ = filepath.EvalSymlinks(vaultPath)
	v.databaseFilename = filepath.Join(vaultPath, vaultFileName)
	v.vaultInfoFilename = filepath.Join(vaultPath, vaultInfoFileName)
	v.logger.Debug("checking provided vault paths")
	if err := v.checkPaths(); err != nil {
		return nil, err
	}

	v.logger.Debug("loading vault info")
	var err error
	v.vaultInfo, err = v.loadVaultInfo()
	if err != nil {
		return nil, errors.Wrap(err, "could not load vault info")
	}

	v.logger.
		WithField("db_path", vaultFileName).
		WithField("info_path", vaultInfoFileName).
		Debug("initialized paths")

	return &v, nil
}

func (v *Vault) openEncryptedDatabase(path string, dbKey []byte) (err error) {
	// The raw key for the sqlcipher database is given
	// by the first 64 characters of the hex-encoded key
	dbName := fmt.Sprintf(
		"%s?_pragma_key=x'%s'&_pragma_cipher_compatibility=3",
		path,
		hex.EncodeToString(dbKey)[:masterKeyLength],
	)

	v.db, err = sql.Open("sqlite3", dbName)
	if err != nil {
		return errors.Wrap(err, "could not open database")
	}

	return nil
}

func (v *Vault) checkPaths() error {
	if _, err := os.Stat(v.databaseFilename); os.IsNotExist(err) {
		return errors.New("vault does not exist: " + v.databaseFilename)
	}

	if _, err := os.Stat(v.vaultInfoFilename); os.IsNotExist(err) {
		return errors.New("vault info file does not exist: " + v.vaultInfoFilename)
	}

	return nil
}

func (v *Vault) generateAndSetDBKey(credentials *VaultCredentials) error {
	if credentials.DBKey != nil {
		v.logger.Debug("skipping database key generation, already set")
		return nil
	}

	if credentials.Password == "" {
		return errors.New("empty vault password provided")
	}

	if credentials.KeyfilePath == "" && v.vaultInfo.HasKeyfile == 1 {
		return errors.New("you should specify a keyfile")
	} else if credentials.KeyfilePath != "" && v.vaultInfo.HasKeyfile == 0 {
		return errors.New("you are specifying an unnecessary keyfile")
	}

	v.logger.Debug("generating master password")
	masterPassword, err := v.generateMasterPassword([]byte(credentials.Password), credentials.KeyfilePath)
	if err != nil {
		return errors.Wrap(err, "could not generate vault unlock key")
	}

	v.logger.Debug("extracting salt from database")
	keySalt, err := v.extractSalt()
	if err != nil {
		return errors.Wrap(err, "could not get master password salt")
	}

	v.logger.Debug("deriving decryption key")
	credentials.DBKey, err = v.deriveKey(masterPassword, keySalt)
	if err != nil {
		return errors.Wrap(err, "could not derive database key from master password")
	}

	return nil
}

// Open : setup a connection to the Enpass database. Call this before doing anything.
func (v *Vault) Open(credentials *VaultCredentials) error {
	v.logger.Debug("generating database key")
	if err := v.generateAndSetDBKey(credentials); err != nil {
		return errors.Wrap(err, "could not generate database key")
	}

	v.logger.Debug("opening encrypted database")
	if err := v.openEncryptedDatabase(v.databaseFilename, credentials.DBKey); err != nil {
		return errors.Wrap(err, "could not open encrypted database")
	}

	var tableName string
	err := v.db.QueryRow(`
		SELECT name
		FROM sqlite_master
		WHERE type='table' AND name='item'
	`).Scan(&tableName)
	if err != nil {
		return errors.Wrap(err, "could not connect to database")
	} else if tableName != "item" {
		return errors.New("could not connect to database")
	}

	return nil
}

// Close : close the connection to the underlying database. Always call this in the end.
func (v *Vault) Close() {
	if v.db != nil {
		err := v.db.Close()
		v.logger.WithError(err).Debug("closed vault")
	}
}

// GetEntries : return the password entries in the Enpass database.
func (v *Vault) GetEntries(cardType string, filters []string) ([]Card, error) {
	if v.db == nil || v.vaultInfo.VaultName == "" {
		return nil, errors.New("vault is not initialized")
	}

	rows, err := v.db.Query(`
		SELECT uuid, type, created_at, field_updated_at, title,
		       subtitle, note, trashed, item.deleted, category,
		       label, value, key, last_used, sensitive, item.icon
		FROM item
		INNER JOIN itemfield ON uuid = item_uuid
	`)

	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve cards from database")
	}

	var cards []Card

	for rows.Next() {
		var card Card

		// read the database columns into Card object
		if err := rows.Scan(
			&card.UUID, &card.Type, &card.CreatedAt, &card.UpdatedAt, &card.Title,
			&card.Subtitle, &card.Note, &card.Trashed, &card.Deleted, &card.Category,
			&card.Label, &card.value, &card.itemKey, &card.LastUsed, &card.Sensitive, &card.Icon,
		); err != nil {
			return nil, errors.Wrap(err, "could not read card from database")
		}

		card.RawValue = card.value

		// if item has been deleted
		if card.IsDeleted() {
			continue
		}

		// if we specify a type filter
		if cardType != "" && card.Type != cardType {
			continue
		}

		// check any supplied title filters
		if len(filters) > 0 {
			found := false

			for _, filter := range filters {
				if strings.Contains(strings.ToLower(card.Title), strings.ToLower(filter)) {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		cards = append(cards, card)
	}

	return cards, nil
}

func (v *Vault) GetEntry(cardType string, filters []string, unique bool) (*Card, error) {
	cards, err := v.GetEntries(cardType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve cards")
	}

	var ret *Card
	for _, card := range cards {
		if card.IsTrashed() || card.IsDeleted() {
			continue
		} else if ret == nil {
			ret = &card
		} else if unique {
			return nil, errors.New("multiple cards match that title")
		} else {
			break
		}
	}

	if ret == nil {
		return nil, errors.New("card not found")
	}

	return ret, nil
}
