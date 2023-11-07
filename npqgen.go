package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

/*   Navigational Path Query Generator
     ---------------------------------

     - This tool takes as input a schema-like structure that represents a graph-database
       and generates queries that represent randomly generated paths through the above schema.

*/

type WaypointMode int

const (
	NodeType WaypointMode = iota
	UnLabel
	RelType
	RelInv
	// UnRel
)

type waypoint interface {
	AmWayPoint()
	GetMode(prev waypoint) WaypointMode
	GetLabel() string
	SetLabelFrom(variable string)
	SetLabelTo(variable string)
	StringRep() string
}

type Relationship struct {
	label      string
	from       *Node
	to         *Node
	properties map[string]string
	varFrom    string
	varTo      string
}

func (r *Relationship) StringRep() string {
	return fmt.Sprint(r.label, "(", r.varFrom, ",", r.varTo, ")")
}

func (r *Relationship) AmWayPoint() {}

func (r *Relationship) GetLabel() string { return r.label }

func (r *Relationship) SetLabelFrom(variable string) { r.varFrom = variable }
func (r *Relationship) SetLabelTo(variable string)   { r.varTo = variable }

func (r *Relationship) GetMode(prev waypoint) WaypointMode {
	if prev.GetLabel() == r.from.label {
		return RelType
	}
	return RelInv
}

func (r *Relationship) String() string {
	return fmt.Sprintln(r.label)
}

type Node struct {
	label      string
	unlabelled bool // used to indicate nodes that the path jumps over, but which do not add classes
	properties map[string]string
	varFrom    string
	varTo      string
}

func (r *Node) AmWayPoint() {}

func (r *Node) StringRep() string {
	if r.unlabelled {
		return ""
	}
	return fmt.Sprint(r.label, "(", r.varFrom, ")")
}

func (r *Node) GetLabel() string { return r.label }

func (r *Node) SetLabelFrom(variable string) { r.varFrom = variable }
func (r *Node) SetLabelTo(variable string)   { r.varTo = variable }

func (r *Node) GetMode(prev waypoint) WaypointMode {
	if r.unlabelled {
		return UnLabel
	}
	return NodeType
}

func (r *Node) String() string {
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

type Path struct {
	stops []waypoint
}

func convertToAlphabetic(n int) string {
	result := ""
	for n > 0 {
		mod := (n - 1) % 26
		result = string('a'+mod) + result
		n = (n - mod) / 26
	}
	return result
}

func (p Path) String() string {
	var elements []string

	if len(p.stops) == 0 {
		return ""
	}
	lenSoFar := 0
	prev := p.stops[0]
	var prevPrev waypoint = (*Node)(nil)
	prev.SetLabelFrom(convertToAlphabetic(lenSoFar))

	for _, next := range p.stops {

		prevVar := convertToAlphabetic(lenSoFar)

		switch next.GetMode(prev) {
		case NodeType, UnLabel: // need new label

			lenSoFar++
			nextVar := convertToAlphabetic(lenSoFar)
			next.SetLabelFrom(nextVar)

			if prevPrev != nil && prev.GetMode(prevPrev) == RelInv {
				prev.SetLabelFrom(nextVar)
			} else {
				prev.SetLabelTo(nextVar)
			}

			elements = append(elements, prev.StringRep())

			next.SetLabelTo(nextVar)
		case RelType: // don't need label
			next.SetLabelFrom(prevVar)
		case RelInv: // don't need new label
			next.SetLabelTo(prevVar)
			// case UnRel: // two new label
			// 	next.SetLabelFrom(nextVar)
			// 	lenSoFar++
			// 	nextestVar := convertToAlphabetic(lenSoFar)
			// 	next.SetLabelTo(nextestVar)
		}
		prevPrev = prev
		prev = next
	}
	if prevPrev != nil && prev.GetMode(prevPrev) == NodeType {
		elements = append(elements, prev.StringRep())
	}

	return strings.Join(elements, ", ")
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
	pedestrian_stat := Node{label: "pedestrian_standing"}

	vehicle := Node{label: "vehicle"}
	vehicle_moving := Node{label: "vehicle_moving"}
	vehicle_stopped := Node{label: "vehicle_stopped"}
	vehicle_parked := Node{label: "vehicle_parked"}

	bus := Node{label: "bus"}

	bicycle := Node{label: "bicycle"}

	cycle_with_rider := Node{label: "cycle_with_rider"}
	cycle_without_rider := Node{label: "cycle_without_rider"}

	sample := Node{label: "sample"}
	instance := Node{label: "instance", unlabelled: true}
	sample_annotation := Node{label: "sample_annotation", unlabelled: true}

	out.AddNode(pedestrian)
	out.AddNode(pedestrian_mov)
	out.AddNode(pedestrian_stat)
	out.AddNode(vehicle)
	out.AddNode(vehicle_moving)
	out.AddNode(vehicle_stopped)
	out.AddNode(vehicle_parked)
	out.AddNode(bus)
	out.AddNode(bicycle)
	out.AddNode(cycle_with_rider)
	out.AddNode(cycle_without_rider)
	out.AddNode(sample)
	out.AddNode(instance)
	out.AddNode(sample_annotation)

	out.AddRel("OF", "instance", "pedestrian")
	out.AddRel("OF", "instance", "vehicle")
	out.AddRel("OF", "instance", "bus")
	out.AddRel("OF", "instance", "bicycle")
	out.AddRel("FIRST_ANNOTATION", "pedestrian", "sample_annotation")
	out.AddRel("LAST_ANNOTATION", "pedestrian", "sample_annotation")
	out.AddRel("OF", "sample_annotation", "sample")
	out.AddRel("NEXT", "sample", "sample")
	out.AddRel("NEXT", "instance", "instance")
	out.AddRel("HAS", "instance", "pedestrian_moving")
	out.AddRel("HAS", "instance", "pedestrian_standing")
	out.AddRel("HAS", "instance", "vehicle_moving")
	out.AddRel("HAS", "instance", "vehicle_stopped")
	out.AddRel("HAS", "instance", "vehicle_parked")
	out.AddRel("HAS", "instance", "cycle_with_rider")
	out.AddRel("HAS", "instance", "cycle_without_rider")

	return out
}

func (g Graph) GetPath(length int, startingNode string) ([]waypoint, error) {
	startNode, ok := g.Nodes[startingNode]
	if !ok {
		return []waypoint{}, errors.New("starting Point nonexistant")
	}

	var tmpPath []waypoint

	tmpPath = append(tmpPath, &startNode)

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

	tmpPath = append(tmpPath, &nextRelToTake)
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
	graph := GetManualRep()

	// ==============================================
	// Command-Line Argument Parsing

	flagSet := flag.NewFlagSet("npqgen", flag.ExitOnError)

	// input flags
	// graph := flagSet.String("graph", "", "The graph over which new queries are generated.")
	length := flagSet.Int("length", 0, "The length of the path that defines the query.")
	startWith := flagSet.String("startWith", "pedestrian", "The starting node. Must exist.")

	flagSet.Parse(os.Args[1:])

	if _, ok := graph.Nodes[*startWith]; !ok || *length == 0 {
		flagSet.Usage()
		os.Exit(-1)
	}

	// END Command-Line Argument Parsing
	// ==============================================

	path, _ := graph.GetPath(*length, *startWith)

	fmt.Println("Produced path", path)
	fmt.Println("Produced path", Path{stops: path})
}
