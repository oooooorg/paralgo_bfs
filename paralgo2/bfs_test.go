package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

var (
	testGraph Graph
	N         = 300
)

func TestMain(m *testing.M) {
	testGraph = genGraph(N)
	code := m.Run()
	
	os.Exit(code)
}

func TestCorrectness(t *testing.T) {
	startNode := Node(0)
	
	t.Run("seq_bfs", func(t *testing.T) {
		dist := SeqBFS(testGraph, startNode)
		if err := checkCorrect(dist, N); err != nil {
			t.Errorf("Sequential BFS no correct: %v", err)
		}
	})
	
	t.Run("par_bfs (4 procs)", func(t *testing.T) {
		runtime.GOMAXPROCS(4)
		
		dist := ParBFS(testGraph, startNode)
		if err := checkCorrect(dist, N); err != nil {
			t.Errorf("Parallel BFS no correct: %v", err)
		}
	})
}

func TestPerformanceComparison(t *testing.T) {
	runs := 5
	startNode := Node(0)
	
	runtime.GC()
	
	runtime.GOMAXPROCS(1)
	var totalSeq time.Duration
	fmt.Printf("Sequential BFS\n")
	
	for i := 0; i < runs; i++ {
		runtime.GC()
		start := time.Now()
		SeqBFS(testGraph, startNode)
		dur := time.Since(start)
		totalSeq += dur
		fmt.Printf("Run %d: %v\n", i+1, dur)
	}
	avgSeq := totalSeq / time.Duration(runs)
	
	runtime.GOMAXPROCS(4)
	
	var totalPar time.Duration
	fmt.Printf("Parallel BFS\n")
	for i := 0; i < runs; i++ {
		runtime.GC()
		start := time.Now()
		ParBFS(testGraph, startNode)
		dur := time.Since(start)
		totalPar += dur
		fmt.Printf("Run %d: %v\n", i+1, dur)
	}
	avgPar := totalPar / time.Duration(runs)
	
	speedup := float64(avgSeq) / float64(avgPar)
	
	fmt.Printf("\nSpeedup is %.2fx\n", speedup)
}

func BenchmarkBFS_Seq(b *testing.B) {
	runtime.GOMAXPROCS(1)
	startNode := Node(0)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		SeqBFS(testGraph, startNode)
	}
}

func BenchmarkBFS_Par_4Procs(b *testing.B) {
	runtime.GOMAXPROCS(4)
	startNode := Node(0)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ParBFS(testGraph, startNode)
	}
}

func genGraph(n int) Graph {
	totalNodes := n * n * n
	adj := make([][]Node, totalNodes)
	
	idx := func(x, y, z int) Node {
		return Node(x*n*n + y*n + z)
	}
	
	dirs := [][3]int{
		{1, 0, 0}, {-1, 0, 0},
		{0, 1, 0}, {0, -1, 0},
		{0, 0, 1}, {0, 0, -1},
	}
	
	var wg sync.WaitGroup
	workers := runtime.NumCPU()
	chunk := (totalNodes + workers - 1) / workers
	
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(pID int) {
			defer wg.Done()
			
			start := pID * chunk
			end := start + chunk
			if end > totalNodes {
				end = totalNodes
			}
			
			for i := start; i < end; i++ {
				tmp := i
				z := tmp % n
				tmp /= n
				y := tmp % n
				x := tmp / n
				
				neighbors := make([]Node, 0, 6)
				for _, d := range dirs {
					nx, ny, nz := x+d[0], y+d[1], z+d[2]
					if nx >= 0 && nx < n && ny >= 0 && ny < n && nz >= 0 && nz < n {
						neighbors = append(neighbors, idx(nx, ny, nz))
					}
				}
				adj[i] = neighbors
			}
		}(w)
	}
	wg.Wait()
	
	return Graph{Adj: adj}
}

func checkCorrect(dist []int32, n int) error {
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			for z := 0; z < n; z++ {
				id := int32(x*n*n + y*n + z)
				expected := int32(x + y + z)
				if dist[id] != expected {
					return fmt.Errorf("Wrong dist for node (%d,%d,%d) [id %d]: got %d, want %d",
						x, y, z, id, dist[id], expected)
				}
			}
		}
	}
	return nil
}
