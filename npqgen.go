package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

/*   Navigational Path Query Generator
     ---------------------------------

     - This tool takes as input a schema-like structure that represents a graph-database
       and generates queries that represent randomly generated paths through the above schema.

*/

type waypoint interface {
	AmWayPoint()
}

type Relationship struct {
	label      string
	from       *Node
	to         *Node
	properties map[string]string
}

func (r Relationship) AmWayPoint() {}

func (r Relationship) String() string {
	return fmt.Sprintln(r.label)
}

type Node struct {
	label      string
	unlabelled bool // used to indicate nodes that the path jumps over, but which do not add classes
	properties map[string]string
}

func (r Node) AmWayPoint() {}

func (r Node) String() string {
  return fmt.Sprintln(r.label)
}

type Graph struct {
	Nodes map[string]Node
	Edges []Relationship
}

func (g *Graph) AddNode(node Node) error {
	_, ok := g.Nodes[node.label]
	if ok {
		return errors.New("Adding node with existing label")
	}

	g.Nodes[node.label] = node

	return nil
}

func (g *Graph) AddRel(label string, from, to string) error {
	fromNode, ok := g.Nodes[from]
	if !ok {
		return errors.New("Node not found")
	}
	toNode, ok := g.Nodes[to]
	if !ok {
		return errors.New("Node not found")
	}

	newRel := Relationship{
		label: label,
		from:  &fromNode,
		to:    &toNode,
	}
	g.Edges = append(g.Edges, newRel)
	return nil
}

// TODO:
// - Create a manual representation for now
// - Simple algorithm that picks a node and searches blindly for a number of steps
// - then produces NPGQ based on that trace

func GetManualRep() Graph {
	var out Graph
	out.Nodes = make(map[string]Node)

	pedestrian := Node{label: "pedestrian"}
	pedestrian_mov := Node{label: "pedestrian_moving"}
	pedestrian_stat := Node{label: "pedestrian_stationay"}
	sample := Node{label: "sample"}
	unlabelA := Node{label: "unlabelA", unlabelled: true}
	unlabelB := Node{label: "unlabelB", unlabelled: true}
	unlabelC := Node{label: "unlabelC", unlabelled: true}

	out.AddNode(pedestrian)
	out.AddNode(pedestrian_mov)
	out.AddNode(pedestrian_stat)
	out.AddNode(sample)
	out.AddNode(unlabelA)
	out.AddNode(unlabelB)
	out.AddNode(unlabelC)

	out.AddRel("of", "unlabelA", "pedestrian")
	out.AddRel("first_annotation", "pedestrian", "unlabelB")
	out.AddRel("first_annotation", "pedestrian", "unlabelB")
	out.AddRel("of", "unlabelB", "sample")
	out.AddRel("next", "sample", "sample")
	out.AddRel("has", "unlabelA", "pedestrian_moving")
	out.AddRel("next", "pedestrian_moving", "pedestrian_stationay")
	out.AddRel("has", "unlabelC", "pedestrian_stationay")

	return out
}

func (g Graph) GetPath(length int, startingNode string) ([]waypoint, error) {
	startNode, ok := g.Nodes[startingNode]
	if !ok {
		return []waypoint{}, errors.New("starting Point nonexistant")
	}

	var tmpPath []waypoint

	tmpPath = append(tmpPath, startNode)

	if length <= 1 {
		return tmpPath, nil
	}

	var neighbours []Relationship

	for _, rel := range g.Edges {
		if rel.from.label == startingNode || rel.to.label == startingNode {
			neighbours = append(neighbours, rel)
		}
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	toPick := r1.Intn(len(neighbours))

	nextRelToTake := neighbours[toPick]

	tmpPath = append(tmpPath, nextRelToTake)
	var remPath []waypoint
	// direction
	if nextRelToTake.to.label == startingNode {
		remPath, _ = g.GetPath(length-1, nextRelToTake.from.label)
	} else {
		remPath, _ = g.GetPath(length-1, nextRelToTake.to.label)
	}

	return append(tmpPath, remPath...), nil
}

func main() {
	// ==============================================
	// Command-Line Argument Parsing

	// flagSet := flag.NewFlagSet("npqGen", flag.ExitOnError)

	// // input flags
	// graph := flagSet.String("graph", "", "The graph over which new queries are generated.")

	// flagSet.Parse(os.Args[1:])

	// if *graph == "" {
	// 	flagSet.Usage()
	// 	os.Exit(-1)
	// }

	// // END Command-Line Argument Parsing
	// // ==============================================

	graph := GetManualRep()

	path, _ := graph.GetPath(3, "pedestrian")

	fmt.Println("Produced path", path)
}
