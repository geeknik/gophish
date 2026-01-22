package models

import (
	"errors"
	"net"
	"time"

	"github.com/gophish/gophish/encryption"
	log "github.com/gophish/gophish/logger"
)

const DefaultIMAPFolder = "INBOX"
const DefaultIMAPFreq = 60 // Every 60 seconds

// IMAP contains the attributes needed to handle logging into an IMAP server to check
// for reported emails
type IMAP struct {
	UserId                      int64     `json:"-" gorm:"column:user_id"`
	Enabled                     bool      `json:"enabled"`
	Host                        string    `json:"host"`
	Port                        uint16    `json:"port,string,omitempty"`
	Username                    string    `json:"username"`
	Password                    string    `json:"password"`
	TLS                         bool      `json:"tls"`
	IgnoreCertErrors            bool      `json:"ignore_cert_errors"`
	Folder                      string    `json:"folder"`
	RestrictDomain              string    `json:"restrict_domain"`
	DeleteReportedCampaignEmail bool      `json:"delete_reported_campaign_email"`
	LastLogin                   time.Time `json:"last_login,omitempty"`
	ModifiedDate                time.Time `json:"modified_date"`
	IMAPFreq                    uint32    `json:"imap_freq,string,omitempty"`
	// OAuth 2.0 fields for Microsoft 365
	UseOAuth2     bool   `json:"use_oauth2" gorm:"column:use_oauth2"`
	OAuthTenantID string `json:"oauth_tenant_id" gorm:"column:oauth_tenant_id"`
	OAuthClientID string `json:"oauth_client_id" gorm:"column:oauth_client_id"`
	OAuthSecret   string `json:"oauth_client_secret" gorm:"column:oauth_client_secret"`
}

// ErrIMAPHostNotSpecified is thrown when there is no Host specified
// in the IMAP configuration
var ErrIMAPHostNotSpecified = errors.New("No IMAP Host specified")

// ErrIMAPPortNotSpecified is thrown when there is no Port specified
// in the IMAP configuration
var ErrIMAPPortNotSpecified = errors.New("No IMAP Port specified")

// ErrInvalidIMAPHost indicates that the IMAP server string is invalid
var ErrInvalidIMAPHost = errors.New("Invalid IMAP server address")

// ErrInvalidIMAPPort indicates that the IMAP Port is invalid
var ErrInvalidIMAPPort = errors.New("Invalid IMAP Port")

// ErrIMAPUsernameNotSpecified is thrown when there is no Username specified
// in the IMAP configuration
var ErrIMAPUsernameNotSpecified = errors.New("No Username specified")

// ErrIMAPPasswordNotSpecified is thrown when there is no Password specified
// in the IMAP configuration
var ErrIMAPPasswordNotSpecified = errors.New("No Password specified")

// ErrInvalidIMAPFreq is thrown when the frequency for polling the
// IMAP server is invalid
var ErrInvalidIMAPFreq = errors.New("Invalid polling frequency")

// TableName specifies the database tablename for Gorm to use
func (im IMAP) TableName() string {
	return "imap"
}

func (im *IMAP) BeforeSave() error {
	key := GetEncryptionKey()
	if key != nil {
		if im.Password != "" {
			encrypted, err := encryption.Encrypt(key, im.Password)
			if err != nil {
				log.Error("Failed to encrypt IMAP password: ", err)
				return err
			}
			im.Password = encrypted
		}
		if im.OAuthSecret != "" {
			encrypted, err := encryption.Encrypt(key, im.OAuthSecret)
			if err != nil {
				log.Error("Failed to encrypt OAuth secret: ", err)
				return err
			}
			im.OAuthSecret = encrypted
		}
	}
	return nil
}

func (im *IMAP) AfterFind() error {
	key := GetEncryptionKey()
	if key != nil {
		if im.Password != "" {
			decrypted, err := encryption.Decrypt(key, im.Password)
			if err != nil {
				log.Error("Failed to decrypt IMAP password: ", err)
				return err
			}
			im.Password = decrypted
		}
		if im.OAuthSecret != "" {
			decrypted, err := encryption.Decrypt(key, im.OAuthSecret)
			if err != nil {
				log.Error("Failed to decrypt OAuth secret: ", err)
				return err
			}
			im.OAuthSecret = decrypted
		}
	}
	return nil
}

// ErrOAuthConfigIncomplete is thrown when OAuth 2.0 is enabled but configuration is incomplete
var ErrOAuthConfigIncomplete = errors.New("OAuth 2.0 configuration incomplete: tenant_id, client_id, and client_secret are required")

// Validate ensures that IMAP configs/connections are valid
func (im *IMAP) Validate() error {
	if im.Host == "" {
		return ErrIMAPHostNotSpecified
	}
	if im.Port == 0 {
		return ErrIMAPPortNotSpecified
	}
	if im.Username == "" {
		return ErrIMAPUsernameNotSpecified
	}

	if im.UseOAuth2 {
		if im.OAuthTenantID == "" || im.OAuthClientID == "" || im.OAuthSecret == "" {
			return ErrOAuthConfigIncomplete
		}
	} else {
		if im.Password == "" {
			return ErrIMAPPasswordNotSpecified
		}
	}

	// Set the default value for Folder
	if im.Folder == "" {
		im.Folder = DefaultIMAPFolder
	}

	// Make sure im.Host is an IP or hostname. NB will fail if unable to resolve the hostname.
	ip := net.ParseIP(im.Host)
	_, err := net.LookupHost(im.Host)
	if ip == nil && err != nil {
		return ErrInvalidIMAPHost
	}

	// Make sure 1 >= port <= 65535
	if im.Port < 1 || im.Port > 65535 {
		return ErrInvalidIMAPPort
	}

	// Make sure the polling frequency is between every 30 seconds and every year
	// If not set it to the default
	if im.IMAPFreq < 30 || im.IMAPFreq > 31540000 {
		im.IMAPFreq = DefaultIMAPFreq
	}

	return nil
}

// GetIMAP returns the IMAP server owned by the given user.
func GetIMAP(uid int64) ([]IMAP, error) {
	im := []IMAP{}
	count := 0
	err := db.Where("user_id=?", uid).Find(&im).Count(&count).Error

	if err != nil {
		log.Error(err)
		return im, err
	}
	return im, nil
}

// PostIMAP updates IMAP settings for a user in the database.
func PostIMAP(im *IMAP, uid int64) error {
	err := im.Validate()
	if err != nil {
		log.Error(err)
		return err
	}

	// Delete old entry. TODO: Save settings and if fails to Save below replace with original
	err = DeleteIMAP(uid)
	if err != nil {
		log.Error(err)
		return err
	}

	// Insert new settings into the DB
	err = db.Save(im).Error
	if err != nil {
		log.Error("Unable to save to database: ", err.Error())
	}
	return err
}

// DeleteIMAP deletes the existing IMAP in the database.
func DeleteIMAP(uid int64) error {
	err := db.Where("user_id=?", uid).Delete(&IMAP{}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

func SuccessfulLogin(im *IMAP) error {
	err := db.Model(&im).Where("user_id = ?", im.UserId).Update("last_login", time.Now().UTC()).Error
	if err != nil {
		log.Error("Unable to update database: ", err.Error())
	}
	return err
}
