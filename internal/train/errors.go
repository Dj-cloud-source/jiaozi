package train

import "errors"

var (
	ErrInvalidTrain       = errors.New("invalid train")
	ErrTrainAlreadyExists = errors.New("train already exists")
	ErrTrainNotFound      = errors.New("train not found")
	ErrTrainStatusInvalid = errors.New("train status invalid")
)
