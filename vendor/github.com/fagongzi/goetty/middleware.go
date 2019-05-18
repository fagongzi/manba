package goetty

// Middleware goetty middleware
type Middleware interface {
	PreRead(conn IOSession) (bool, interface{}, error)
	PostRead(msg interface{}, conn IOSession) (bool, interface{}, error)
	PreWrite(msg interface{}, conn IOSession) (bool, interface{}, error)
	PostWrite(msg interface{}, conn IOSession) (bool, error)
	Closed(conn IOSession)
	Connected(conn IOSession)
	WriteError(err error, conn IOSession)
	ReadError(err error, conn IOSession) error
}

// BaseMiddleware defined default reutrn value
type BaseMiddleware struct {
}

// PostWrite default reutrn value
func (sm *BaseMiddleware) PostWrite(msg interface{}, conn IOSession) (bool, error) {
	return true, nil
}

// PreWrite default reutrn value
func (sm *BaseMiddleware) PreWrite(msg interface{}, conn IOSession) (bool, interface{}, error) {
	return true, msg, nil
}

// PreRead default reutrn value
func (sm *BaseMiddleware) PreRead(conn IOSession) (bool, interface{}, error) {
	return true, nil, nil
}

// PostRead default reutrn value
func (sm *BaseMiddleware) PostRead(msg interface{}, conn IOSession) (bool, interface{}, error) {
	return false, true, nil
}

// Closed default option
func (sm *BaseMiddleware) Closed(conn IOSession) {

}

// Connected default option
func (sm *BaseMiddleware) Connected(conn IOSession) {

}

// WriteError conn write err
func (sm *BaseMiddleware) WriteError(err error, conn IOSession) {
}

// ReadError conn read err
func (sm *BaseMiddleware) ReadError(err error, conn IOSession) error {
	return err
}
