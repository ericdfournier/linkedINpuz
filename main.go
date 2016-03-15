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
	"sort"
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
	P []bool     // Phenotype
	F int        // Fitness
	U uuid.UUID  // UUID
}

/* generate random integers over specified range */
func Random(min, max int) int {

	rand.Seed(time.Now().UTC().UnixNano())

	return rand.Intn(max-min) + min

}

/* view phenotype of candidate solution */
func (s Solution) Phenotype() {

	fmt.Println(s.U.String())
	fmt.Println(s.G[3])
	fmt.Println(s.G[2])
	fmt.Println(s.G[1])
	fmt.Println(s.G[0])
	fmt.Println(s.P)

}

/* view phenotypes of top solutions */
func TopPhenotypes(s chan Solution, tF int) {

	var c Solution

	for i := 0; i < cap(s); i++ {

		c = <-s

		if c.F == tF {

			c.Phenotype()

		}

	}

}

/* make function for new solution struct */
func MakeParameters() Parameters {

	QueueSize := 1000
	PoolSize := runtime.NumCPU()
	MaxEvolutions := 100

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

	Phenotype := make([]bool, 8)
	Fitness := 0
	UUID := uuid.NewV4()

	return Solution{
		G: Genotype,
		P: Phenotype,
		F: Fitness,
		U: UUID,
	}

}

/* clone solution */
func CloneSolution(s Solution) Solution {

	c := MakeSolution()
	c.P = s.P
	c.G = s.G
	c.F = s.F

	return c

}

/* evaluate solution Fitness */
func Fitness(s Solution) Solution {

	for _, seq := range sequences {

		var test string = ""

		for _, ind := range seq {

			test += s.G[ind[0]][ind[1]]

		}

		for e, exp := range expressions {

			pattern := regexp.MustCompile(exp)
			pattern.Longest()
			match := pattern.FindString(test)

			if utf8.RuneCountInString(match) < 4 {

				s.P[e] = false

			} else {

				s.P[e] = true

				/* Insert eggregious hack to overcome the go-lang regexp package's
				failure to support backreferencing expressions a-la '/1' */

				if e == 2 {

					set1, _ := regexp.MatchString(string(match[0]), string(match[2]))
					set2, _ := regexp.MatchString(string(match[1]), string(match[3]))

					if set1 && set2 {

						s.P[e] = true

					} else {

						s.P[e] = false

					}

				}

			}

		}

	}

	s.F = 0

	for _, b := range s.P {

		if b {

			s.F += 1

		}

	}

	return s

}

func TopFitness(s chan Solution) int {

	f := make([]int, cap(s))

	for i := 0; i < cap(s); i++ {

		c := <-s
		f[i] = c.F
		s <- c

	}

	sort.Ints(f)
	tF := f[len(f)-1]

	return tF

}

/* spawn new random solution */
func Spawn() Solution {

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

	s = Fitness(s)

	return s

}

/* worker pool function to parallelize spawning process */
func Spawner(q chan Solution, t chan bool, wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		select {

		case _, ok := <-t:

			if !ok {

				return

			}

			q <- Spawn()

		default:

			return

		}

	}

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

	n = Fitness(n)

	if n.F > s.F {

		return n

	} else {

		return s

	}

}

/* worker pool function to parallelize spawning process */
func Mutator(q chan Solution, t chan bool, wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		select {

		case _, ok := <-t:

			if !ok {

				return

			}

			q <- Mutate(<-q)

		default:

			return

		}

	}

}

/* perform intersection on two solutions for gene exchange */
func Intersect(s1, s2 Solution) (o1, o2 Solution) {

	iR := Random(0, 4)
	jR := Random(0, 4)
	c := MakeSolution()

	for i := 0; i < 4; i++ {

		for j := 0; j < 4; j++ {

			if i <= iR && j <= jR {

				c.G[i][j] = s1.G[i][j]

			} else {

				c.G[i][j] = s2.G[i][j]

			}

		}

	}

	c = Fitness(c)

	if c.F <= s1.F && c.F > s2.F {

		return s1, c

	} else if c.F > s1.F && c.F <= s2.F {

		return s2, c

	} else {

		return s1, s2

	}

}

/* worker pool function to parallelize spawning process */
func Intersector(q chan Solution, t chan bool, wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		select {

		case _, ok := <-t:

			if !ok {

				return
			}

			s1 := <-q
			s2 := <-q

			o1, o2 := Intersect(s1, s2)

			q <- o1
			q <- o2

		default:

			return

		}

	}

}

/* initialize a new set of candidate solutions */
func Initialize(p Parameters) chan Solution {

	fmt.Println("Initializating...")

	s := make(chan Solution, p.Q)

	st := make(chan bool, p.Q)
	mt := make(chan bool, p.Q)
	it := make(chan bool, p.Q)

	var swg, mwg, iwg sync.WaitGroup

	swg.Add(p.P)
	mwg.Add(p.P)
	iwg.Add(p.P)

	for i := 0; i < p.Q; i++ {

		st <- true
		mt <- true
		it <- true

	}

	for j := 0; j < p.P; j++ {

		go Spawner(s, st, &swg)

	}

	swg.Wait()

	for k := 0; k < p.P; k++ {

		go Mutator(s, mt, &mwg)

	}

	mwg.Wait()

	for r := 0; r < p.P; r++ {

		go Intersector(s, it, &iwg)

	}

	iwg.Wait()

	fmt.Println("Initialization Complete!!!")

	return s

}

func Evolve(s chan Solution, p Parameters) {

	mt := make(chan bool, p.Q)
	it := make(chan bool, p.Q)

	var mwg, iwg sync.WaitGroup

	mwg.Add(p.P)
	iwg.Add(p.P)

	for i := 0; i < p.Q; i++ {

		mt <- true
		it <- true

	}

	for k := 0; k < p.P; k++ {

		go Mutator(s, mt, &mwg)

	}

	mwg.Wait()

	for r := 0; r < p.P; r++ {

		go Intersector(s, it, &iwg)

	}

	iwg.Wait()

}

/* solve problem */
func Solve() {

	p := MakeParameters()
	s := Initialize(p)

	for i := 0; i < p.M; i++ {

		tF := TopFitness(s)

		if i == p.M {

			fmt.Println("Process Terminated: Maximum Evolution Count Reached...")
			fmt.Println("Top Fitness: ", tF)
			break

		}

		if tF == 8 {

			TopPhenotypes(s, tF)
			fmt.Println("Process Terminated: Convergence Achieved")
			break

		} else {

			Evolve(s, p)
			fmt.Println("Evolution ", (i + 1), " Complete...")
			fmt.Println("Top Fitness: ", tF)
			TopPhenotypes(s, tF)

		}

	}

	return

}
