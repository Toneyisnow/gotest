package types

type SWAccount struct {

	AccountId string
	Address string
}

func CreateAccount(accId string, address string) *SWAccount {

	acc := new(SWAccount)
	acc.AccountId = accId
	acc.Address = address

	return acc
}