// Subgraph structure that is used to create Navigational Path Queries synthetically
package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

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

func (r *Relationship) Reflexive() bool {
	return (*r.from).label == (*r.to).label
}

func (r *Relationship) StringRep() (out string) {
	var toStar bool
	if r.Reflexive() && r1.Intn(100) > 50 {
		toStar = true
	}

	if toStar {
		out = fmt.Sprint(r.label, "*(", r.varFrom, ",", r.varTo, ")")
	} else {
		out = fmt.Sprint(r.label, "(", r.varFrom, ",", r.varTo, ")")
	}

	return out
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
	Nodes map[string]*Node
	Edges []*Relationship
}

func (g *Graph) AddNode(node Node) error {
	_, ok := g.Nodes[node.label]
	if ok {
		return errors.New("Adding node with existing label")
	}

	g.Nodes[node.label] = &node

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

	newRel := &Relationship{
		label: label,
		from:  fromNode,
		to:    toNode,
	}
	g.Edges = append(g.Edges, newRel)
	return nil
}

type Subgraph struct {
	parent      Graph            // the parent  graph, used to compute the local neighbourhood
	current     Graph            // the  nodes and edges currently selected for the subgraph
	order       map[waypoint]int // indicates the  order in which things were added to the subgraph
	parentCache map[waypoint]struct{}
}

func GetSubGraph(parent Graph) Subgraph {
	var empty struct{}

	var out Subgraph
	out.current.Nodes = make(map[string]*Node)
	out.parent = parent

	out.order = make(map[waypoint]int)
	out.parentCache = make(map[waypoint]struct{})

	for _, n := range parent.Nodes {
		out.parentCache[n] = empty
	}

	for _, e := range parent.Edges {
		out.parentCache[e] = empty
	}

	return out
}

func (s *Subgraph) Neighbourhood() (out []waypoint) {
	for _, n := range s.parent.Nodes {
		// skip if already present
		if _, ok := s.order[n]; ok {
			continue
		}

		for _, e := range s.current.Edges {
			if n.label == e.from.label || n.label == e.to.label {
				out = append(out, n)
			}
		}
	}

	for _, e := range s.parent.Edges {
		// skip if already present
		if _, ok := s.order[e]; ok {
			continue
		}

		for _, n := range s.current.Nodes {
			if n.label == e.from.label || n.label == e.to.label {
				out = append(out, e)
			}
		}
	}

	return removeDuplicate(out)
}

func (s *Subgraph) AddNode(node *Node) {
	s.order[node] = len(s.order)
	s.current.Nodes[node.label] = node
}

// RandomGrow picks a random waypoint among the local neighbourhood and adds it to the subgraph
func (s *Subgraph) RandomGrow() {
	// fmt.Println("Computing new neighbours")
	// fmt.Println("Current order: ", s.order)

	neighbours := s.Neighbourhood()

	// fmt.Println("Gotten neighbours", neighbours)

	if len(neighbours) == 0 {
		// check if we already have every waypoint possible
		if len(s.order) == len(s.parent.Edges)+len(s.parent.Nodes) {
			// fmt.Println("Size of parent ", len(s.parent.Edges)+len(s.parent.Nodes))
			return // do nothing as you can't grow more
		}

		// check if we have no neighbourhood due to new subgraph
		if len(s.order) == 0 {
			for i := range s.parent.Edges {
				neighbours = append(neighbours, s.parent.Edges[i])
			}
			for _, v := range s.parent.Nodes {
				neighbours = append(neighbours, v)
			}
		}

	}

	toPick := r1.Intn(len(neighbours))
	newItem := neighbours[toPick]

	// fmt.Println("Adding item ", newItem)

	// add new item
	s.order[newItem] = len(s.order)

	switch newItemType := newItem.(type) {
	case *Node:
		s.current.Nodes[newItemType.label] = newItemType
	case *Relationship:
		s.current.Edges = append(s.current.Edges, newItemType)
	}
}

// AllDifferntQuery produces a navigational path query where all variables are different
func (s *Subgraph) AllDifferentQuery() string {
	var sb []string

	count := 0

	for wp := range s.order {
		switch wpType := wp.(type) {
		case *Node:
			wpType.SetLabelFrom(convertToAlphabetic(count + 1))
		case *Relationship:
			wpType.SetLabelFrom(convertToAlphabetic(count + 1))
			count++
			wpType.SetLabelTo(convertToAlphabetic(count + 1))
		}
		count++
		sb = append(sb, wp.StringRep())
	}

	return strings.Join(slices.DeleteFunc(sb, func(s string) bool { return len(s) == 0 }), ", ")
}

// MaximallyJoined produces an NP query where we have as many joins as are feasible with the schema
func (s *Subgraph) MaximallyJoined() string {
	var sb []string

	for wp, count := range s.order {
		switch wpType := wp.(type) {
		case *Node:
			wpType.SetLabelFrom(convertToAlphabetic(count + 1))
		case *Relationship:
			wpType.SetLabelFrom(convertToAlphabetic(count + 1))
			count++
			wpType.SetLabelTo(convertToAlphabetic(count + 1))
		}
		sb = append(sb, wp.StringRep())
	}

	return strings.Join(slices.DeleteFunc(sb, func(s string) bool { return len(s) == 0 }), ", ")
}

type Path struct {
	stops []waypoint
}

func (p *Path) Merge(other Path) {
	p.stops = append(p.stops, other.stops...)
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

func (p *Path) String() string {
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
	out.Nodes = make(map[string]*Node)

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

func (g Graph) GetSubGraph() Subgraph {
	return GetSubGraph(g)
}

func (g Graph) GetSubGraphStartingNode(label string) (Subgraph, error) {
	node, ok := g.Nodes[label]
	if !ok {
		return Subgraph{}, errors.New("no existing node in graph with label " + label + "!")
	}

	out := GetSubGraph(g)
	out.AddNode(node)

	return out, nil
}

func (g Graph) GetPath(length int, startingNode string) (Path, error) {
	startNode, ok := g.Nodes[startingNode]
	if !ok {
		return Path{}, errors.New("starting Point nonexistant")
	}

	var tmpPath []waypoint

	tmpPath = append(tmpPath, startNode)

	if length <= 1 {
		return Path{stops: tmpPath}, nil
	}

	var neighbours []*Relationship

	for _, rel := range g.Edges {
		if rel.from.label == startingNode || rel.to.label == startingNode {
			neighbours = append(neighbours, rel)
		}
	}

	toPick := r1.Intn(len(neighbours))
	nextRelToTake := neighbours[toPick]

	// fmt.Println("Adding element, ", nextRelToTake)

	tmpPath = append(tmpPath, nextRelToTake)
	curPath := Path{stops: tmpPath}
	var remPath Path

	// direction
	if nextRelToTake.to.label == startingNode {
		remPath, _ = g.GetPath(length-1, nextRelToTake.from.label)
	} else {
		remPath, _ = g.GetPath(length-1, nextRelToTake.to.label)
	}

	curPath.Merge(remPath)

	return curPath, nil
}
