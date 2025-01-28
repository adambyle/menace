// Package game processes the functionality of a game of Tic-Tac-Toe.
//
// It also includes tools relevant to MENACE's algorithm, such as identifying
// when game boards are rotated versions of each other.
package game

import "fmt"

// Symbol is used to represent whose turn it is, spaces on a game board,
// and game outcomes.
type Symbol byte

const (
	Empty Symbol = iota // empty spaces on game board; outcome of an unfinished game
	X                   // first player
	O                   // second player
	Cat                 // outcome of a tie game
)

// Other flips X to O and O to X. It has no effect on other Symbol values.
func (s Symbol) Other() Symbol {
	switch s {
	case X:
		return O
	case O:
		return X
	default:
		return s
	}
}

// Player checks whether this is a player symbol (X or O).
func (s Symbol) Player() bool {
	return s == X || s == O
}

func (s Symbol) String() string {
	switch s {
	case X:
		return "X"
	case O:
		return "O"
	case Cat:
		return "Cat"
	default:
		return "."
	}
}

const BoardDim = 3
const Rotations = 4 // number of times you can rotate a square board

// normalizeRotations binds a number of rotations to the range 0-3.
func normalizeRotations(rots int) int {
	if rots < 0 {
		return Rotations + (rots % Rotations)
	} else {
		return rots % Rotations
	}
}

// Position represents a space on a game board.
type Position struct {
	Row, Col int
}

func (p Position) String() string {
	return fmt.Sprintf("%d,%d", p.Row, p.Col)
}

// Valid checks whether Row and Col fields are in bounds
// of the game board.
func (p Position) Valid() error {
	if p.Row < 0 || p.Row >= BoardDim || p.Col < 0 || p.Col >= BoardDim {
		return fmt.Errorf("position %v out of bounds", p)
	}
	return nil
}

// Transform changes a position to follow a board transformed in the same way.
// Rotations occur first, then transposition.
func (p Position) Transform(rots int, transpose bool) Position {
	rots = normalizeRotations(rots)
	var tp Position
	switch rots {
	case 1:
		tp.Row, tp.Col = p.Col, BoardDim-p.Row-1
	case 2:
		tp.Row, tp.Col = BoardDim-p.Row-1, BoardDim-p.Col-1
	case 3:
		tp.Row, tp.Col = BoardDim-p.Col-1, p.Row
	default:
		tp = p
	}
	if transpose {
		tp.Row, tp.Col = tp.Col, tp.Row
	}
	return tp
}

// Board contains a grid of symbols as part of a game state.
// It contains values of X, O, and Empty.
//
// The zero-value is an empty board.
type Board [BoardDim][BoardDim]Symbol

func (b Board) String() string {
	s := "["
	for r := range BoardDim {
		for c := range BoardDim {
			s += b[r][c].String()
		}
		if r != BoardDim-1 {
			s += " "
		}
	}
	return s + "]"
}

// Pretty returns a multi-line representation of the game state.
func (b Board) Pretty() string {
	s := ""
	for r := range BoardDim {
		for c := range BoardDim {
			s += b[r][c].String()
		}
		s += "\n"
	}
	return s
}

// Space returns the symbol at the given position.
// Equivalent to unchecked b[p.Row][p.Col].
func (b *Board) Space(p Position) (*Symbol, error) {
	if err := p.Valid(); err != nil {
		return nil, err
	}
	return &b[p.Row][p.Col], nil
}

// Rotate turns and/or mirrors a board over the top-left to bottom-right diagonal.
// Rotations occur first, then transposition.
func (b Board) Transform(rots int, transpose bool) Board {
	rots = normalizeRotations(rots)
	var tb Board
	switch rots {
	case 1:
		for r := range BoardDim {
			for c := range BoardDim {
				tb[r][c] = b[BoardDim-c-1][r]
			}
		}
	case 2:
		for r := range BoardDim {
			for c := range BoardDim {
				tb[r][c] = b[BoardDim-r-1][BoardDim-c-1]
			}
		}
	case 3:
		for r := range BoardDim {
			for c := range BoardDim {
				tb[r][c] = b[c][BoardDim-r-1]
			}
		}
	default:
		for r := range BoardDim {
			for c := range BoardDim {
				tb[r][c] = b[r][c]
			}
		}
	}
	if transpose {
		for r := range BoardDim {
			for c := range r {
				tb[r][c], tb[c][r] = tb[c][r], tb[r][c]
			}
		}
	}
	return tb
}

