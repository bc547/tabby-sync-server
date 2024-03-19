package configstore

import (
	"encoding/json"
	"errors"
	"go.etcd.io/bbolt"
	"time"
)

type Config struct {
	Id         string    `json:"id"`
	Name       string    `json:"name"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type ConfigStore struct {
	FileName string
	DB       *bbolt.DB
}

func Open(FileName string) (*ConfigStore, error) {
	db, err := bbolt.Open(FileName, 0644, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	cstore := &ConfigStore{
		FileName: FileName,
		DB:       db,
	}
	return cstore, nil
}

func (cstore *ConfigStore) GetSyncTokens() ([]string, error) {
	SyncTokens := []string{}
	err := cstore.DB.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bbolt.Bucket) error {
			SyncTokens = append(SyncTokens, string(name))
			return nil
		})
	})
	return SyncTokens, err
}

func (cstore *ConfigStore) CreateSyncToken(SyncToken string) error {
	err := cstore.DB.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket([]byte(SyncToken))
		return err
	})
	return err
}

func (cstore *ConfigStore) DeleteSyncToken(SyncToken string) error {
	err := cstore.DB.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket([]byte(SyncToken))
	})
	return err
}

func (cstore *ConfigStore) IsValidSyncToken(SyncToken string) (bool, error) {
	var valid = true
	err := cstore.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(SyncToken))
		if b == nil {
			valid = false
		}
		return nil
	})
	return valid, err
}

func (cstore *ConfigStore) SaveConfig(SyncToken string, Cfg *Config) error {
	key := Cfg.Id
	value, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}

	err = cstore.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(SyncToken))
		if b == nil {
			return errors.New("SaveConfig: synctoken bucket does not exist")
		}
		err = b.Put([]byte(key), value)
		return err
	})
	return err
}

func (cstore *ConfigStore) LoadConfigs(SyncToken string) ([]Config, error) {
	cfgs := []Config{} // must be [] empty slice! never nil slice

	err := cstore.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(SyncToken))
		if b == nil {
			return errors.New("LoadConfigs: synctoken bucket does not exist")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			cfg := Config{}
			err := json.Unmarshal(v, &cfg)
			if err != nil {
				return err
			}
			cfgs = append(cfgs, cfg)
		}
		return nil
	})
	return cfgs, err
}

func (cstore *ConfigStore) LoadConfig(SyncToken string, ConfigID string) (*Config, error) {
	var cfg Config
	err := cstore.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(SyncToken))
		if b == nil {
			return errors.New("LoadConfig: synctoken bucket does not exist")
		}
		v := b.Get([]byte(ConfigID))
		if v == nil {
			return errors.New("LoadConfig: ConfigID key does not exist")
		}

		err := json.Unmarshal(v, &cfg)
		return err
	})

	return &cfg, err
}

func (cstore *ConfigStore) DeleteConfig(SyncToken string, ConfigID string) error {
	err := cstore.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(SyncToken))
		if b == nil {
			return errors.New("DeleteConfig: synctoken bucket does not exist")
		}
		return b.Delete([]byte(ConfigID))
	})
	return err
}
