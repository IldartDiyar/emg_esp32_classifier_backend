package cerrors

import "errors"

var ErrNotFound = errors.New("not found")
var ErrDeviceBusy = errors.New("device busy")
var ErrIncorrectRep = errors.New("incorrect rep")
var ErrMovementNotAllowed = errors.New("movement not allowed")
var ErrSomethingWentWrong = errors.New("something went wrong")
