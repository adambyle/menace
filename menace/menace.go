// Package menace simulates MENACE, the Matchbox Educable Noughts and Crosses Engine
// created by Donald Mitchie in 1961. It learns to play Tic-Tac-Toe over time.
package menace

import (
	"fmt"
	"maps"
	"math/rand"

	"github.com/adambyle/menace/game"
)

// Menace represents an instance of the MENACE machine that is able to play both X and O,
// unlike the original MENACE, which only played X.
//
// The zero-value is an invalid state. Please use New().
type Menace struct {
	// Mapping of game boards to the boxes used to decide which move to make.
	// Some valid boards are not keys to this map, but there is a transformation for
	// every such board that IS in the map.
	boxes   map[game.Board]*Box
	options *Options
}

// Default returns a MENACE instance with the default options. See DefaultOptions().
func Default() Menace {
	menace, err := New(DefaultOptions())
	if err != nil {
		panic("default options failed")
	}
	return menace
}

// New creates a distinct instance of MENACE that can learn to play for both X and O.
func New(options Options) (Menace, error) {
	// Box for the first board, which won't be discovered by traversal.
	firstBox := newBox(game.New())
	menace := Menace{
		map[game.Board]*Box{
			{}: &firstBox,
		},
		&options,
	}
	for i, b := range menace.options.Beads {
		if b < 1 {
			return Menace{}, fmt.Errorf("beads for layer %d < 0", i)
		}
	}
	if options.WinReward < 0 {
		return Menace{}, fmt.Errorf("reward is negative")
	}
	// Create boxes for all unique board states.
	const layerCount = 9
	var (
		nextNodes = []game.Game{game.New()} // nodes to process on the next step
		nodes     []game.Game               // nodes to process this layer
		layers    = [layerCount][]*Box{}    // keep track of boxes per layer for later processing
	)
	layers[0] = []*Box{&firstBox}
	// Probe layers up to 8 symbols played. Full boards do not propogate.
	for l := range layerCount {
		nodes, nextNodes = nextNodes, nil
		layerBeads := menace.options.Beads[l]
		for _, n := range nodes {
			// Completed board states can be added to nextNodes, but not processed,
			// because we always need to test for duplicate board states, even complete ones.
			// Boxes for completed games are created but not populated with beads,
			// for futures-tracking purposes.
			if n.Completed() {
				continue
			}
			var (
				moves = n.Moves()               // the valid moves on this node
				box   = menace.boxes[n.Board()] // the working node's box
				nexts = make(map[*Box]bool)     // the found/created branching games
			)
			// Collect nodes for the next layer by finding unique next game states
			// from the current one.
		moves:
			for _, mv := range moves {
				next, err := n.Move(mv)
				if err != nil {
					panic("illegal move came from Game.Moves()")
				}
				nb := next.Board()
				// If a box already exists for a similar board, link to that one and move on.
				for _, existing := range nextNodes {
					eb := existing.Board()
					if _, _, ok := nb.Transformation(eb); ok {
						ebx := menace.boxes[eb]
						// This board might STILL be a unique branch from the working node.
						// Update nexts/beads if true.
						if !nexts[ebx] {
							box.beads[mv] = layerBeads
							box.totalBeads += layerBeads
							box.nexts[mv] = ebx
							nexts[ebx] = true
						}
						continue moves
					}
				}
				nextNodes = append(nextNodes, next)
				// Register this unique move with the box corresponding
				// to the node we're branching off of.
				box.beads[mv] = layerBeads
				box.totalBeads += layerBeads
				nextBox := newBox(next)
				menace.boxes[nb] = &nextBox
				box.nexts[mv] = &nextBox
				nexts[&nextBox] = true
				if l+1 < layerCount {
					layers[l+1] = append(layers[l+1], &nextBox)
				}
			}
		}
	}
	return menace, nil
}

