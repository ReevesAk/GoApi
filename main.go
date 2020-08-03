package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	bolt "go.etcd.io/bbolt"
)

type Database struct {
	database bolt.DB
}

//User details
type PostRequest struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

var request []PostRequest

func openDatabase() (*bolt.DB, error) {

	database, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	return database, err
}

// RegisterPost allows user to register.
func registerPost(write http.ResponseWriter, read *http.Request) {
	boltDb, err := openDatabase()
	handleError(err)
	write.Header().Set("Content-Type", "application/json")
	var newRequest PostRequest
	json.NewDecoder(read.Body).Decode(&newRequest)
	newRequest.Id = strconv.Itoa(len(request) + 1)
	request = append(request, newRequest)
	if err := json.NewEncoder(write).Encode(request); err != nil {
		fmt.Println(err)
	}
	byteToString, _ := json.Marshal(request)

	key := []byte(newRequest.Email)
	value := byteToString

	if err := boltDb.Update(func(tx *bolt.Tx) error {
		save, err := tx.CreateBucket(key)
		if err != nil {
			return err
		}
		if err := save.Put([]byte(newRequest.Email), []byte(byteToString)); err != nil {
			return err
		}
		fmt.Printf("Data received by client: %s\n", value)

		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

// LoginPost checks if a user exists and allows
// existing user access.
func loginPost(write http.ResponseWriter, read *http.Request) {
	boltDb, _ := openDatabase()
	write.Header().Set("Content-Type", "application/json")
	var userData PostRequest
	json.NewDecoder(read.Body).Decode(&userData)

	if err := boltDb.View(func(tx *bolt.Tx) error {
		value := tx.Bucket([]byte(userData.Email))
		if value == nil {
			log.Fatal("user does not exist")
		} else {
			fmt.Println("Login was successful")
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

}

func main() {

	db := Database{}

	address := ":3030"
	router := mux.NewRouter()

	router.HandleFunc("/", registerPost).Methods("POST")
	router.HandleFunc("/{email}", loginPost).Methods("POST")

	log.Fatal(http.ListenAndServe(address, router))
	db.database.Close()
}

func handleError(err error) {
	if err != nil {
		log.Printf("cause of error: %s\n", err)
	}
	return
}
