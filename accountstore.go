package grunway

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/amattn/deeperror"
)

type AccountStore interface {

	// Common startup, shutdown, utility methods
	Startup(attribs string) error
	Shutdown() error

	// CREATE
	// returns created account's accountid
	CreateAccount(email, password string) (*Account, error)

	// DELETE
	// returns true when id is valid account is successfully removed.  if no account with id, returns false
	// returns non-nil err on failure to execute
	DeleteAccount(id int64) (bool, error)

	// QUERY
	AllAccounts() ([]*Account, error)
	AccountWithId(id int64) (*Account, error)
	AccountWithEmail(q string) (*Account, error)
	EmailAddressAvailable(email string) (bool, error)

	// UPDATE
	ChangeUserEmail(id int64, newEmail string) error
	ChangeUserPassword(id int64, newPassword string) error

	// AUTH
	Login(submittedEmail, submittedPassword string) (*Account, error)
}

type Account struct {
	// System fields
	Pkey     int64  `meddler:"pkey,pk"`
	Passhash []byte `meddler:"passhash"`
	// Salt      []byte `meddler:"-"` // may not be necessary if using certain algos (bcrypt)
	SecretKey string `meddler:"secretkey"`

	// Times
	Created   time.Time `meddler:"tsadd,utctimez"`
	Modified  time.Time `meddler:"tsmod,utctimez"`
	LastLogin time.Time `meddler:"lastlogin,utctimez"`

	// User fields
	Email string `meddler:"email"`
	Name  string `meddler:"name"`
}

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
// UTILITIES

func encryptPassword(password string) ([]byte, error) {
	// bcrypt.DefaultCost can be substituted for any number between
	// bcrypt.MinimumCost and bcrypt.MaximumCost, inclusively.
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func generateSecretKey() (string, error) {
	c := 64
	data := make([]byte, c)
	n, err := io.ReadFull(rand.Reader, data)
	if n != len(data) || err != nil {
		innerErr := deeperror.NewHTTPError(1297539203, "Key Generation Error", err, http.StatusInternalServerError)
		return "", innerErr
	}

	return base64.URLEncoding.EncodeToString(data), nil
}

// Yes this could be better... good enough for now.
func SimpleEmailValidation(email string) bool {
	if len(email) < 5 || len(email) > 254 {
		return false
	}
	if strings.Contains(email, "@") == false {
		return false
	}
	if strings.Contains(email, ".") == false {
		return false
	}
	if strings.LastIndex(email, ".") < strings.Index(email, "@") {
		return false
	}
	if strings.LastIndex(email, ".")-1 == strings.Index(email, "@") {
		return false
	}
	if strings.Index(email, "@") == 0 {
		return false
	}
	if strings.LastIndex(email, ".") == len(email)-1 {
		return false
	}

	return true
}

func SimplePasswordValidation(pw string) bool {
	if len(pw) < 8 || len(pw) > 1024 {
		return false
	}

	if strings.ContainsAny(pw, "01234567890!@#$%%^&*()_+-=[]{}\\|;:'\"`~,.<>/?") == false {
		return false
	}

	return true
}
