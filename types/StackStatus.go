package types

type StackStatus int

const (
	UP StackStatus = iota
	DOWN
)

func (s StackStatus) String() string {
	switch s {
	case UP:
		return "UP"
	case DOWN:
		return "DOWN"
	default:
		panic("Unknown StackStatus")
	}
}
