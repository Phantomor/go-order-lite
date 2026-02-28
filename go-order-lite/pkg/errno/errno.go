package errno

type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

var (
	OK               = &Error{Code: 0, Msg: "ok"}
	InvalidParam     = &Error{Code: 1001, Msg: "invalid params"}
	Unauthorized     = &Error{Code: 1002, Msg: "unauthorized"}
	InvalidAmount    = &Error{Code: 1003, Msg: "invalid amount"}
	LoginInvalid     = &Error{Code: 1004, Msg: "login invalid"}
	PermissionDenied = &Error{Code: 1005, Msg: "permission denied"}
	UserNotFound     = &Error{Code: 2001, Msg: "user not found"}
	UserExists       = &Error{Code: 2002, Msg: "user exists"}
	PasswordWrong    = &Error{Code: 2003, Msg: "password wrong"}
	InvalidUser      = &Error{Code: 2004, Msg: "invalid user"}
	MissingRequestId = &Error{Code: 2005, Msg: "missing request id"}
	OrderNotFound    = &Error{Code: 3001, Msg: "order not found"}
	DuplicateOrder   = &Error{Code: 3002, Msg: "duplicate order"}
	InternalError    = &Error{Code: 5000, Msg: "internal error"}
)
