package models

import (
	"fmt"

	"github.com/jinzhu/gorm"

	check "gopkg.in/check.v1"
)

func (s *ModelsSuite) TestPostSMTP(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "foo@example.com",
		UserId:      1,
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, nil)
	ss, err := GetSMTPs(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ss), check.Equals, 1)
}

func (s *ModelsSuite) TestPostSMTPNoHost(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		FromAddress: "foo@example.com",
		UserId:      1,
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrHostNotSpecified)
}

func (s *ModelsSuite) TestPostSMTPNoFrom(c *check.C) {
	smtp := SMTP{
		Name:   "Test SMTP",
		UserId: 1,
		Host:   "1.1.1.1:25",
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrFromAddressNotSpecified)
}

func (s *ModelsSuite) TestPostValidFromWithName(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "Foo Bar <foo@example.com>",
		UserId:      1,
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.IsNil)
}

func (s *ModelsSuite) TestPostInvalidFromEmail(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "example.com",
		UserId:      1,
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrInvalidFromAddress)
}

func (s *ModelsSuite) TestPostSMTPValidHeader(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "foo@example.com",
		UserId:      1,
		Headers: []Header{
			Header{Key: "Reply-To", Value: "test@example.com"},
			Header{Key: "X-Mailer", Value: "gophish"},
		},
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, nil)
	ss, err := GetSMTPs(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ss), check.Equals, 1)
}

func (s *ModelsSuite) TestSMTPGetDialer(ch *check.C) {
	host := "localhost"
	port := 25
	smtp := SMTP{
		Host:             fmt.Sprintf("%s:%d", host, port),
		IgnoreCertErrors: false,
	}
	d, err := smtp.GetDialer()
	ch.Assert(err, check.Equals, nil)

	dialer := d.(*Dialer).Dialer
	ch.Assert(dialer.Host, check.Equals, host)
	ch.Assert(dialer.Port, check.Equals, port)
	ch.Assert(dialer.TLSConfig.ServerName, check.Equals, host)
	ch.Assert(dialer.TLSConfig.InsecureSkipVerify, check.Equals, smtp.IgnoreCertErrors)
}

func (s *ModelsSuite) TestGetInvalidSMTP(ch *check.C) {
	_, err := GetSMTP(-1, 1)
	ch.Assert(err, check.Equals, gorm.ErrRecordNotFound)
}

func (s *ModelsSuite) TestDefaultDeniedDial(ch *check.C) {
	host := "169.254.169.254"
	port := 25
	smtp := SMTP{
		Host: fmt.Sprintf("%s:%d", host, port),
	}
	d, err := smtp.GetDialer()
	ch.Assert(err, check.Equals, nil)
	_, err = d.Dial()
	ch.Assert(err, check.ErrorMatches, ".*upstream connection denied.*")
}

func (s *ModelsSuite) TestDKIMValidationMissingDomain(c *check.C) {
	smtp := SMTP{
		Name:           "Test SMTP",
		Host:           "1.1.1.1:25",
		FromAddress:    "foo@example.com",
		UserId:         1,
		DKIMEnabled:    true,
		DKIMSelector:   "mail",
		DKIMPrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----",
	}
	err := smtp.Validate()
	c.Assert(err, check.NotNil)
}

func (s *ModelsSuite) TestDKIMValidationMissingSelector(c *check.C) {
	smtp := SMTP{
		Name:           "Test SMTP",
		Host:           "1.1.1.1:25",
		FromAddress:    "foo@example.com",
		UserId:         1,
		DKIMEnabled:    true,
		DKIMDomain:     "example.com",
		DKIMPrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----",
	}
	err := smtp.Validate()
	c.Assert(err, check.NotNil)
}

func (s *ModelsSuite) TestDKIMValidationMissingPrivateKey(c *check.C) {
	smtp := SMTP{
		Name:         "Test SMTP",
		Host:         "1.1.1.1:25",
		FromAddress:  "foo@example.com",
		UserId:       1,
		DKIMEnabled:  true,
		DKIMDomain:   "example.com",
		DKIMSelector: "mail",
	}
	err := smtp.Validate()
	c.Assert(err, check.NotNil)
}

