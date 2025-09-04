package main

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type settings struct {
	// DEFAULT,LIGHT,DARK,MONO
	Theme        string
	MaxScrolloff int
}

func getSettings() (settings, error) {
	dataFilePath, err := getDataFilePath()
	if err != nil {
		return settings{}, err
	}

	db, err := bolt.Open(dataFilePath, 0600, nil)
	if err != nil {
		return settings{}, err
	}
	defer db.Close()

	var sett settings
	err = db.View(func(tx *bolt.Tx) error {
		bucketSettings := tx.Bucket([]byte("Settings"))
		if bucketSettings == nil {
			return fmt.Errorf("bucket doesn't exist")
		}

		gobSettings := bucketSettings.Get([]byte("ALL"))
		if gobSettings == nil {
			return fmt.Errorf("key doesn't exist")
		}

		sett, err = fromGob[settings](gobSettings)
		return err
	})
	return sett, err
}

func (a *app) updateSettings(sett settings) error {
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
		bucketSettings, err := tx.CreateBucketIfNotExists([]byte("Settings"))
		if err != nil {
			return err
		}

		gobSettings, err := toGob(sett)
		if err != nil {
			return err
		}

		return bucketSettings.Put([]byte("ALL"), gobSettings)
	})
	return err
}

func initSettings() error {
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
		bucketSettings, err := tx.CreateBucketIfNotExists([]byte("Settings"))
		if err != nil {
			return err
		}

		value := bucketSettings.Get([]byte("ALL"))
		if value != nil {
			return nil
		}

		gobSettings, err := toGob(settings{
			Theme:        "DEFAULT",
			MaxScrolloff: 2,
		})
		if err != nil {
			return err
		}

		return bucketSettings.Put([]byte("ALL"), gobSettings)
	})
	return err
}
