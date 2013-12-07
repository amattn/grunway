package grunway

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/amattn/deeperror"
	_ "github.com/lib/pq"
)

type PostgresAccountStore struct {
	DB *sql.DB

	host string
	port uint16

	createAccountStmt       *sql.Stmt
	queryAccountByEmailStmt *sql.Stmt
}

// #     #
// #     # ###### #      #####  ###### #####   ####
// #     # #      #      #    # #      #    # #
// ####### #####  #      #    # #####  #    #  ####
// #     # #      #      #####  #      #####       #
// #     # #      #      #      #      #   #  #    #
// #     # ###### ###### #      ###### #    #  ####
//

func scanAccountRow(rowPtr *sql.Row) (*Account, error) {
	acctPtr := new(Account)
	err := rowPtr.Scan(
		&(acctPtr.Pkey),
		&(acctPtr.Name),
		&(acctPtr.Email),
		&(acctPtr.Passhash),
		&(acctPtr.PublicKey),
		&(acctPtr.SecretKey),
		&(acctPtr.Created),
		&(acctPtr.Modified),
	)

	if err != nil {
		return nil, deeperror.New(3044022520, InternalServerErrorPrefix, err)
	}
	return acctPtr, nil
}

//  #####
// #     # ###### ##### #    # #####
// #       #        #   #    # #    #
//  #####  #####    #   #    # #    #
//       # #        #   #    # #####
// #     # #        #   #    # #
//  #####  ######   #    ####  #
//

func (store *PostgresAccountStore) Startup(attribs string) error {
	db, err := sql.Open("postgres", attribs)
	if err != nil {
		return deeperror.New(1136587310, "CANNOT STARTUP "+store.StoreName(), err)
	}

	err = db.Ping()
	if err != nil {
		return deeperror.New(1136587311, "CANNOT COMMUNICATE WITH "+store.StoreName(), err)
	}
	store.DB = db
	var query string
	// store.createAccountStmt, err = db.Prepare("INSERT INTO accounts(email, passhash) VALUES($1, $2)")
	query = "INSERT INTO accounts(name, email, passhash) VALUES($1, $2, $3) RETURNING pkey, name, email, passhash, publickey, secretkey, created, modified"
	store.createAccountStmt, err = db.Prepare(query)
	if err != nil {
		return deeperror.New(1136587312, "failure to prepare query:"+query, err)
	}

	query = "SELECT pkey, name, email, passhash, publickey, secretkey, created, modified FROM accounts WHERE email = $1"
	store.queryAccountByEmailStmt, err = db.Prepare(query)
	if err != nil {
		return deeperror.New(1136587313, "failure to prepare query:"+query, err)
	}

	return nil
}

func (store *PostgresAccountStore) Shutdown() error {
	if store.DB != nil {
		store.DB.Close()
		log.Println("1238887564 Shutting down ", store.StoreName(), "at", store.CurrentHostPort())
		store.DB = nil
	}
	return nil
}

func (store *PostgresAccountStore) DBName() string {
	return "accounts"
}
func (store *PostgresAccountStore) StoreName() string {
	return "AccountStore"
}
func (store *PostgresAccountStore) CurrentHostPort() string {
	return fmt.Sprintf("%s:%d", store.host, store.port)
}

//  #####
// #     # #####  ######   ##   ##### ######
// #       #    # #       #  #    #   #
// #       #    # #####  #    #   #   #####
// #       #####  #      ######   #   #
// #     # #   #  #      #    #   #   #
//  #####  #    # ###### #    #   #   ######
//

func (store *PostgresAccountStore) CreateAccount(name, email, password string) (*Account, error) {

	passhash, err := encryptPassword(password)
	if err != nil {
		derr := deeperror.New(3568377720, InternalServerErrorPrefix, err)
		derr.DebugMsg = "encryptPassword failure"
		return nil, derr
	}

	rowPtr := store.createAccountStmt.QueryRow(name, email, passhash)
	if rowPtr == nil {
		derr := deeperror.New(3568377721, InternalServerErrorPrefix, err)
		derr.DebugMsg = fmt.Sprintln("createAccountStmt.Exec QueryRow", "email:", email)
		return nil, derr
	}

	acctPtr, err := scanAccountRow(rowPtr)
	if err != nil {
		derr := deeperror.New(3568377723, InternalServerErrorPrefix, err)
		derr.DebugMsg = "Scan failure"
		return nil, derr
	}

	return acctPtr, nil
}

// ######
// #     # ###### #      ###### ##### ######
// #     # #      #      #        #   #
// #     # #####  #      #####    #   #####
// #     # #      #      #        #   #
// #     # #      #      #        #   #
// ######  ###### ###### ######   #   ######
//
// returns true when id is valid account is successfully removed.  if no account with id, returns false
// returns non-nil err on failure to execute
func (store *PostgresAccountStore) DeleteAccount(id int64) (bool, error) {
	return false, deeperror.NewTODOError(330767743)
}

//  #####
// #     # #    # ###### #####  #   #
// #     # #    # #      #    #  # #
// #     # #    # #####  #    #   #
// #   # # #    # #      #####    #
// #    #  #    # #      #   #    #
//  #### #  ####  ###### #    #   #
//

func (store *PostgresAccountStore) AllAccounts() ([]*Account, error) {
	return []*Account{}, deeperror.NewTODOError(3300996837)

}
func (store *PostgresAccountStore) AccountWithId(id int64) (*Account, error) {
	return nil, deeperror.NewTODOError(3300996838)

}
func (store *PostgresAccountStore) AccountWithEmail(email string) (*Account, error) {
	rowPtr := store.queryAccountByEmailStmt.QueryRow(email)
	if rowPtr == nil {
		derr := deeperror.NewHTTPError(3404797117, "Email Not Found", nil, http.StatusNotFound)
		return nil, derr
	}

	acctPtr, err := scanAccountRow(rowPtr)
	if err != nil {
		derr := deeperror.New(3134354280, InternalServerErrorPrefix, err)
		derr.DebugMsg = "Scan failure"
		return nil, derr
	}
	return acctPtr, nil
}

func (store *PostgresAccountStore) EmailAddressAvailable(email string) (bool, error) {
	return false, deeperror.NewTODOError(3300996840)
}

// #     #
// #     # #####  #####    ##   ##### ######
// #     # #    # #    #  #  #    #   #
// #     # #    # #    # #    #   #   #####
// #     # #####  #    # ######   #   #
// #     # #      #    # #    #   #   #
//  #####  #      #####  #    #   #   ######
//

func (store *PostgresAccountStore) ChangeUserEmail(id int64, newEmail string) error {
	return deeperror.NewTODOError(3300996831)

}
func (store *PostgresAccountStore) ChangeUserPassword(id int64, newPassword string) error {
	return deeperror.NewTODOError(3300996832)

}

//    #
//   # #   #    # ##### #    #
//  #   #  #    #   #   #    #
// #     # #    #   #   ######
// ####### #    #   #   #    #
// #     # #    #   #   #    #
// #     #  ####    #   #    #
//

func (store *PostgresAccountStore) Login(submittedEmail, submittedPassword string) (*Account, error) {
	accountPtr, err := store.AccountWithEmail(submittedEmail)
	if err != nil {
		return nil, deeperror.New(3770650641, "Auth Failure", err)
	}

	err = bcrypt.CompareHashAndPassword(accountPtr.Passhash, []byte(submittedPassword))
	if err != nil {
		return nil, deeperror.New(3770650642, "Auth Failure", err)
	}
	return accountPtr, nil
}