func (s *ModelsSuite) TestDKIMValidationInvalidPrivateKey(c *check.C) {
	smtp := SMTP{
		Name:           "Test SMTP",
		Host:           "1.1.1.1:25",
		FromAddress:    "foo@example.com",
		UserId:         1,
		DKIMEnabled:    true,
		DKIMDomain:     "example.com",
		DKIMSelector:   "mail",
		DKIMPrivateKey: "not-a-valid-pem-key",
	}
	err := smtp.Validate()
	c.Assert(err, check.NotNil)
	c.Assert(err, check.Equals, ErrInvalidDKIMKey)
}

func (s *ModelsSuite) TestDKIMValidationComplete(c *check.C) {
	// Valid PKCS#1 RSA private key (minimal test key - not for production use)
	validPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA2mKqHD5mXZSsGB2m7wZvbGRvDMzV6k5yjMA8RgvMbAFEqEjH
NfvGRvMxWOzJqBP1aNGK0SYk0MZsqaSXBqf0M2Q8P8xHqSTRNF9q4kHq6mRNsq4H
DqhAJKqfVMJBkqGZq5T8OxsvMNFqF0j8q0VNT5qF9V1aT0F9V1qF9V1aT0F9V1qF
9V1aT0F9V1qF9V1aT0F9V1qF9V1aT0F9V1qF9V1aT0F9V1qF9V1aT0F9V1qF9V1a
T0F9V1qF9V1aT0F9V1qF9V1aT0F9V1qF9V1aT0F9V1qF9V1aT0F9V1qF9V1aT0F9
V1qF9V1aT0F9V1qF9V1aT0F9V1qF9V1aT0F9V1qF9wIDAQABAoIBAC8q9k1L8k0K
5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr
5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5V
qB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5t
Jr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E
5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB
5tJr5E5VqBECgYEA7k+E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB
5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr
5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5wKBgQDq
R5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJ
r5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5
VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqBEKBgQDqR5tJ
r5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5
VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5
tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5wKBgQCqR5tJr5E5VqB5tJr5E
5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB
5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr
5E5VqB5tJr5E5VqB5tJr5E5VqBQKBgQDqR5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB
5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr
5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5VqB5tJr5E5V
qB5tJr5E5VqB5tJr
-----END RSA PRIVATE KEY-----`
	smtp := SMTP{
		Name:           "Test SMTP with DKIM",
		Host:           "1.1.1.1:25",
		FromAddress:    "foo@example.com",
		UserId:         1,
		DKIMEnabled:    true,
		DKIMDomain:     "example.com",
		DKIMSelector:   "mail",
		DKIMPrivateKey: validPrivateKey,
	}
	err := smtp.Validate()
	c.Assert(err, check.IsNil)
}

func (s *ModelsSuite) TestDKIMDisabledNoValidation(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "foo@example.com",
		UserId:      1,
		DKIMEnabled: false,
	}
	err := smtp.Validate()
	c.Assert(err, check.IsNil)
}

func (s *ModelsSuite) TestHelloHostnameConfiguration(c *check.C) {
	smtp := SMTP{
		Name:          "Test SMTP",
		Host:          "1.1.1.1:25",
		FromAddress:   "foo@example.com",
		UserId:        1,
		HelloHostname: "mail.example.com",
	}
	d, err := smtp.GetDialer()
	c.Assert(err, check.Equals, nil)

	dialer := d.(*Dialer).Dialer
	c.Assert(dialer.LocalName, check.Equals, "mail.example.com")
}

func (s *ModelsSuite) TestHelloHostnameDefaultWhenNotSet(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "foo@example.com",
		UserId:      1,
	}
	d, err := smtp.GetDialer()
	c.Assert(err, check.Equals, nil)

	dialer := d.(*Dialer).Dialer
	c.Assert(dialer.LocalName, check.Not(check.Equals), "localhost")
}
