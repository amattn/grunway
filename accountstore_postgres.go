package grunway

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/amattn/deeperror"
	_ "github.com/lib/pq"
)

type PostgresAccountStore struct {
	DB *sql.DB

	host string
	port uint16

	createAccountStmt *sql.Stmt
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

	// store.createAccountStmt, err = db.Prepare("INSERT INTO accounts(email, passhash) VALUES($1, $2)")
	store.createAccountStmt, err = db.Prepare("INSERT INTO accounts(email, passhash) VALUES($1, $2) RETURNING pkey, name, email, secretkey, created, modified")
	if err != nil {
		log.Fatal(err)
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

func (store *PostgresAccountStore) CreateAccount(email, password string) (*Account, error) {

	passhash, err := encryptPassword(password)
	if err != nil {
		derr := deeperror.New(3568377720, InternalServerErrorPrefix, err)
		derr.DebugMsg = "encryptPassword failure"
		return nil, derr
	}

	acctPtr := new(Account)

	rowPtr := store.createAccountStmt.QueryRow(email, passhash)
	if rowPtr == nil {
		derr := deeperror.New(3568377721, InternalServerErrorPrefix, err)
		derr.DebugMsg = fmt.Sprintln("createAccountStmt.Exec QueryRow", "email:", email)
		return nil, derr
	}

	err = rowPtr.Scan(
		&(acctPtr.Pkey),
		&(acctPtr.Name),
		&(acctPtr.Email),
		&(acctPtr.SecretKey),
		&(acctPtr.Created),
		&(acctPtr.Modified),
	)
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
func (store *PostgresAccountStore) AccountWithEmail(q string) (*Account, error) {
	return nil, deeperror.NewTODOError(3300996839)

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
	return nil, deeperror.NewTODOError(3300996833)

}
