package gominesweeper

import (
	"testing"

	. "gopkg.in/check.v1"
)

/*
* Example (5x5, 5 mines):
*
*	 -------------------
*	| * | 2 | 1 | 1 | * |
*	|---+---+---+---+---|
*	| 2 | 3 | * | 2 | 1 |
*	|---+---+---+---+---|
*	| 1 | * | 2 | 1 | 0 |
*	|---+---+---+---+---|
*	| 1 | 1 | 2 | 1 | 1 |
*	|---+---+---+---+---|
*	| 0 | 0 | 1 | * | 1 |
*	 -------------------
*
 */

// Hook up gocheck into the go test runner
func Test(t *testing.T) { TestingT(t) }

type MSSuite struct {
	selector Selector
}

var _ = Suite(&MSSuite{
	selector: RandomSelector,
})

func (s *MSSuite) BenchmarkSelector(c *C) {
}

func (s *MSSuite) TestSelector(c *C) {
	// verify dimensions
	points, err := s.selector(2, 2, 5)
	c.Check(err, Equals, ErrExceedDimensions)

	// verify output
	points, err = s.selector(3, 2, 5)
	c.Assert(err, IsNil)
	c.Assert(points, HasLen, 5)

	// verify each point is unique
	pointmap := make(map[Position]struct{})
	for _, point := range points {
		c.Logf("Checking point %+v", point)
		c.Check(point.X < 3, Equals, true)
		c.Check(point.Y < 2, Equals, true)
		if _, ok := pointmap[point]; ok {
			c.Errorf("Duplicate point for %+v", point)
		}
	}
}

func (s *MSSuite) TestBlock(c *C) {
	b := NewBlock(2)
	c.Check(b.Check(), Equals, Unknown)
	b.ToggleFlag()
	c.Check(b.Check(), Equals, Flagged)
	c.Check(b.Select(), Equals, Flagged)
	b.ToggleFlag()
	c.Check(b.Check(), Equals, Unknown)
	c.Check(b.Select(), Equals, 2)
	c.Check(b.Check(), Equals, 2)
	c.Check(b.Select(), Equals, Checked)
}

func (s *MSSuite) BenchmarkMinefield(c *C) {
}

func (s *MSSuite) TestMinefield(c *C) {
	// mismatch points
	_, err := Minefield(make(map[Position]*Block)).init(5, 5, 5, func(width, height, max uint) ([]Position, error) {
		return []Position{}, nil
	})
	c.Check(err, Equals, ErrBadCount)

	// duplicate points
	_, err = Minefield(make(map[Position]*Block)).init(5, 5, 5, func(width, height, max uint) ([]Position, error) {
		return []Position{{1, 2}, {3, 4}, {0, 0}, {1, 2}, {4, 0}}, nil
	})
	c.Check(err, Equals, ErrDupPoint)

	// out of bounds
	_, err = Minefield(make(map[Position]*Block)).init(5, 5, 5, func(width, height, max uint) ([]Position, error) {
		return []Position{{1, 2}, {3, 4}, {0, 0}, {5, 7}, {4, 0}}, nil
	})
	c.Assert(err, Equals, ErrOutOfBounds)

	// success
	expected := Minefield(map[Position]*Block{
		Position{0, 0}: NewBlock(Mine), Position{0, 1}: NewBlock(2), Position{0, 2}: NewBlock(1), Position{0, 3}: NewBlock(1), Position{0, 4}: NewBlock(0),
		Position{1, 0}: NewBlock(2), Position{1, 1}: NewBlock(3), Position{1, 2}: NewBlock(Mine), Position{1, 3}: NewBlock(1), Position{1, 4}: NewBlock(0),
		Position{2, 0}: NewBlock(1), Position{2, 1}: NewBlock(Mine), Position{2, 2}: NewBlock(2), Position{2, 3}: NewBlock(2), Position{2, 4}: NewBlock(1),
		Position{3, 0}: NewBlock(2), Position{3, 1}: NewBlock(2), Position{3, 2}: NewBlock(1), Position{3, 3}: NewBlock(1), Position{3, 4}: NewBlock(Mine),
		Position{4, 0}: NewBlock(Mine), Position{4, 1}: NewBlock(1), Position{4, 2}: NewBlock(0), Position{4, 3}: NewBlock(1), Position{4, 4}: NewBlock(1),
	})
	minefield, err := Minefield(make(map[Position]*Block)).init(5, 5, 5, func(width, height, max uint) ([]Position, error) {
		return []Position{{1, 2}, {3, 4}, {0, 0}, {2, 1}, {4, 0}}, nil
	})
	c.Assert(err, IsNil)
	c.Check(minefield, DeepEquals, expected)
}

