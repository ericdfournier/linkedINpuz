/* Copyright ©2015 The linkedINpuz Authors. All
rights reserved. Use of this source code is governed
by a BSD-style license that can be found in the
LICENSE file. */

package linkedINpuz

import (
	"fmt"
	"github.com/satori/go.uuid"
	"math/rand"
	"regexp"
	"runtime"
	"sync"
	"time"
	"unicode/utf8"
)

/* gene pool from which solutions are drawn */
var genes = []string{
	"IN",
	"ENGINER",
	"MBERS",
	"LNKD",
	"RD",
	"ABCDEFGHIJLMNOPQSTUVXYZ",
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	"NECT"}

/* regular expressions that must be satisfied as constraints */
var expressions = []string{
	`[LINKED]*IN`,
	`(ENG|INE|E|R)*`,
	`([MBERS]*)`,
	`.(LN|K|D)*`,
	`R+D`,
	`[^WORK]*ING?`,
	`.*[IN].*`,
	`C{0}N[NECT]*`}

/* sequences of the solutions that must be checked against the constraints */
var sequences = [][][]int{
	{{0, 0}, {0, 1}, {0, 2}, {0, 3}},
	{{1, 0}, {1, 1}, {1, 2}, {1, 3}},
	{{2, 0}, {2, 1}, {2, 2}, {2, 3}},
	{{3, 0}, {3, 1}, {3, 2}, {3, 3}},
	{{3, 0}, {2, 0}, {1, 0}, {0, 0}},
	{{3, 1}, {2, 1}, {1, 1}, {0, 1}},
	{{3, 2}, {2, 2}, {1, 2}, {0, 2}},
	{{3, 3}, {2, 3}, {1, 3}, {0, 3}}}

/* parameter settings for the solver */
type Parameters struct {
	Q int // Queue Size
	P int // Pool Size
	M int // Max Evolutions
}

/* candidate solution attributes*/
type Solution struct {
	G [][]string // Genotype
	F []bool     // Fitness
	S int        // Score
	U uuid.UUID  // UUID
}

/* generate random integers over specified range */
func Random(min, max int) int {

	rand.Seed(time.Now().UTC().UnixNano())

	return rand.Intn(max-min) + min

}

/* make function for new solution struct */
func MakeParameters() Parameters {

	QueueSize := 10000
	PoolSize := runtime.NumCPU()
	MaxEvolutions := 1000

	return Parameters{
		Q: QueueSize,
		P: PoolSize,
		M: MaxEvolutions,
	}

}

/* make function for new solution struct */
func MakeSolution() Solution {

	Genotype := make([][]string, 4)

	for i := 0; i < 4; i++ {

		Genotype[i] = make([]string, 4)

	}

	Fitness := make([]bool, 8)
	Score := 0
	UUID := uuid.NewV4()

	return Solution{
		G: Genotype,
		F: Fitness,
		S: Score,
		U: UUID,
	}

}

/* clone solution */
func CloneSolution(s Solution) Solution {

	c := MakeSolution()
	c.F = s.F
	c.G = s.G
	c.S = s.S

	return c

}

/* spawn new random solution */
func Spawn(wg sync.WaitGroup) Solution {

	defer wg.Done()

	s := MakeSolution()

	for i := 0; i < 4; i++ {

		for j := 0; j < 4; j++ {

			rnd := Random(0, 2)

			switch rnd {

			case 0:

				if i == 0 {
					subset := genes[0]
					ind := Random(0, len(subset))
					s.G[i][j] = string(subset[ind])
				} else if i == 1 {
					subset := genes[1]
					ind := Random(0, len(subset))
					s.G[i][j] = string(subset[ind])
				} else if i == 2 {
					subset := genes[2]
					ind := Random(0, len(subset))
					s.G[i][j] = string(subset[ind])
				} else if i == 3 {
					subset := genes[3]
					ind := Random(0, len(subset))
					s.G[i][j] = string(subset[ind])
				}

			case 1:

				if j == 0 {
					subset := genes[4]
					ind := Random(0, len(subset))
					s.G[i][j] = string(subset[ind])
				} else if j == 1 {
					subset := genes[5]
					ind := Random(0, len(subset))
					s.G[i][j] = string(subset[ind])
				} else if j == 2 {
					subset := genes[6]
					ind := Random(0, len(subset))
					s.G[i][j] = string(subset[ind])
				} else if j == 3 {
					subset := genes[7]
					ind := Random(0, len(subset))
					s.G[i][j] = string(subset[ind])
				}

			}
		}

	}

	s.Fitness()

	return s

}

