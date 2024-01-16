package main

import (
	"container/heap"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/fogleman/gg"
	"github.com/kelvins/geocoder"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

const R = 6371

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

type Node struct {
	x, y    float64
	g, h, f float64
	parent  *Node
}

type PriorityQueue []*Node

func (pq PriorityQueue) Len() int            { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool  { return pq[i].f < pq[j].f }
func (pq PriorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(x interface{}) { *pq = append(*pq, x.(*Node)) }
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func heuristic(a, b *Node) float64 {
	return math.Hypot(a.x-b.x, a.y-b.y)
}

func aStar(start, goal *Node, nodes []*Node) []*Node {
	openSet := &PriorityQueue{}
	heap.Init(openSet)
	heap.Push(openSet, start)

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)
		if current == goal {
			path := []*Node{}
			for current != nil {
				path = append([]*Node{current}, path...)
				current = current.parent
			}
			return path
		}

		for _, neighbor := range nodes {
			if neighbor == current {
				continue
			}
			tentativeG := current.g + heuristic(current, neighbor)
			if tentativeG < neighbor.g {
				neighbor.parent = current
				neighbor.g = tentativeG
				neighbor.h = heuristic(neighbor, goal)
				neighbor.f = neighbor.g + neighbor.h
				heap.Push(openSet, neighbor)
			}
		}
	}
	return nil
}

func kMeans(points []orb.Point, k int) [][]orb.Point {
	centroids := make([]orb.Point, k)
	clusters := make([][]orb.Point, k)

	for i := range centroids {
		centroids[i] = points[rand.Intn(len(points))]
	}

	for {
		for i := range clusters {
			clusters[i] = nil
		}

		for _, point := range points {
			minDist := math.MaxFloat64
			var minIndex int
			for i, centroid := range centroids {
				dist := planar.Distance(point, centroid)
				if dist < minDist {
					minDist = dist
					minIndex = i
				}
			}
			clusters[minIndex] = append(clusters[minIndex], point)
		}

		newCentroids := make([]orb.Point, k)
		for i, cluster := range clusters {
			var sumX, sumY float64
			for _, point := range cluster {
				sumX += point.X()
				sumY += point.Y()
			}
			newCentroids[i] = orb.Point{sumX / float64(len(cluster)), sumY / float64(len(cluster))}
		}

		if planar.Equal(centroids, newCentroids) {
			break
		}
		centroids = newCentroids
	}

	return clusters
}

func main() {

	file, err := os.Open("data.geojson")
	if err != nil {
		log.Fatalf("Error opening GeoJSON file: %v", err)
	}
	defer file.Close()

	fc, err := geojson.DecodeFeatureCollection(file)
	if err != nil {
		log.Fatalf("Error decoding GeoJSON: %v", err)
	}

	fmt.Println("Number of features:", len(fc.Features))

	const S = 1024
	dc := gg.NewContext(S, S)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

	for _, feature := range fc.Features {
		if geom, ok := feature.Geometry.(orb.LineString); ok {
			for i := 0; i < len(geom)-1; i++ {
				dc.DrawLine(geom[i][0], geom[i][1], geom[i+1][0], geom[i+1][1])
			}
		}
	}

	dc.Stroke()
	dc.SavePNG("output.png")

	lat1, lon1 := 52.2296756, 21.0122287
	lat2, lon2 := 41.8919300, 12.5113300
	distance := haversine(lat1, lon1, lat2, lon2)
	fmt.Printf("Distance: %.2f km\n", distance)

	start := &Node{x: 0, y: 0}
	goal := &Node{x: 10, y: 10}
	nodes := []*Node{
		start,
		goal,
		{x: 1, y: 1},
		{x: 2, y: 2},
		{x: 3, y: 3},
		{x: 4, y: 4},
		{x: 5, y: 5},
	}

	path := aStar(start, goal, nodes)
	for _, node := range path {
		fmt.Printf("Node: (%.2f, %.2f)\n", node.x, node.y)
	}

	address := geocoder.Address{
		Street:  "Central Park West",
		Number:  115,
		City:    "New York",
		State:   "New York",
		Country: "United States",
	}

	location, err := geocoder.Geocoding(address)
	if err != nil {
		log.Fatalf("Error geocoding address: %v", err)
	}

	fmt.Printf("Latitude: %.6f, Longitude: %.6f\n", location.Latitude, location.Longitude)

	points := []orb.Point{
		{1, 1}, {2, 2}, {3, 3}, {8, 8}, {9, 9}, {10, 10},
	}
	clusters := kMeans(points, 2)
	for i, cluster := range clusters {
		fmt.Printf("Cluster %d:\n", i)
		for _, point := range cluster {
			fmt.Printf("  (%.2f, %.2f)\n", point.X(), point.Y())
		}
	}
}
