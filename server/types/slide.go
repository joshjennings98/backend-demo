package types

type SlideType = int

const (
	SlideTypePlain SlideType = iota
	SlideTypeCodeblock
	SlideTypeCommand
	SlideTypeIframe
)

type Slide struct {
	ID        int
	Content   string
	SlideType SlideType
}