/* worker pool function to parallelize spawning process */
func Spawner(q chan Solution, wg sync.WaitGroup) chan Solution {

	wg.Add(cap(q))

	for {

		select {

		case q <- Spawn(wg):

		default:

			break

		}

	}

	return q

}

/* evaluate solution Fitness */
func (s Solution) Fitness() {

	for N, seq := range sequences {

		var test string = ""

		for _, ind := range seq {

			test += s.G[ind[0]][ind[1]]

		}

		// DEBUG
		if N == 4 {
			fmt.Println(test)
		}

		for e, exp := range expressions {

			pattern := regexp.MustCompile(exp)
			pattern.Longest()
			match := pattern.FindString(test)

			if utf8.RuneCountInString(match) < 4 {

				s.F[e] = false

			} else {

				s.F[e] = true

				/* Insert eggregious hack to overcome the go-lang regexp library's
				failure to support backreferencing expressions a-la '/1' */

				if e == 2 {

					set1, _ := regexp.MatchString(string(match[0]), string(match[2]))
					set2, _ := regexp.MatchString(string(match[1]), string(match[3]))

					if set1 && set2 {

						s.F[e] = true

					} else {

						s.F[e] = false

					}

				}

			}

		}

	}

	for _, b := range s.F {

		if b {

			s.S += 1

		}

	}

	return

}

/* mutate solution and retain if fitness improvement */
func Mutate(s Solution) Solution {

	n := CloneSolution(s)
	i := Random(0, 4)
	j := Random(0, 4)

	rnd := Random(0, 2)

	switch rnd {

	case 0:

		if j == 0 {
			subset := genes[4]
			ind := Random(0, len(subset))
			n.G[i][j] = string(subset[ind])
		} else if j == 1 {
			subset := genes[5]
			ind := Random(0, len(subset))
			n.G[i][j] = string(subset[ind])
		} else if j == 2 {
			subset := genes[6]
			ind := Random(0, len(subset))
			n.G[i][j] = string(subset[ind])
		} else if j == 3 {
			subset := genes[7]
			ind := Random(0, len(subset))
			n.G[i][j] = string(subset[ind])
		}

	case 1:

		if i == 0 {
			subset := genes[0]
			ind := Random(0, len(subset))
			n.G[i][j] = string(subset[ind])
		} else if i == 1 {
			subset := genes[1]
			ind := Random(0, len(subset))
			n.G[i][j] = string(subset[ind])
		} else if i == 2 {
			subset := genes[2]
			ind := Random(0, len(subset))
			n.G[i][j] = string(subset[ind])
		} else if i == 3 {
			subset := genes[3]
			ind := Random(0, len(subset))
			n.G[i][j] = string(subset[ind])
		}

	}

	n.Fitness()

	if n.S > s.S {

		return n

	} else {

		return s

	}

}

/* view phenotype of candidate solution */
func (s Solution) Phenotype() {

	fmt.Println(s.U.String())
	fmt.Println(s.G[3])
	fmt.Println(s.G[2])
	fmt.Println(s.G[1])
	fmt.Println(s.G[0])
	fmt.Println(s.F)

}

/* worker pool function to parallelize spawning process */
func Mutator(q chan Solution) chan Solution {

	for {

		s := <-q
		s = Mutate(s)

		select {

		case q <- s:

		default:

			break

		}

	}

	return q

}

/* initialize a new set of candidate solutions */
func Initialize(p Parameters) chan Solution {

	q := make(chan Solution, p.Q)

	var wg sync.WaitGroup

	for i := 0; i < p.P; i++ {

		go Spawner(q, wg)

	}

	wg.Wait()

	return q

}

/* perform evolutionary iteration on candidate solutions */
func Evolve(q chan Solution) chan Solution {

	return q

}

/* solve Problem */
func Solve(p Parameters) {

	q := Initialize(p)

	for i := 0; i < p.M; i++ {

		if i == p.M {

			fmt.Println("Process Terminated: Maximum Evolution Count Reached")
			break

		} else {

			q = Evolve(q)

		}

	}

	return

}
