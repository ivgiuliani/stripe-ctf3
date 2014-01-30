package command

import (
	"fmt"
	"github.com/goraft/raft"
	"stripe-ctf.com/sqlcluster/sql"
)

type QueryCommand struct {
	Create       bool   `json:c`
	Name         string `json:n`
	FavoriteWord string `json:w`
	FriendCount  int    `json:f`
}

// Creates a new write command.
func NewQueryCommand(createQuery bool, name, word string, count int) *QueryCommand {
	return &QueryCommand{
		Create:       createQuery,
		Name:         name,
		FavoriteWord: word,
		FriendCount:  count,
	}
}

// The name of the command in the log.
func (c *QueryCommand) CommandName() string {
	return "sql"
}

// Writes a value to a key.
func (c *QueryCommand) Apply(server raft.Server) (interface{}, error) {
	db := server.Context().(*sql.SQL)

	output, err := db.CacheExec(c.Create, c.Name, c.FriendCount, c.FavoriteWord)
	if err != nil {
		return nil, err
	}

	formatted := fmt.Sprintf("SequenceNumber: %d\n%s", output.SequenceNumber, output.Output)

	return []byte(formatted), nil
}
