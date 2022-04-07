package apiutils

// TODO: move this code to web3server.go after issue#3248 is solved
type (
	// JSONWriter is writer for web3 request
	JSONWriter struct {
		Write       func(interface{}) error
		IsWriteOnce bool
	}
)
