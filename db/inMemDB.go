package db

import (
	"sync"
	"time"
)

// NewInMemoryDB a constructor to build a new inMemoryDB.
func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{storage: make(map[string]Host), m: sync.Mutex{}}
}

// InMemoryDB a in memory DB implementing the db.HostDB interface.
type InMemoryDB struct {
	storage map[string]Host
	m       sync.Mutex
}

// WithCustomStorage allows to put a custom map as DB storage.
// Returns a new InMemoryDB.
func (db *InMemoryDB) WithCustomStorage(storage map[string]Host) *InMemoryDB {
	return &InMemoryDB{storage: storage, m: sync.Mutex{}}
}

// GetHosts returns a paginated result of all hosts.
// It returns an error if no host was found or all entries are beeing skiped.
func (db *InMemoryDB) GetHosts(skip, limit int) ([]HostInfo, error) {
	hosts := make([]HostInfo, 0)
	for _, value := range db.storage {
		hosts = append(hosts, value.HostInfo)
	}

	if len(hosts) == 0 {
		return []HostInfo{}, ErrHostsNotFound
	}

	records := len(hosts)
	if records < skip {
		return []HostInfo{}, ErrAllEntriesSkipped
	} else if records < (skip + limit) {
		return hosts[skip:], nil
	}

	return hosts[skip:(skip + limit)], nil
}

// GetHost returns a host with the matching hostname.
// If no host could be found it will return an error.
func (db *InMemoryDB) GetHost(hostname string) (HostInfo, error) {
	host, found := db.storage[hostname]
	if !found {
		return HostInfo{}, ErrHostNotFound
	}

	return host.HostInfo, nil
}

// GetStatsByHostname gets all Stats in a paginated form from a specific host.
// It returns errors if no host is found or if all entries are beeing skiped.
func (db *InMemoryDB) GetStatsByHostname(hostname string, skip, limit int) ([]Stats, error) {
	host, found := db.storage[hostname]
	if !found {
		return []Stats{}, ErrHostNotFound
	}

	records := len(host.Stats)
	if records < skip {
		return []Stats{}, ErrAllEntriesSkipped
	} else if records < (skip + limit) {
		return host.Stats[skip:], nil
	}

	return host.Stats[skip:(skip + limit)], nil
}

// InsertStats into the DB.
// To do so it takes the hostname of the Hostname field and creates a new host inside the DB and/or adds the stat to it.
// The HostInfos are also beeing updated.
// This implementation won't return an error but its declared to implement the db.HostDB interface.
func (db *InMemoryDB) InsertStats(stats Stats) error {
	db.m.Lock()
	defer db.m.Unlock()

	host, found := db.storage[stats.Hostname]
	if !found {
		hostInfo := HostInfo{Hostname: stats.Hostname, StatsAmount: 1, LastInsert: time.Now()}
		db.storage[stats.Hostname] = Host{HostInfo: hostInfo, Stats: []Stats{stats}}
		return nil
	}

	host.Stats = append(host.Stats, stats)
	host.HostInfo.StatsAmount++
	host.HostInfo.LastInsert = time.Now()

	db.storage[stats.Hostname] = host
	return nil
}