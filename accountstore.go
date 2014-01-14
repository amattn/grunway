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

const (
	MAX_EMAIL_LENGTH    = 256
	MAX_PASSWORD_LENGTH = 1024
)

type AccountStore interface {

	// Common startup, shutdown, utility methods
	Startup(attribs string) error
	Shutdown() error

	// CREATE
	// returns created account's accountid
	CreateAccount(name, email, password string) (*Account, error)

	// DELETE
	// returns true when pkey is valid account is successfully removed.  if no account with pkey, returns false
	// returns non-nil err on failure to execute
	DeleteAccount(pkey int64) (bool, error)

	// QUERY
	AllAccounts() ([]*Account, error)
	AccountWithId(pkey int64) (MaybeAccount, error)
	AccountWithEmail(q string) (MaybeAccount, error)
	AccountWithPublicKey(q string) (MaybeAccount, error)
	EmailAddressAvailable(email string) (bool, *deeperror.DeepError)

	// UPDATE
	ChangeUserEmail(pkey int64, newEmail string) error
	ChangeUserPassword(pkey int64, newPassword string) error
	UpdateUserLastLogin(pkey int64) (MaybeAccount, error)

	// AUTH
	Login(submittedEmail, submittedPassword string) (*Account, error)
}

type Account struct {
	// System fields
	PKey int64 `meddler:"pkey,pk"`

	// User fields
	Name      string `meddler:"name"`
	Email     string `meddler:"email"`
	Passhash  []byte `meddler:"passhash"`
	PublicKey string `meddler:"publickey"`
	SecretKey string `meddler:"secretkey"`

	// Times
	Created   time.Time `meddler:"created,utctimez"`
	Modified  time.Time `meddler:"modified,utctimez"`
	LastLogin time.Time `meddler:"lastlogin,utctimez"`
}

type MaybeAccount struct {
	hiddenAccount *Account
}

func MakeMaybeAccount(acct *Account) MaybeAccount {
	return MaybeAccount{acct}
}

func (ma MaybeAccount) AccountOrCrash(optionalErrNum int64) *Account {
	if ma.IsNil() {
		errNum := optionalErrNum
		if errNum == 0 {
			errNum = 2857365840
		}
		deeperror.Fatal(errNum, "Fatal Error", nil)
	}
	return ma.hiddenAccount
}

func (ma MaybeAccount) HasValidAccountPointer() bool {
	if ma.hiddenAccount == nil {
		return false
	}
	return true
}

func (ma MaybeAccount) IsNil() bool {
	return !ma.HasValidAccountPointer()
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
	if len(email) < 5 || len(email) > MAX_EMAIL_LENGTH {
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
	if len(pw) < 8 || len(pw) > MAX_PASSWORD_LENGTH {
		return false
	}

	if strings.ContainsAny(pw, "01234567890!@#$%%^&*()_+-=[]{}\\|;:'\"`~,.<>/?") == false {
		return false
	}

	return true
}
