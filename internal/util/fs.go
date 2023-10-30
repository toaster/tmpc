package util

// Filesystem permission constants
const (
	PermOtherExec = 1 << iota
	PermOtherWrite
	PermOtherRead
	PermGroupExec
	PermGroupWrite
	PermGroupRead
	PermUserExec
	PermUserWrite
	PermUserRead
)
