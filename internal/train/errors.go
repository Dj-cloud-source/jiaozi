package train

import "errors"

var (
	ErrInvalidTrain       = errors.New("invalid train")
	ErrTrainAlreadyExists = errors.New("train already exists")
)
