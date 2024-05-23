package types

type SlideType = int

const (
	SlideTypePlain SlideType = iota
	SlideTypeCodeblock
	SlideTypeCommand
)

type Slide struct {
	ID        int
	Content   string
	SlideType SlideType
}
