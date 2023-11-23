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

// GetManualRep is a hard-coded representation of the schema graph of the SCENES dataset
func GetManualRep() Graph {
	var out Graph
	out.Nodes = make(map[string]*Node)

	// nodes

	// instance-labels
	out.AddNode(Node{label: "pedestrian"})
	out.AddNode(Node{label: "vehicle"})
	out.AddNode(Node{label: "car"})
	out.AddNode(Node{label: "bus"})
	out.AddNode(Node{label: "bicycle"})
	out.AddNode(Node{label: "animals"})
	out.AddNode(Node{label: "debris"})
	out.AddNode(Node{label: "trafficcone"})
	out.AddNode(Node{label: "bicycle_rack"})

	// attribute-labels
	out.AddNode(Node{label: "pedestrian_moving"})
	out.AddNode(Node{label: "pedestrian_standing"})
	out.AddNode(Node{label: "vehicle_moving"})
	out.AddNode(Node{label: "vehicle_stopped"})
	out.AddNode(Node{label: "vehicle_parked"})
	out.AddNode(Node{label: "cycle_with_rider"})
	out.AddNode(Node{label: "cycle_without_rider"})
	out.AddNode(Node{label: "stationary"})
	out.AddNode(Node{label: "walking"})
	out.AddNode(Node{label: "running"})
	out.AddNode(Node{label: "sitting"})

	// basic-labels
	out.AddNode(Node{label: "sample"})
	out.AddNode(Node{label: "instance", unlabelled: true})
	out.AddNode(Node{label: "sample_annotation", unlabelled: true})

	// relationships
	// to instance-labels
	out.AddRel("OF", "instance", "pedestrian")
	out.AddRel("OF", "instance", "vehicle")
	out.AddRel("OF", "instance", "bus")
	out.AddRel("OF", "instance", "bicycle")
	out.AddRel("OF", "instance", "bicycle_rack")
	out.AddRel("OF", "instance", "trafficcone")
	out.AddRel("OF", "instance", "debris")
	out.AddRel("OF", "instance", "animals")
	out.AddRel("OF", "instance", "car")

	// annotation-labels
	out.AddRel("FIRST_ANNOTATION", "pedestrian", "sample_annotation")
	out.AddRel("LAST_ANNOTATION", "pedestrian", "sample_annotation")
	out.AddRel("FIRST_ANNOTATION", "vehicle", "sample_annotation")
	out.AddRel("LAST_ANNOTATION", "vehicle", "sample_annotation")
	out.AddRel("FIRST_ANNOTATION", "bus", "sample_annotation")
	out.AddRel("LAST_ANNOTATION", "bus", "sample_annotation")
	out.AddRel("FIRST_ANNOTATION", "bicycle", "sample_annotation")
	out.AddRel("LAST_ANNOTATION", "bicycle", "sample_annotation")
	out.AddRel("FIRST_ANNOTATION", "bicycle_rack", "sample_annotation")
	out.AddRel("LAST_ANNOTATION", "bicycle_rack", "sample_annotation")
	out.AddRel("FIRST_ANNOTATION", "trafficcone", "sample_annotation")
	out.AddRel("LAST_ANNOTATION", "trafficcone", "sample_annotation")
	out.AddRel("FIRST_ANNOTATION", "debris", "sample_annotation")
	out.AddRel("LAST_ANNOTATION", "debris", "sample_annotation")
	out.AddRel("FIRST_ANNOTATION", "animals", "sample_annotation")
	out.AddRel("LAST_ANNOTATION", "animals", "sample_annotation")
	out.AddRel("FIRST_ANNOTATION", "car", "sample_annotation")
	out.AddRel("LAST_ANNOTATION", "car", "sample_annotation")

	// basic-relations
	out.AddRel("OF", "sample_annotation", "sample")
	out.AddRel("NEXT", "sample", "sample")
	out.AddRel("NEXT", "instance", "instance")
	// attribute-relations
	out.AddRel("HAS", "instance", "pedestrian_moving")
	out.AddRel("HAS", "instance", "pedestrian_standing")
	out.AddRel("HAS", "instance", "vehicle_moving")
	out.AddRel("HAS", "instance", "vehicle_stopped")
	out.AddRel("HAS", "instance", "vehicle_parked")
	out.AddRel("HAS", "instance", "cycle_with_rider")
	out.AddRel("HAS", "instance", "cycle_without_rider")
	out.AddRel("HAS", "instance", "stationary")
	out.AddRel("HAS", "instance", "walking")
	out.AddRel("HAS", "instance", "sitting")
	out.AddRel("HAS", "instance", "running")

	return out
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

	// end command-Line Argument Parsing
	// ==============================================

	if subgraph {
		subgraph, err := graph.GetSubGraphStartingNode(startWith)
		check(err)

		for i := 0; i <= length; i++ {
			subgraph.RandomGrow()
		}

		fmt.Println("internal:")
		for i, k := range subgraph.order {
			fmt.Println(i, " ", convertToAlphabetic(k+1))
		}

		fmt.Println("Produced subgraph query", subgraph.MaximallyJoined())
	} else {
		path, err := graph.GetPath(length, startWith)
		check(err)

		fmt.Println("internal:")
		for i := range path.stops {
			fmt.Println(path.stops[i])
		}
		fmt.Println("Produced path query", path.String())
	}
}