// Move retrieves MENACE's decision for a certain game state.
//
// If moved returns false, the specified box exists but is empty.
func (m Menace) Move(gm game.Game) (
	move game.Position, result game.Game, moved bool, err error,
) {
	var (
		b   = gm.Board()
		box = m.Box(b)
	)
	if box == nil {
		err = fmt.Errorf("no box found for %v", gm)
		return
	}
	rots, tp, ok := b.Transformation(box.game.Board())
	if !ok {
		panic("Menace.Box() returned unmatching game state")
	}
	if box.totalBeads == 0 {
		// No move made; box is empty.
		return
	}
	beadIndex := rand.Intn(box.totalBeads)
	var mv game.Position
	for m, beads := range box.beads {
		beadIndex -= beads
		if beadIndex < 0 {
			mv = m
			break
		}
	}
	tmv := mv.Transform(rots, tp)
	result, err = gm.Move(tmv)
	if err != nil {
		return
	}
	return tmv, result, true, nil
}

func (m Menace) adjust(moves map[game.Game]game.Position, amount int) {
	for gm, mv := range moves {
		var (
			gb  = gm.Board()
			box = m.Box(gb)
		)
		if box == nil {
			continue
		}
		var (
			bb          = box.game.Board()
			rots, t, ok = bb.Transformation(gb)
		)
		if !ok {
			panic("Menace.Box() returned unmatching game state")
		}
		tmv := mv.Transform(rots, t)
		box.Tune(map[game.Position]int{
			tmv: amount,
		})
		continue
	}
}

// Punish adjusts MENACE's strategy based on the choices it made for a losing game.
//
// It takes a mapping of game states to the move it made in that state.
func (m Menace) Punish(moves map[game.Game]game.Position) {
	m.adjust(moves, -1)
}

// Punish adjusts MENACE's strategy based on the choices it made for a winning or drawing game.
//
// It takes a mapping of game states to the move it made in that state.
func (m Menace) Reward(moves map[game.Game]game.Position, win bool) {
	var reward int
	if win {
		reward = m.options.WinReward
	} else {
		reward = m.options.DrawReward
	}
	m.adjust(moves, reward)
}

// Box retrieves the box for the given game state, or a transformation
// of the given board state.
func (m Menace) Box(board game.Board) *Box {
	for rots := range game.Rotations {
		var tb game.Board
		tb = board.Transform(rots, false)
		if box, ok := m.boxes[tb]; ok {
			return box
		}
		tb = board.Transform(rots, true)
		if box, ok := m.boxes[tb]; ok {
			return box
		}
	}
	return nil
}

// Box holds beads representing the move choices MENACE can make.
//
// Users should not draw beads manually from the box, as when boxes
// are retrieved using Menace.Box(), the returned box may represent
// a transformed version of the requested board. Instead, use Menace.Move().
type Box struct {
	game       game.Game
	totalBeads int
	beads      map[game.Position]int
	nexts      map[game.Position]*Box
}

func newBox(gm game.Game) Box {
	return Box{
		game:       gm,
		totalBeads: 0,
		beads:      make(map[game.Position]int),
		nexts:      make(map[game.Position]*Box),
	}
}

// Game returns the game state this box is associated with.
func (b *Box) Game() game.Game {
	return b.game
}

// TotalBeads returns the total number of beads (weights) in this box.
func (b *Box) TotalBeads() int {
	return b.totalBeads
}

// Beads returns a mapping of legal moves to the number of beads (weight)
// associated with that move.
func (b *Box) Beads() map[game.Position]int {
	return maps.Clone(b.beads)
}

// Nexts returns a mapping of legal moves to the box that results
// from that move.
func (b *Box) Nexts() map[game.Position]*Box {
	return maps.Clone(b.nexts)
}

// Tune adjusts the number of beads in boxes. It ensures only
// legal moves have beads, and that beads do not go negative.
func (b *Box) Tune(beads map[game.Position]int) {
	for mv, delta := range beads {
		if _, ok := b.beads[mv]; ok {
			capped := max(delta, -b.beads[mv])
			b.beads[mv] += capped
			b.totalBeads += capped
		}
	}
}

// Options controls MENACE's bead management.
type Options struct {
	Beads      [9]int // beads per move, depending on layer (0=start)
	WinReward  int    // beads added for MENACE's winning moves
	DrawReward int    // beads added for MENACE's drawing moves
}

// DefaultOptions returns the default MENACE bead controls.
//
// Beads:  [5 5 4 4 3 3 2 2 1]
// Reward: 3
func DefaultOptions() Options {
	return Options{
		[...]int{4, 4, 3, 3, 2, 2, 1, 1, 1},
		3,
		1,
	}
}
