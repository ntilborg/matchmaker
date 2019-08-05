package matchmaker

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MatchMaker struct hold required data, and act as function receiver
type MatchMaker struct {
	maxPlayers   int
	WaitTime     time.Duration
	mutex        sync.Mutex
	pools        []*pool
	expiredPools map[string]struct{}
}

// Option struct define engine option configuration
type Option struct {
	MaxPlayers int
	WaitTime   time.Duration
}

// New function return MatchMaker struct
func New(opt Option) *MatchMaker {
	return &MatchMaker{
		maxPlayers:   opt.MaxPlayers,
		WaitTime:     opt.WaitTime,
		expiredPools: make(map[string]struct{}),
	}
}

func (m *MatchMaker) getAvailablePool(playerID int32) *pool {

	// TODO: improve find pools
	// currently we just loop through pools on engine
	for _, v := range m.pools {
		able, duplicate := v.ableToJoin(playerID)

		if duplicate {
			return nil
		}

		if able {
			return v
		}
	}

	//No pool found? Create new one
	return m.createPool()
}

func (m *MatchMaker) createPool() *pool {
	//Create the pool unique ID as string (due to int32)
	poolID := fmt.Sprintf("%d", uuid.New().ID())
	p := newPool(poolID, m.maxPlayers)

	m.mutex.Lock()
	//Add yourself as well
	m.pools = append(m.pools, p)
	m.mutex.Unlock()

	return p
}

// GetNumberOfPools return number of pools
func (m *MatchMaker) GetNumberOfPools() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.pools)
}

//Join an active or new pool
func (m *MatchMaker) Join(playerID int32) *PoolResp {
	var (
		p     = m.getAvailablePool(playerID)
		timer = time.NewTimer(m.WaitTime)
	)

	if p == nil {
		fmt.Println("Warning: Player", playerID, "is not joining a new pool!")
		return nil
	}

	go func() {
		select {
		case <-timer.C:

			fmt.Println("Time is up!")

			m.mutex.Lock()

			if _, ok := m.expiredPools[p.id]; !ok {

				p.respChan = PoolResp{
					PoolID:   p.id,
					TimeIsUp: true,
					Players:  p.players,
				}
				m.expiredPools[p.id] = struct{}{}
			}

			if p.currentPlayerCount == len(p.players) {
				// remove items on pool
				p.players = nil
				// remove pool from expired map
				delete(m.expiredPools, p.id)
			}

			fmt.Println("Cleanup finished")
			m.mutex.Unlock()
			break
		}
	}()

	return p.add(playerID)
}
