package matchmaker

import (
	"fmt"
	"sync"
)

type (
	pool struct {
		id         string
		maxPlayers int

		m       sync.Mutex
		players []int32

		currentPlayerCount int

		respChan PoolResp
	}

	// PoolResp is the response for joining the pool
	PoolResp struct {
		PoolID   string
		IsFull   bool
		TimeIsUp bool
		Players  []int32
	}
)

// NewPool func create new pool
func newPool(id string, maxItem int) *pool {
	return &pool{
		id:         id,
		maxPlayers: maxItem,
	}
}

//Able to join? And new player is duplicate?
func (p *pool) ableToJoin(playerID int32) (bool, bool) {

	//TODO deadlock ?

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

func (p *pool) add(player int32) *PoolResp {

	p.m.Lock()

	p.players = append(p.players, player)
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
