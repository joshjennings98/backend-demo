package server

type slideType = int

const (
	slideTypePlain slideType = iota
	slideTypeCodeblock
	slideTypeCommand
)

type slide struct {
	id        int
	content   string
	slideType slideType
}
