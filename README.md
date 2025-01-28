# Menace Simulator

This is a simulation of MENACE (Matchbox Educable Noughts and Crosses Engine), a computer made of matchboxes that can learn to play tic-tac-toe.

MENACE was first designed and built in 1961 by Donald Mitchie, and it was one of the first uses of artificial intelligence. It uses reinforcement learning to reward or punish initially random moves, based on whether the computer wins or loses.

You can read more [on Wikipedia](https://en.wikipedia.org/wiki/Matchbox_Educable_Noughts_and_Crosses_Engine).

## Background

One of my resolutions for 2025 was to learn a programming language I had never used before, so I chose Go. It's a departure from my usual favorite, Rust, in terms of memory management, syntactic complexity, and so many other things. I chose this relatively simple project as a way to get comfortable with the language. To aid my learning and familiarize myself with basic Go patterns, I didn't use generative AI in any part of the development process.

## Implementation

The original MENACE was designed only to play as the starting player, but my implementation of MENACE can be trained to play both X and O. The implementation also contains boxes for trivial states, including for the end of the game and for states where there is only one possible move. End-game boxes are empty and have no beads, but exist along with the other trivial boxes for the purpose of linking boxes to their successor boxes (see [`menace.Box.Nexts()`](menace/menace.go)).

The number of starting beads, as well as number of beads added or taken away for wins, draws, and losses, match the original MENACE.

