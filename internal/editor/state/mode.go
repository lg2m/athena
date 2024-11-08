package state

type Mode uint8

const (
	Normal Mode = iota
	Insert
)

func (m Mode) String() string {
	switch m {
	case Normal:
		return "NOR"
	case Insert:
		return "INS"
	default:
		return "UNK"
	}
}