// Transformation tests if a board is a transformation of another board
// by transposition and rotation. Returns the rotations and transformations
// (in that order) on the other board needed to produce this one.
// If ok is false, the boards are not related.
func (b Board) Transformation(other Board) (rots int, transposed bool, ok bool) {
	// Test all combinations of rotations and transpositions (8).
	for rots := range Rotations {
		rb := other.Transform(rots, false)
		if rb == b {
			return rots, false, true
		}
		trb := other.Transform(rots, true)
		if trb == b {
			return rots, true, true
		}
	}
	return 0, false, false
}

// Game represents an ongoing or completed game of Tic-Tac-Toe.
// The zero-value has an invalid value of Empty for turn and cannot be played.
type Game struct {
	board Board
	turn  Symbol
}

// New creates a game with an empty board where it is X's turn.
func New() Game {
	return Game{turn: X}
}

// Board returns the board state for this game.
func (g Game) Board() Board {
	return g.board
}

// Turn returns the player whose turn it is. For valid game
// states, will be X or O.
func (g Game) Turn() Symbol {
	return g.turn
}

func (g Game) String() string {
	return fmt.Sprintf("%v %v to move", g.board, g.turn)
}

// Pretty returns a multi-line representation of the game state.
func (g Game) Pretty() string {
	w := g.Winner()
	switch w {
	case X, O:
		return fmt.Sprintf("%v\n%v wins\n", g.board.Pretty(), w)
	case Cat:
		return fmt.Sprintf("%v\nCat game\n", g.board.Pretty())
	default:
		return fmt.Sprintf("%v\n%v to move\n", g.board.Pretty(), g.turn)
	}
}

// Winner checks for a winner.
// Returns X or O if those players have made a line of their symbol.
// Returns Empty if the game is still going, and Cat if the game is a draw.
func (g Game) Winner() Symbol {
	b := g.board
	// Check for lines.
rowCheck:
	for r := range BoardDim {
		check := b[r][0]
		if !check.Player() {
			continue
		}
		for c := 1; c < BoardDim; c++ {
			if b[r][c] != check {
				continue rowCheck
			}
		}
		return check
	}
colCheck:
	for c := range BoardDim {
		check := b[0][c]
		if !check.Player() {
			continue
		}
		for r := 1; r < BoardDim; r++ {
			if b[r][c] != check {
				continue colCheck
			}
		}
		return check
	}
	// Diagonals.
	var check Symbol
	check = b[0][0]
	if check.Player() {
		complete := true
		for i := 1; i < BoardDim; i++ {
			if b[i][i] != check {
				complete = false
				break
			}
		}
		if complete {
			return check
		}
	}
	check = b[0][BoardDim-1]
	if check.Player() {
		complete := true
		for i := 1; i < BoardDim; i++ {
			if b[i][BoardDim-i-1] != check {
				complete = false
				break
			}
		}
		if complete {
			return check
		}
	}
	// Check for filled board.
	for r := range BoardDim {
		for c := range BoardDim {
			if b[r][c] == Empty {
				return Empty
			}
		}
	}
	// No empty spaces; tie game.
	return Cat
}

// Completed checks whether the game is complete.
// Equivalent to g.Winner() != Empty.
func (g Game) Completed() bool {
	return g.Winner() != Empty
}

// SpacesFilled returns the number of spaces with an X or O in it.
func (g Game) SpacesFilled() int {
	s := 0
	for r := range BoardDim {
		for c := range BoardDim {
			if g.board[r][c] != Empty {
				s++
			}
		}
	}
	return s
}

// Playable checks whether the player whose turn it is can make a move.
// Returns an error if the turn is in a bad state (not X or O),
// or if the game is won or drawn.
func (g Game) Playable() error {
	switch {
	case !g.turn.Player():
		return fmt.Errorf("invalid turn: %v", g.turn)
	case g.Completed():
		return fmt.Errorf("game is complete")
	default:
		return nil
	}
}

// Moves collects empty spaces on the game board.
//
// Returns empty if the game is not playable in this state.
func (g Game) Moves() []Position {
	if err := g.Playable(); err != nil {
		return nil
	}
	var spaces []Position
	for r := range BoardDim {
		for c := range BoardDim {
			if g.board[r][c] == Empty {
				spaces = append(spaces, Position{r, c})
			}
		}
	}
	return spaces
}

// Move creates a new game state with a symbol placed in the
// specified position and the turn switched.
func (g Game) Move(mv Position) (Game, error) {
	if err := g.Playable(); err != nil {
		return Game{}, err
	}
	s, err := g.board.Space(mv)
	if err != nil {
		return Game{}, err
	}
	if *s != Empty {
		return Game{}, fmt.Errorf("position %v is not empty", mv)
	}
	*s = g.turn
	g.turn = g.turn.Other()
	return g, nil
}
