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

	createAccountStmt           *sql.Stmt
	queryAccountByEmailStmt     *sql.Stmt
	queryAccountByPublicKeyStmt *sql.Stmt
	updateAccountLastLoginStmt  *sql.Stmt
}

// #     #
// #     # ###### #      #####  ###### #####   ####
// #     # #      #      #    # #      #    # #
// ####### #####  #      #    # #####  #    #  ####
// #     # #      #      #####  #      #####       #
// #     # #      #      #      #      #   #  #    #
// #     # ###### ###### #      ###### #    #  ####
//

func scanAccountRow(rowPtr *sql.Row) (*Account, *deeperror.DeepError) {
	acctPtr := new(Account)
	err := rowPtr.Scan(
		&(acctPtr.PKey),
		&(acctPtr.Name),
		&(acctPtr.Email),
		&(acctPtr.Passhash),
		&(acctPtr.PublicKey),
		&(acctPtr.SecretKey),
		&(acctPtr.LastLogin),
		&(acctPtr.Created),
		&(acctPtr.Modified),
	)

	if err != nil {
		if err == sql.ErrNoRows {
			derr := deeperror.New(3044022520, NotFoundPrefix, err)
			derr.DebugMsg = "No rows returned"
			return nil, nil
		} else {
			derr := deeperror.New(3044022521, InternalServerErrorPrefix, err)
			derr.DebugMsg = "Scan failure"
			return nil, derr
		}
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

	fieldsClause := "pkey, name, email, passhash, publickey, secretkey, lastlogin, created, modified"

	queries := map[string]**sql.Stmt{
		"INSERT INTO accounts(name, email, passhash) VALUES($1, $2, $3) RETURNING " + fieldsClause: &store.createAccountStmt,
		"SELECT " + fieldsClause + " FROM accounts WHERE email = $1":                               &store.queryAccountByEmailStmt,
		"SELECT " + fieldsClause + " FROM accounts WHERE publickey = $1":                           &store.queryAccountByPublicKeyStmt,
		"UPDATE accounts SET lastlogin = now() WHERE pkey = $1 RETURNING " + fieldsClause:          &store.updateAccountLastLoginStmt,
	}

	return PrepareStatements(store.DB, queries)
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

	acctPtr, derr := scanAccountRow(rowPtr)
	if derr != nil {
		innerDerr := deeperror.NewHTTPError(3568377723, derr.EndUserMsg, derr, derr.StatusCode)
		innerDerr.DebugMsg = "Scan returned error"
		return nil, innerDerr
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
		return nil, nil
	}

	acctPtr, err := scanAccountRow(rowPtr)
	if err != nil {
		derr := deeperror.New(3134354280, InternalServerErrorPrefix, err)
		derr.DebugMsg = "Scan failure"
		return nil, derr
	}
	return acctPtr, nil
}
func (store *PostgresAccountStore) AccountWithPublicKey(email string) (*Account, error) {
	rowPtr := store.queryAccountByPublicKeyStmt.QueryRow(email)
	if rowPtr == nil {
		return nil, nil
	}

	acctPtr, err := scanAccountRow(rowPtr)
	if err != nil {
		derr := deeperror.New(3134354280, InternalServerErrorPrefix, err)
		derr.DebugMsg = "Scan failure"
		return nil, derr
	}
	return acctPtr, nil
}

func (store *PostgresAccountStore) EmailAddressAvailable(email string) (bool, *deeperror.DeepError) {

	acct, err := store.AccountWithEmail(email)

	if err != nil {
		derr := deeperror.NewHTTPError(3237725249, "Could not query email address", err, http.StatusInternalServerError)
		return false, derr
	} else {
		return (acct == nil), nil
	}
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
func (store *PostgresAccountStore) UpdateUserLastLogin(pkey int64) (*Account, error) {

	rowPtr := store.updateAccountLastLoginStmt.QueryRow(pkey)
	if rowPtr == nil {
		derr := deeperror.NewHTTPError(3809025802, "pkey Not Found", nil, http.StatusNotFound)
		return nil, derr
	}

	acctPtr, err := scanAccountRow(rowPtr)
	if err != nil {
		derr := deeperror.New(3809025803, InternalServerErrorPrefix, err)
		derr.DebugMsg = "Scan failure"
		return nil, derr
	}

	return acctPtr, nil
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
	acct, err := store.AccountWithEmail(submittedEmail)
	if err != nil {
		return nil, deeperror.New(3770650641, "Auth Failure", err)
	}

	err = bcrypt.CompareHashAndPassword(acct.Passhash, []byte(submittedPassword))
	if err != nil {
		return nil, deeperror.New(3770650642, "Auth Failure", err)
	}

	acct, err = store.UpdateUserLastLogin(acct.PKey)
	if err != nil {
		return nil, deeperror.NewHTTPError(3770650643, "Auth Failure", err, http.StatusInternalServerError)
	}

	return acct, nil
}
