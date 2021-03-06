package matchmaker

import (
	"fmt"
	"sync"

	allocation "agones.dev/agones/pkg/apis/allocation/v1"
)

type (
	pool struct {
		id         uint32
		maxPlayers int

		m       sync.Mutex
		players []uint32

		currentPlayerCount int

		respChan PoolResp
	}

	// PoolResp is the response for joining the pool
	PoolResp struct {
		PoolID   uint32
		IsFull   bool
		TimeIsUp bool
		Players  []uint32
		Gs       *allocation.GameServerAllocation
	}
)

// NewPool func create new pool
func newPool(id uint32, maxItem int) *pool {
	return &pool{
		id:         id,
		maxPlayers: maxItem,
	}
}

//Able to join? And new player is duplicate?
func (p *pool) ableToJoin(playerID uint32) (bool, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	for _, player := range p.players {
		if player == playerID {
			//Player is already in the pool
			fmt.Println("Warning: Player", playerID, "is trying to join twice!")
			return false, true
		}
	}

	return len(p.players) < p.maxPlayers, false
}

func (p *pool) add(playerID uint32) *PoolResp {

	p.m.Lock()

	//Check if duplicate
	duplicate := false

	for _, player := range p.players {
		if player == playerID {
			duplicate = true
		}
	}

	if duplicate == false {
		p.players = append(p.players, playerID)
	}

	if len(p.players) == p.maxPlayers {
		p.respChan = PoolResp{
			PoolID:  p.id,
			IsFull:  true,
			Players: p.players,
		}
	} else {
		p.respChan = PoolResp{
			PoolID:  p.id,
			IsFull:  false,
			Players: p.players,
		}
	}

	p.m.Unlock()

	return &p.respChan
}
