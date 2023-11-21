package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	flag "github.com/spf13/pflag"
)

/*   Navigational Path Query Generator
     ---------------------------------

     - This tool takes as input a schema-like structure that represents a graph-database
       and generates queries that represent randomly generated paths through the above schema.

*/

// To fix: add a starting point for subgraph
// To fix: path computation broken for some reason

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var (
	s1 rand.Source = rand.NewSource(time.Now().UnixNano())
	r1 *rand.Rand  = rand.New(s1)
)

func removeDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func main() {
	graph := GetManualRep()

	// ==============================================
	// Command-Line Argument Parsing

	flagSet := flag.NewFlagSet("npqgen", flag.ExitOnError)

	// input flags
	// graph := flagSet.String("graph", "", "The graph over which new queries are generated.")
	var length int
	flagSet.IntVarP(&length, "length", "l", 0, "The length of the path that defines the query.")
	var startWith string
	flagSet.StringVarP(&startWith, "startWith", "s", "pedestrian", "The starting node. Must exist.")
	var subgraph bool
	flagSet.BoolVarP(&subgraph, "branching", "b", false,
		"Generate queries that feature use general subgraphs as their structure.")

	flagSet.Parse(os.Args[1:])

	if _, ok := graph.Nodes[startWith]; !ok || length == 0 {
		if flagSet == nil {
			log.Panicln("wat")
		}
		fmt.Println("Usage of npqgen: ")
		flagSet.PrintDefaults()
		os.Exit(-1)
	}

	// END Command-Line Argument Parsing
	// ==============================================

	if subgraph {
		subgraph, err := graph.GetSubGraphStartingNode(startWith)
		check(err)

		for i := 0; i <= length; i++ {
			subgraph.RandomGrow()
		}

		fmt.Println("Produced subgraph query", subgraph.MaximallyJoined())
	} else {
		path, err := graph.GetPath(length, startWith)
		check(err)

		fmt.Println("Produced path query", path.String())
	}
}
