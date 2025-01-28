package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/adambyle/menace/game"
	"github.com/adambyle/menace/menace"
)

func main() {
	m := menace.Default()
	fmt.Println("MENACE simulator")
	fmt.Println("By Adam Byle")
main:
	for {
		fmt.Println("\nChoose mode:")
		fmt.Println("X: Play against human (You are X)")
		fmt.Println("O: Play against human (You are O)")
		fmt.Println("T: Train against self")
		fmt.Println("B: Train against random-player")
		fmt.Println("R: Reset")
		fmt.Println("Q: Quit")
		var choice string
		fmt.Scanln(&choice)
		switch strings.ToLower(choice) {
		case "q":
			break main
		case "x":
			play(&m, game.O)
		case "o":
			play(&m, game.X)
		case "t":
			var count int
			fmt.Println("How many games?")
			fmt.Scanln(&count)
			train(&m, count)
		case "b":
			var count int
			fmt.Println("How many games?")
			fmt.Scanln(&count)
			trainRandom(&m, count)
		case "r":
			m = menace.Default()
		}
	}
}

type moves = map[game.Game]game.Position

func train(m *menace.Menace, count int) {
games:
	for i := range count {
		var (
			gm  = game.New()
			mvs = map[game.Symbol]moves{
				game.X: {},
				game.O: {},
			}
		)
		for !gm.Completed() {
			mv, next, moved, err := m.Move(gm)
			if err != nil {
				log.Fatal("training failure:", err)
			}
			if !moved {
				// MENACE resigned.
				m.Punish(mvs[gm.Turn()])
				continue games
			}
			mvs[gm.Turn()][gm] = mv
			gm = next
		}
		if (i+1)%5000 == 0 {
			fmt.Println(i+1, "games done")
		}
		if gm.Winner() == game.Cat {
			m.Reward(mvs[gm.Winner()], false)
			m.Reward(mvs[gm.Winner().Other()], false)
			continue
		}
		m.Reward(mvs[gm.Winner()], true)
		m.Punish(mvs[gm.Winner().Other()])
	}
	fmt.Println("Training done!")
}

func trainRandom(m *menace.Menace, count int) {
	turn := game.X
	for i := range count {
		trainOnce(m, turn)
		turn = turn.Other()
		if (i+1)%5000 == 0 {
			fmt.Println(i+1, "games done")
		}
	}
}

func trainOnce(m *menace.Menace, turn game.Symbol) {
	var (
		gm  = game.New()
		mvs = make(moves)
	)
	for !gm.Completed() {
		if gm.Turn() == turn {
			mv, next, moved, err := m.Move(gm)
			if err != nil {
				log.Fatal("training failure:", err)
			}
			if !moved {
				// MENACE resigned.
				m.Punish(mvs)
				return
			}
			mvs[gm] = mv
			gm = next
		} else {
			var (
				valid = gm.Moves()
				mi    = rand.Intn(len(valid))
				mv    = valid[mi]
			)
			next, err := gm.Move(mv)
			if err != nil {
				panic("valid move not valid")
			}
			gm = next
		}
	}
	switch gm.Winner() {
	case turn:
		m.Reward(mvs, true)
	case turn.Other():
		m.Punish(mvs)
	default:
		m.Reward(mvs, false)
	}
}

func play(m *menace.Menace, turn game.Symbol) {
	var (
		gm  = game.New()
		mvs = make(moves)
	)
	for !gm.Completed() {
		fmt.Println()
		fmt.Println(gm.Pretty())
		if gm.Turn() == turn {
			mv, next, moved, err := m.Move(gm)
			if err != nil {
				log.Fatal("illegal MENACE move:", err)
			}
			if !moved {
				fmt.Println("MENACE resigns!")
				m.Punish(mvs)
				return
			}
			fmt.Println("MENACE plays", mv)
			mvs[gm] = mv
			gm = next
		} else {
			var mv game.Position
			for {
				fmt.Println("Enter move (row col 0-2):")
				var r, c int
				fmt.Scanln(&r, &c)
				mv = game.Position{Row: r, Col: c}
				if err := mv.Valid(); err != nil {
					fmt.Println("Invalid move:", err)
					continue
				}
				next, err := gm.Move(mv)
				if err != nil {
					fmt.Println("Invalid move:", err)
					continue
				}
				gm = next
				break
			}
		}
	}
	fmt.Println()
	fmt.Println(gm.Pretty())
	switch gm.Winner() {
	case turn:
		fmt.Println("MENACE wins")
		m.Reward(mvs, true)
	case turn.Other():
		fmt.Println("MENACE loses")
		m.Punish(mvs)
	default:
		fmt.Println("Draw")
		m.Reward(mvs, false)
	}
}
