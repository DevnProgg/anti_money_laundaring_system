package models

// Account represents a customer account.
type Account struct {
	AccountID   string `db:"account_id"`
	HolderName  string `db:"holder_name"`
	Address     string `db:"address"`
	DateOfBirth string `db:"date_of_birth"`
}
