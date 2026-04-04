/* This is a cache helper. Maybe the host wants low effort
   caching, maybe the host wants caching in memory, maybe
   the host wants caching on the disk, so we create helper
   functions to determine for us what kind of cache the
   host wants and to act depending on the kind of cache.

   These functions are for the high effort caching system.
   I don't know how to get keys for fiber's caching
   middleware, it's why I recommend high effort caching. */

package main

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/DmitriyVTitov/size"
	"github.com/dgraph-io/ristretto"
)

// Disk-based caches
var userCaches sync.Map;
// Global memory-based cache
var memoryCache *ristretto.Cache;

// Cache initiator for both memory-based and disk-based caches
func initCache() error {
	go func() {
		var sig = make(chan os.Signal, 1);
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM);

		<-sig
		userCaches.Range(func(key, value any) bool {
			if db, ok := value.(*sql.DB); ok {
				db.Close();
			}
			return true;
		});

		os.Exit(0);
	}();
	if settings.Cache.Memory {
		if settings.Cache.Path != "" {
			if files, err1 := os.ReadDir(settings.Cache.Path); len(files) == 0 && err1 != nil {
				return err1;
			} else {
				for _, file := range files {
					os.Remove(filepath.Join(settings.Cache.Path, file.Name()));
				}
			}
		}
		if cache, err1 := ristretto.NewCache(&ristretto.Config{
			NumCounters: int64(settings.Cache.MaxSize * 10),
			MaxCost: int64(settings.Cache.MaxSize),
			BufferItems: 64,
		}); err1 != nil {
			return err1;
		} else {
			memoryCache = cache;
		}
		return nil;
	} else {
		for _, user := range settings.Users {
			var path string = filepath.Join(settings.Cache.Path, fmt.Sprintf("%s.db", user.Name));
			if f, err1 := os.Create(path); err1 != nil {
				return err1;
			} else {
				f.Close();
				var db, err2 = sql.Open("sqlite3", path);
				if err2 != nil {
					return err2;
				}
				userCaches.Store(user.Name, db);
				if _, err3 := db.Exec("PRAGMA journal_mode=WAL"); err3 != nil {
					return err3;
				}
				if _, err3 := db.Exec(`create table if not exists cache (
					key TEXT PRIMARY KEY,
					value BLOB,
					recent_use INTEGER
				)`); err3 != nil {
					return err3;
				}
				if _, err3 := db.Exec("create index if not exists idx_key on cache(key)"); err3 != nil {
					return err3;
				}
				db.SetMaxOpenConns(1);
			}
		}
		return nil
	}
}

// Remove unpopular items for others. For disk-based caches (!settings.Cache.Memory)
func cleanCache() error {
	for _, user := range settings.Users {
		var path string = filepath.Join(settings.Cache.Path, fmt.Sprintf("%s.db", user.Name));
		if stat, err1 := os.Stat(path); err1 != nil {
			return err1;
		} else {
			if stat.Size() > int64(settings.Cache.MaxSize) {
				var db, _ = userCaches.Load(user.Name);

				if _, err2 := db.(*sql.DB).Exec("delete from cache where recent_use < ?", time.Now().UnixMilli()); err2 != nil {
					return err2;
				}
			}
		}
	}
	return nil;
}

// Get a value from a user's cache
func getCache(user string, key string) (interface{}, bool, error) {
	if settings.Cache.Memory {
		if v, o := memoryCache.Get(user+":"+key); o {
			return v, o, nil
		} else { return nil, o, nil; }
	} else {
		if err := cleanCache(); err != nil {
			return nil, false, err;
		}
		if db, ok := userCaches.Load(user); !ok {
			return nil, false, nil;
		} else {
			var query, err = db.(*sql.DB).Query("select value from cache where key=?", key);
			defer query.Close();
			if err != nil { return nil, false, err; }
			if query.Next() {
				var val []byte;
				if err1 := query.Scan(&val); err1 != nil {
					return nil, false, err1;
				}

				go func(k string) {
					db.(*sql.DB).Exec("update cache set recent_use = ? where key = ?", time.Now().UnixMilli(), k);
				}(key);
				switch val[0] {
					case 0x00: return val[1:], true, nil;
					case 0x01: return string(val[1:]), true, nil;
					case 0x02: return val[1] == 0x01, true, nil;
					case 0x03: return gob.NewDecoder(bytes.NewReader(val)), true, nil;
				}
			}
			return nil, false, nil;
		}
	}
}

// Set a value to a user's cache
func setCache(user string, key string, value interface{}) (bool, error) {
	if settings.Cache.Memory {
		var set bool = memoryCache.Set(user+":"+key, value, int64(size.Of(value)));
		return set, nil;
	} else {
		if err := cleanCache(); err != nil {
			return false, err;
		}
		var db, ok = userCaches.Load(user);
		if !ok {
			return false, nil;
		}
		var buffer bytes.Buffer;
		if b, ok := value.([]byte); ok {
			buffer.WriteByte(0x00);
			buffer.Write(b);
		} else if s, ok := value.(string); ok {
			buffer.WriteByte(0x01);
			buffer.WriteString(s);
		} else if bl, ok := value.(bool); ok {
			buffer.WriteByte(0x02);
			if bl {
				buffer.WriteByte(0x01);
			} else {
				buffer.WriteByte(0x00);
			}
		} else {
			buffer.WriteByte(0x03);
			if err1 := gob.NewEncoder(&buffer).Encode(value); err1 != nil {
				return false, err1;
			}
		}
		if _, err1 := db.(*sql.DB).Exec(
			"insert or replace into cache (key, value, recent_use) values (?, ?, ?)",
			key, buffer.Bytes(), time.Now().UnixMilli(),
		); err1 != nil {
			return false, err1;
		}
		return true, nil;
	}
}

