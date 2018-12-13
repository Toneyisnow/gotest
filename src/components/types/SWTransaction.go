package types


type SWTransaction struct {

	SubjectAccount *SWAccount
	ObjectAccount *SWAccount
	Amount int64

}

func CreateTransaction(sAcc *SWAccount, oAcc *SWAccount, amount int64) *SWTransaction {

	transaction := new(SWTransaction)
	transaction.SubjectAccount = sAcc
	transaction.ObjectAccount = oAcc

	return transaction
}
