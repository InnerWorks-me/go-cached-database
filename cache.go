package cacheddatabase

import (
	"context"
	"encoding/json"
)

func WithCache[T, U any](cda *Adapter[T], key string, retrieveFn func() (U, error)) (U, error) {
	ctx := context.Background()

	// Try and get entry from Redis
	getRes := cda.redisClient.Get(ctx, key)
	if getRes.Err() != nil {
		// Generic Redis error. Fetch from DB
		return retrieveAndCache(cda, key, retrieveFn)
	}

	objStr, err := getRes.Result()
	if err == nil {
		// Object found in Redis. Return cached element
		var redisResult U
		err := json.Unmarshal([]byte(objStr), &redisResult)
		if err != nil {
			// Unmarshal error. Fetch from DB
			return retrieveAndCache(cda, key, retrieveFn)
		}
		return redisResult, nil
	} else {
		// Obj not found in redis
		return retrieveAndCache(cda, key, retrieveFn)
	}
}

func retrieveAndCache[T, U any](cda *Adapter[T], key string, retrieveFn func() (U, error)) (U, error) {
	ctx := context.Background()

	// Retrieve record from DB
	resultFromDb, err := retrieveFn()
	if err != nil {
		return resultFromDb, err
	}

	// Asynchronously add entry to cache
	go func() {
		asStr, err := json.Marshal(resultFromDb)
		if err == nil {
			go cda.redisClient.Set(ctx, key, asStr, cda.cacheTTL)
		}

	}()

	return resultFromDb, nil
}
