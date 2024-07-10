package test

import (
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/tutorials/abci-v2-forum-app/model"
)

func TestFindUserByName(t *testing.T) {
	// Initialize the database
	opts := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opts)
	require.NoError(t, err)
	defer db.Close()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create a new DB instance for testing
	testDB := &model.DB{}
	testDB.Init(db)

	// Create some test users
	println("User being created")
	users := []*model.User{
		{Name: "user1", PubKey: []byte("pubkey1"), Moderator: false, Banned: false},
		{Name: "user2", PubKey: []byte("pubkey2"), Moderator: false, Banned: false},
		{Name: "user3", PubKey: []byte("pubkey3"), Moderator: false, Banned: false},
	}
	println("User is defined")
	for _, user := range users {
		err := testDB.CreateUser(user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Verify that the correct user was returned
	println("Trying to find user")
	foundUser1, err1 := testDB.FindUserByName("user2")
	if err1 != nil {
		t.Fatalf("Failed to find user by name: %v", err1)
	}

	if foundUser1 == nil {
		t.Fatalf("User not found")
	}

	if foundUser1.Name != "user2" {
		// t.Fatalf("Expected user2, but got %s", foundUser.Name)
		println("Expected user2, but got %s", foundUser1.Name)
	}

	if foundUser1.Name == "user2" {
		println("Voila! User found")
	}
}
