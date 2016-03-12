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
	"ENGINR",
	"MBERS",
	"LNKD",
	"RD",
	"ABCDEFGHIJLMNOPQSTUVXYZ",
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	"NECT"}

/* regular expressions that must be satisfied as constraints */
var expressions = []string{
	"[LINKED]*IN",
	"(ENG|INE|E|R)*",
	"([MBERS]*)/1",
	".(LN|K|D)*",
	"R+D",
	"[^WORK]*ING?",
	".*[IN].*",
	"C{0}N[NECT]*"}

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

/* spawn new random solution */
func Spawn(wg sync.WaitGroup) Solution {

	defer wg.Done()

	s := MakeSolution()

	for i := 0; i < 4; i++ {

		for j := 0; j < 4; j++ {

			// TODO FIGURE OUT HOW TO INDEX THIS....
			all := genes[i] + genes[j+4]
			ind := Random(0, len(all))
			s.G[i][j] = string(all[ind])

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

	for _, seq := range sequences {

		var test string = ""

		for _, ind := range seq {

			test += s.G[ind[0]][ind[1]]

		}

		for e, exp := range expressions {

			pattern := regexp.MustCompile(exp)
			match := pattern.FindString(test)

			if utf8.RuneCountInString(match) < len(test) {

				s.F[e] = false

			} else {

				s.F[e] = true

			}

		}

	}

	for _, b := range s.F {

		if b {

			s.S++

		}

	}

	return

}

/* mutate solution and retain if fitness improvement */
func (s Solution) Mutate() {

	n := s
	i := Random(0, 4)
	j := Random(0, 4)
	ind := Random(0, len(genes[i+j]))
	n.G[i][j] = string(genes[i+j][ind])
	n.Fitness()

	if n.S > s.S {

		s = n

	}

	return

}

/* worker pool function to parallelize spawning process */
func Mutator(q chan Solution) chan Solution {

	for {

		s := <-q
		s.Mutate()

		select {

		case q <- s:

		default:

			break

		}

	}

	return q

}

/* view phenotype of candidate solution */
func (s Solution) Phenotype() {

	fmt.Println(s.G[0])
	fmt.Println(s.G[1])
	fmt.Println(s.G[2])
	fmt.Println(s.G[3])

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