func (s *MSSuite) TestMinefield_Select(c *C) {
	minefield, err := Minefield(make(map[Position]*Block)).init(5, 5, 5, func(width, height, max uint) ([]Position, error) {
		return []Position{{1, 2}, {3, 4}, {0, 0}, {2, 1}, {4, 0}}, nil
	})
	c.Assert(err, IsNil)

	// out of bounds
	position, err := minefield.Select(2, 10)
	c.Assert(err, Equals, ErrOutOfBounds)

	position, err = minefield.Select(0, 1)
	c.Assert(err, IsNil)
	c.Assert(position, Equals, 2)
	position, err = minefield.Select(0, 1)
	c.Assert(err, IsNil)
	c.Assert(position, Equals, Checked)
	position, err = minefield.Select(4, 2)
	c.Assert(err, IsNil)
	c.Assert(position, Equals, 0)
	position, err = minefield.Select(0, 0)
	c.Assert(err, IsNil)
	c.Assert(position, Equals, Mine)
}

func (s *MSSuite) TestMinefield_ToggleFlag(c *C) {
	minefield, err := Minefield(make(map[Position]*Block)).init(5, 5, 5, func(width, height, max uint) ([]Position, error) {
		return []Position{{1, 2}, {3, 4}, {0, 0}, {2, 1}, {4, 0}}, nil
	})
	c.Assert(err, IsNil)

	minefield.ToggleFlag(0, 1)
	position, err := minefield.Select(0, 1)
	c.Assert(err, IsNil)
	c.Assert(position, Equals, Flagged)
	minefield.ToggleFlag(0, 1)
	position, err = minefield.Select(0, 1)
	c.Assert(err, IsNil)
	c.Assert(position, Equals, 2)
}

func (s *MSSuite) TestMinefield_Display(c *C) {
	minefield, err := Minefield(make(map[Position]*Block)).init(5, 5, 5, func(width, height, max uint) ([]Position, error) {
		return []Position{{1, 2}, {3, 4}, {0, 0}, {2, 1}, {4, 0}}, nil
	})
	c.Assert(err, IsNil)

	expected := map[Position]int{
		Position{0, 0}: Unknown, Position{0, 1}: Unknown, Position{0, 2}: Unknown, Position{0, 3}: Unknown, Position{0, 4}: Unknown,
		Position{1, 0}: Unknown, Position{1, 1}: Unknown, Position{1, 2}: Unknown, Position{1, 3}: Unknown, Position{1, 4}: Unknown,
		Position{2, 0}: Unknown, Position{2, 1}: Unknown, Position{2, 2}: Unknown, Position{2, 3}: Unknown, Position{2, 4}: Unknown,
		Position{3, 0}: Unknown, Position{3, 1}: Unknown, Position{3, 2}: Unknown, Position{3, 3}: Unknown, Position{3, 4}: Unknown,
		Position{4, 0}: Unknown, Position{4, 1}: Unknown, Position{4, 2}: Unknown, Position{4, 3}: Unknown, Position{4, 4}: Unknown,
	}
	actual := minefield.Display()
	c.Assert(actual, DeepEquals, expected)

	minefield.ToggleFlag(0, 3)

	expected[Position{0, 3}] = Flagged
	actual = minefield.Display()
	c.Assert(actual, DeepEquals, expected)

	proximity, err := minefield.Select(4, 2)
	c.Assert(err, IsNil)
	c.Assert(proximity, Equals, 0)

	expected[Position{4, 2}] = 0
	expected[Position{4, 3}] = 1
	expected[Position{3, 3}] = 1
	expected[Position{3, 2}] = 1
	expected[Position{3, 1}] = 2
	expected[Position{4, 1}] = 1
	actual = minefield.Display()
	c.Assert(actual, DeepEquals, expected)

	proximity, err = minefield.Select(4, 2)
	c.Assert(err, IsNil)
	c.Assert(proximity, Equals, Checked)
	actual = minefield.Display()
	c.Assert(actual, DeepEquals, expected)

	proximity, err = minefield.Select(4, 4)
	c.Assert(err, IsNil)
	c.Assert(proximity, Equals, 1)

	expected[Position{4, 4}] = 1
	actual = minefield.Display()
	c.Assert(actual, DeepEquals, expected)

	proximity, err = minefield.Select(0, 4)
	c.Assert(err, IsNil)
	c.Assert(proximity, Equals, 0)

	expected[Position{0, 4}] = 0
	expected[Position{1, 4}] = 0
	expected[Position{2, 4}] = 1
	expected[Position{2, 3}] = 2
	expected[Position{1, 3}] = 1
	actual = minefield.Display()
	c.Assert(actual, DeepEquals, expected)

	proximity, err = minefield.Select(0, 0)
	c.Assert(err, IsNil)
	c.Assert(proximity, Equals, Mine)

	expected[Position{0, 0}] = Mine
	expected[Position{4, 0}] = Mine
	expected[Position{1, 2}] = Mine
	expected[Position{2, 1}] = Mine
	expected[Position{3, 4}] = Mine
	actual = minefield.Display()
	c.Assert(actual, DeepEquals, expected)
}