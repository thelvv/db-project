package entity

type customError string

const ForumNotExistError customError = "Forum not exists"
const DataError customError = "Data error"
const WrongParentError customError = "Wrong parent passed"
const UserDoesntExistsError customError = "User does not exist"

func (err customError) Error() string { // customError implements error interface
	return string(err)
}
