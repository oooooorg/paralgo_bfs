package main

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type Node int32

type Graph struct {
	Adj [][]Node
}

func SeqBFS(g Graph, start Node) []int32 {
	dist := make([]int32, len(g.Adj))
	for i := range dist {
		dist[i] = -1
	}
	
	dist[start] = 0
	queue := make([]Node, 0, 1024)
	queue = append(queue, start)
	
	for i := 0; i < len(queue); i++ {
		u := queue[i]
		d := dist[u]
		for _, v := range g.Adj[u] {
			if dist[v] == -1 {
				dist[v] = d + 1
				queue = append(queue, v)
			}
		}
	}
	return dist
}

func ParBFS(g Graph, start Node) []int32 {
	n := len(g.Adj)
	dist := make([]int32, n)
	
	for i := range dist {
		dist[i] = -1
	}
	
	dist[start] = 0
	frontier := []Node{start}
	
	numProcs := runtime.GOMAXPROCS(0)
	
	for len(frontier) > 0 {
		frontierPiecesNext := make([][]Node, numProcs)
		var wg sync.WaitGroup
		wg.Add(numProcs)
		
		chunkSize := (len(frontier) + numProcs - 1) / numProcs
		
		for p := 0; p < numProcs; p++ {
			go func(pID int) {
				defer wg.Done()
				
				startIdx := pID * chunkSize
				endIdx := startIdx + chunkSize
				if startIdx >= len(frontier) {
					return
				}
				if endIdx > len(frontier) {
					endIdx = len(frontier)
				}
				
				localNext := make([]Node, 0, (endIdx-startIdx)*2)
				
				for i := startIdx; i < endIdx; i++ {
					u := frontier[i]
					currDist := dist[u]
					nextDist := currDist + 1
					
					for _, v := range g.Adj[u] {
						if atomic.CompareAndSwapInt32(&dist[v], -1, nextDist) {
							localNext = append(localNext, v)
						}
					}
				}
				frontierPiecesNext[pID] = localNext
			}(p)
		}
		
		wg.Wait()
		
		totalSize := 0
		for _, piece := range frontierPiecesNext {
			totalSize += len(piece)
		}
		
		if totalSize == 0 {
			break
		}
		
		nextFrontier := make([]Node, 0, totalSize)
		for _, piece := range frontierPiecesNext {
			nextFrontier = append(nextFrontier, piece...)
		}
		frontier = nextFrontier
	}
	
	return dist
}
