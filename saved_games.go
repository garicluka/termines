package main

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
)

func (a *app) saveGame(gInfo gameInfo, gData gameData) error {
	a.wg.Add(1)
	defer a.wg.Done()

	dataFilePath, err := getDataFilePath()
	if err != nil {
		return err
	}

	db, err := bolt.Open(dataFilePath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucketGameInfo, err := tx.CreateBucketIfNotExists([]byte("GameInfo"))
		if err != nil {
			return err
		}

		gobGameInfo, err := toGob(gInfo)
		if err != nil {
			return err
		}

		err = bucketGameInfo.Put([]byte(gInfo.Id.String()), gobGameInfo)
		if err != nil {
			return err
		}

		bucketGameData, err := tx.CreateBucketIfNotExists([]byte("GameData"))
		if err != nil {
			return err
		}

		gobGameData, err := toGob(gData)
		if err != nil {
			return err
		}

		return bucketGameData.Put([]byte(gData.Id.String()), gobGameData)
	})
	return err
}

func (a *app) loadGameInfoAndData(id uuid.UUID) (gameInfo, gameData, error) {
	a.wg.Add(1)
	defer a.wg.Done()

	dataFilePath, err := getDataFilePath()
	if err != nil {
		return gameInfo{}, gameData{}, err
	}

	db, err := bolt.Open(dataFilePath, 0600, nil)
	if err != nil {
		return gameInfo{}, gameData{}, err
	}
	defer db.Close()

	var gInfo gameInfo
	var gData gameData
	err = db.View(func(tx *bolt.Tx) error {
		bucketGameInfo := tx.Bucket([]byte("GameInfo"))
		if bucketGameInfo == nil {
			return fmt.Errorf("bucket doesn't exist")
		}

		gobGInfo := bucketGameInfo.Get([]byte(id.String()))
		if gobGInfo == nil {
			return fmt.Errorf("key doesn't exist")
		}

		gInfo, err = fromGob[gameInfo](gobGInfo)
		if err != nil {
			return err
		}

		bucketGameData := tx.Bucket([]byte("GameData"))
		if bucketGameData == nil {
			return fmt.Errorf("bucket doesn't exist")
		}

		gobGData := bucketGameData.Get([]byte(id.String()))
		if gobGData == nil {
			return fmt.Errorf("key doesn't exist")
		}

		gData, err = fromGob[gameData](gobGData)
		return err
	})
	return gInfo, gData, err
}

func (a *app) loadGameInfos(sortBy string, fieldAll bool, fieldWidth, fieldHeight, fieldMineCount int) ([]gameInfo, error) {
	a.wg.Add(1)
	defer a.wg.Done()

	dataFilePath, err := getDataFilePath()
	if err != nil {
		return []gameInfo{}, err
	}

	db, err := bolt.Open(dataFilePath, 0600, nil)
	if err != nil {
		return []gameInfo{}, err
	}
	defer db.Close()

	var infos []gameInfo
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("GameInfo"))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(_, v []byte) error {
			info, err := fromGob[gameInfo](v)
			if err != nil {
				return err
			}

			infos = append(infos, info)
			return nil
		})
	})
	if err != nil {
		return []gameInfo{}, err
	}

	switch sortBy {
	case "LATEST":
		slices.SortFunc(infos, func(a, b gameInfo) int {
			return b.CreatedAt.Compare(a.CreatedAt)
		})
	case "OLDEST":
		slices.SortFunc(infos, func(a, b gameInfo) int {
			return a.CreatedAt.Compare(b.CreatedAt)
		})
	case "BEST":
		slices.SortFunc(infos, func(a, b gameInfo) int {
			if a.Result == "WON" && b.Result == "LOST" {
				return -1
			}
			if a.Result == "LOST" && b.Result == "WON" {
				return 1
			}

			res := int(a.GameDuration.Milliseconds()) - int(b.GameDuration.Milliseconds())
			return int(res)
		})
	case "WORST":
		slices.SortFunc(infos, func(a, b gameInfo) int {
			if a.Result == "WON" && b.Result == "LOST" {
				return -1
			}
			if a.Result == "LOST" && b.Result == "WON" {
				return 1
			}

			res := int(a.GameDuration.Milliseconds()) - int(b.GameDuration.Milliseconds())
			return int(res)
		})
		slices.Reverse(infos)
	}

	if fieldAll {
		return infos, nil
	}

	filteredInfos := []gameInfo{}
	for _, v := range infos {
		if fieldWidth == v.FieldWidth &&
			fieldHeight == v.FieldHeight &&
			fieldMineCount == v.MineCount {
			filteredInfos = append(filteredInfos, v)
		}
	}

	return filteredInfos, nil
}

func (a *app) deleteGame(id uuid.UUID) error {
	a.wg.Add(1)
	defer a.wg.Done()

	dataFilePath, err := getDataFilePath()
	if err != nil {
		return err
	}

	db, err := bolt.Open(dataFilePath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("GameInfo"))
		if err != nil {
			return err
		}
		if bucket == nil {
			return nil
		}
		err = bucket.Delete([]byte(id.String()))
		if err != nil {
			return err
		}

		bucket, err = tx.CreateBucketIfNotExists([]byte("GameData"))
		if err != nil {
			return err
		}
		if bucket == nil {
			return nil
		}
		return bucket.Delete([]byte(id.String()))
	})
	return err
}
