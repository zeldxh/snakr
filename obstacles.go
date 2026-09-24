package main

import "math/rand"

const numObstacles = 6

func genObstacles(snake []point) []point {
	obstacles := make([]point, 0, numObstacles)
	for len(obstacles) < numObstacles {
		p := point{rand.Intn(boardW), rand.Intn(boardH)}
		clash := false
		for _, s := range snake {
			if s == p {
				clash = true
				break
			}
		}
		if !clash {
			obstacles = append(obstacles, p)
		}
	}
	return obstacles
}

func containsPoint(pts []point, p point) bool {
	for _, o := range pts {
		if o == p {
			return true
		}
	}
	return false
}
