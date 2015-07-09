package gominesweeper

import (
	"errors"
	"math/rand"
	"time"
)

const (
	Mine    = -1
	Flagged = -2
	Checked = -3
	Unknown = -4
)

var (
	ErrExceedDimensions = errors.New("size exceeds max dimensions")
	ErrOutOfBounds      = errors.New("point is out of bounds")
	ErrBadCount         = errors.New("points not equal to specification")
	ErrDupPoint         = errors.New("duplicate point found")
)

// Position represents an point on the X,Y axis
type Position struct {
	X, Y int
}

// Selector is a custom mine selector that given a width, height, and max
// will return a set of positions for placing mines.
type Selector func(width, height, max uint) ([]Position, error)

// RandomSelector is a random mine selector.
func RandomSelector(width, height, max uint) ([]Position, error) {
	size := width * height
	if size <= max {
		return nil, ErrExceedDimensions
	}
	rand.Seed(time.Now().UnixNano())
	scope := make([]uint, size)
	for i := range scope {
		scope[i] = uint(i)
		j := rand.Intn(i + 1)
		scope[i], scope[j] = scope[j], scope[i]
	}
	points := make([]Position, max)
	for i := range points {
		points[i] = Position{int(scope[i] % width), int(scope[i] / width)}
	}
	return points, nil
}

// Block represents a single unit of space that will provide information of the
// number of mines within its proximity.
type Block struct {
	proximity int
	flagged   bool
	checked   bool
}

// NewBlock instantiates a new Block.
func NewBlock(proximity int) *Block {
	return &Block{proximity, false, false}
}

// Check will verify the status of a block while only revealing its proximity
// if the block is selected.
func (b *Block) Check() int {
	if b.flagged {
		return Flagged
	} else if b.checked {
		return b.proximity
	}
	return Unknown
}

// Select reveals the proximity of the block if not already revealed and not
// previously flagged.
func (b *Block) Select() int {
	if b.flagged {
		return Flagged
	} else if b.checked {
		return Checked
	}
	b.checked = true
	return b.proximity
}

// ToggleFlag toggles the flag indicator on the block.
func (b *Block) ToggleFlag() {
	if !b.checked {
		b.flagged = !b.flagged
	}
}

// Minefield describes the layout of all the blocks.
type Minefield map[Position]*Block

// NewMinefield generates a new minefield using the random mine selector
func NewMinefield(width, height, mines uint) (Minefield, error) {
	return Minefield(make(map[Position]*Block)).init(width, height, mines, RandomSelector)
}

// init initializes the minefield.
func (mf Minefield) init(width, height, mines uint, selector Selector) (Minefield, error) {
	minefield, err := selector(width, height, mines)
	if err != nil {
		return nil, err
	} else if len(minefield) != int(mines) {
		return nil, ErrBadCount
	}

	// set the mines on the map
	for _, mine := range minefield {
		// make sure we don't have bogus mines
		if mine.X < 0 || mine.X >= int(width) || mine.Y < 0 || mine.Y >= int(height) {
			return nil, ErrOutOfBounds
		} else if block := mf[mine]; block != nil && block.proximity == Mine {
			return nil, ErrDupPoint
		}

		mf[mine] = NewBlock(Mine)
		for deltaX := -1; deltaX <= 1; deltaX++ {
			if x := mine.X + deltaX; x >= 0 && x < int(width) {
				for deltaY := -1; deltaY <= 1; deltaY++ {
					if deltaX == 0 && deltaY == 0 {
						continue
					} else if y := mine.Y + deltaY; y >= 0 && y < int(height) {
						if block := mf[Position{x, y}]; block != nil {
							if block.proximity != Mine {
								block.proximity++
							}
						} else {
							mf[Position{x, y}] = NewBlock(1)
						}
					}
				}
			}
		}
	}

	// set the remaining blocks
	for x := 0; x < int(width); x++ {
		for y := 0; y < int(height); y++ {
			if _, ok := mf[Position{x, y}]; !ok {
				mf[Position{x, y}] = NewBlock(0)
			}
		}
	}
	return mf, nil
}

// Select will select an individual block and return the proximity to its
// neighboring mines.  If the proximity is 0, then Select will recursively
// reveal its neighbors as well.
func (mf Minefield) Select(x, y int) (int, error) {
	pos := Position{x, y}
	block, ok := mf[pos]
	if !ok {
		return 0, ErrOutOfBounds
	}

	proximity := block.Select()
	if proximity == 0 {
		for deltaX := -1; deltaX <= 1; deltaX++ {
			for deltaY := -1; deltaY <= 1; deltaY++ {
				if deltaX == 0 && deltaY == 0 {
					continue
				}
				mf.Select(x+deltaX, y+deltaY)
			}
		}
	} else if proximity == Mine {
		for _, block := range mf {
			if block.proximity == Mine {
				block.Select()
			}
		}
	}
	return proximity, nil
}

// ToggleFlag toggles the flag on a particular mine.
func (mf Minefield) ToggleFlag(x, y int) {
	if block, ok := mf[Position{x, y}]; ok {
		block.ToggleFlag()
	}
}

// Display returns the current state of all the blocks.
func (mf Minefield) Display() map[Position]int {
	display := make(map[Position]int)
	for pos, block := range mf {
		display[pos] = block.Check()
	}
	return display
}