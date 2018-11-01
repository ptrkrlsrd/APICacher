// Copyright © 2018 Petter Karlsrud petterkarlsrud@me.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package acache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/coreos/bolt"
	"github.com/gin-gonic/gin"
	"github.com/ptrkrlsrd/utilities/ucrypt"
)

const (
	// BoltBucketName BoltBucketName is the name of the Bolt Bucket
	BoltBucketName = "acache"
)

// Route Route
type Route struct {
	ID          string `json:"key"`
	URL         string `json:"url"`
	Alias       string `json:"alias"`
	Data        []byte `json:"data"`
	ContentType string `json:"contentType"`
}

// Store Store..
type Store struct {
	DB *bolt.DB
}

// Routes Routes
type Routes []Route

// RouteFromBytes RouteFromBytes...
func RouteFromBytes(bytes []byte) (Route, error) {
	var cacheItem Route
	err := json.Unmarshal(bytes, &cacheItem)
	if err != nil {
		return cacheItem, err
	}

	return cacheItem, nil
}

// NewCache NewCache...
func NewCache(db *bolt.DB) Store {
	return Store{DB: db}
}

//InitBucket InitBucket...
func (store *Store) InitBucket() error {
	return store.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("acache"))
		if err != nil {
			return fmt.Errorf("failed creating bucket with error: %s", err)
		}

		return nil
	})
}

//ListRoutes ListRoutes...
func (store *Store) ListRoutes() (string, error) {
	var output string
	cacheItems, err := store.GetRoutes()
	if err != nil {
		return "", err
	}

	for i, v := range cacheItems {
		output += fmt.Sprintf("%d) %s -> %s\n", i, v.URL, v.Alias)
	}

	return output, nil
}

//PrintAll PrintAll...
func (routes *Routes) PrintAll() error {
	for i, v := range *routes {
		fmt.Printf("%d) %s\n\tAlias: %s\n\tKey: %s\n\tContent-Type: %s\n", i, v.URL, v.Alias, v.ID, v.ContentType)
	}

	return nil
}

//GetRoutes GetRoutes...
func (store Store) GetRoutes() (Routes, error) {
	var cacheItems []Route

	err := store.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BoltBucketName))
		if b == nil {
			return fmt.Errorf("could not find bucket %s", BoltBucketName)
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			if cacheItem, err := RouteFromBytes(v); err == nil {
				cacheItems = append(cacheItems, cacheItem)
			} else {
				return fmt.Errorf("failed reading route from bytes: %v", err)
			}
		}

		return nil
	})

	return cacheItems, err
}

//ContainsURL ContainsURL returns true if the slice of routes contains an URL
func (routes *Routes) ContainsURL(url string) (bool, error) {
	for _, v := range *routes {
		if v.URL == url {
			return true, nil
		}
	}

	return false, nil
}

func fetchItem(url string) ([]byte, *http.Response, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	return body, res, err
}

//AddRoute AddRoute...
func (store *Store) AddRoute(url string, alias string) error {
	data, resp, err := fetchItem(url)
	key := ucrypt.MD5Hash(alias)

	cacheItem := Route{
		ID:          key,
		URL:         url,
		Alias:       alias,
		Data:        data,
		ContentType: resp.Header.Get("Content-Type"),
	}

	jsonData, err := json.Marshal(cacheItem)
	if err != nil {
		return fmt.Errorf("failed marshaling JSON: %v", err)
	}

	err = store.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BoltBucketName))
		if b == nil {
			return fmt.Errorf("failed to update the DB. Have you run 'acache init' yet?")
		}

		return b.Put([]byte(key), jsonData)
	})

	return err
}

//ClearDB ClearDB...
func (store *Store) ClearDB() error {
	return store.DB.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(BoltBucketName))
	})
}

//StartServer StartServer...
func (store *Store) StartServer(addr string) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	cacheItems, err := store.GetRoutes()

	if err != nil {
		return fmt.Errorf("could not get routes: %v", err)
	}

	for _, v := range cacheItems {
		router.GET(v.Alias, func(c *gin.Context) {
			c.Header("Content-Type", v.ContentType)
			c.String(http.StatusOK, string(v.Data))
		})
	}

	return router.Run(addr)
}
