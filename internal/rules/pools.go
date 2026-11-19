package rules

import (
	"fmt"
	"math/rand"
	"sort"
)

// Entrant is a participant considered for pool generation.
type Entrant struct {
	ID     string
	ClubID string
	Seed   int
}

// PoolConstraints come from the organiser's configuration.
type PoolConstraints struct {
	TargetPoolSize     int
	MaxPools           int
	SameClubSeparation bool
	Seed               int64
}

// Pool is a generated group of fighters.
type Pool struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// GeneratePools creates as many balanced pools as the entrant count needs.
// Seeded fighters are spread first; ties are broken by same-club load and a
// controlled random factor so groups stay well mixed without being unfair.
func GeneratePools(entrants []Entrant, c PoolConstraints) []Pool {
	if len(entrants) == 0 {
		return []Pool{}
	}
	target := c.TargetPoolSize
	if target < 2 {
		target = 4
	}
	count := (len(entrants) + target - 1) / target
	if c.MaxPools > 0 && count > c.MaxPools {
		count = c.MaxPools
	}
	if count < 1 {
		count = 1
	}
	rng := rand.New(rand.NewSource(c.Seed))

	ordered := append([]Entrant(nil), entrants...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Seed != ordered[j].Seed {
			return ordered[i].Seed > ordered[j].Seed
		}
		return ordered[i].ID < ordered[j].ID
	})

	pools := make([]Pool, count)
	for i := range pools {
		pools[i] = Pool{Name: fmt.Sprintf("Pool %c", 'A'+i), Members: []string{}}
	}

	for _, entrant := range ordered {
		noise := rng.Intn(count) // controlled randomisation breaks tied pools fairly
		best := 0
		bestMembers := len(pools[0].Members)
		bestClubs := 0
		if c.SameClubSeparation {
			bestClubs = clubLoad(pools[0], entrant.ClubID, entrants)
		}
		for i := range pools {
			members := len(pools[i].Members)
			clubs := 0
			if c.SameClubSeparation {
				clubs = clubLoad(pools[i], entrant.ClubID, entrants)
			}
			better := members < bestMembers ||
				(members == bestMembers && clubs < bestClubs) ||
				(members == bestMembers && clubs == bestClubs && i == noise)
			if better {
				best, bestMembers, bestClubs = i, members, clubs
			}
		}
		pools[best].Members = append(pools[best].Members, entrant.ID)
	}
	return pools
}

func clubLoad(pool Pool, clubID string, entrants []Entrant) int {
	if clubID == "" {
		return 0
	}
	clubs := map[string]string{}
	for _, e := range entrants {
		clubs[e.ID] = e.ClubID
	}
	load := 0
	for _, member := range pool.Members {
		if clubs[member] == clubID {
			load++
		}
	}
	return load
}
