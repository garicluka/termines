package main

import (
	"encoding/binary"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type settings struct {
	// DEFAULT,LIGHT,DARK,MONO
	Theme        string
	MaxScrolloff uint64
	// ON, OFF
	Icons string
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
			return fmt.Errorf("Error: Bucket doesn't exist")
		}

		themeValue := bucketSettings.Get([]byte("Theme"))
		if themeValue == nil {
			return fmt.Errorf("Error: Theme key does not exist.")
		}
		sett.Theme = string(themeValue)

		maxScrolloffValue := bucketSettings.Get([]byte("MaxScrolloff"))
		if maxScrolloffValue == nil {
			return fmt.Errorf("Error: MaxScrolloff key does not exist.")
		}
		sett.MaxScrolloff = binary.BigEndian.Uint64(maxScrolloffValue)

		iconsValue := bucketSettings.Get([]byte("Icons"))
		if iconsValue == nil {
			return fmt.Errorf("Error: Icons key does not exist.")
		}
		sett.Icons = string(iconsValue)

		return nil
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

		err = bucketSettings.Put([]byte("Theme"), []byte(sett.Theme))
		if err != nil {
			return err
		}

		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], sett.MaxScrolloff)
		err = bucketSettings.Put([]byte("MaxScrolloff"), buf[:])
		if err != nil {
			return err
		}

		err = bucketSettings.Put([]byte("Icons"), []byte(sett.Icons))
		if err != nil {
			return err
		}

		return nil
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

		themeValue := bucketSettings.Get([]byte("Theme"))
		if themeValue == nil {
			err = bucketSettings.Put([]byte("Theme"), []byte("DEFAULT"))
			if err != nil {
				return err
			}
		}

		maxScrolloffValue := bucketSettings.Get([]byte("MaxScrolloff"))
		if maxScrolloffValue == nil {
			var buf [8]byte
			binary.BigEndian.PutUint64(buf[:], 2)
			err = bucketSettings.Put([]byte("MaxScrolloff"), buf[:])
			if err != nil {
				return err
			}
		}

		iconsValue := bucketSettings.Get([]byte("Icons"))
		if iconsValue == nil {
			err = bucketSettings.Put([]byte("Icons"), []byte("OFF"))
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
