package oshandler

import "runtime"

// Handler interface defines methods for OS-specific operations
type Handler interface {
	Handle(absPath, action string) error
	HandleAlias(aliasName, command, action string) error
}

// New returns the appropriate handler based on OS
func New() Handler {
	switch runtime.GOOS {
	case "darwin":
		return &macHandler{}
	case "linux":
		return &linuxHandler{}
	case "windows":
		return &windowsHandler{}
	default:
		return nil
	}
}
