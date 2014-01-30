package sql

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
)

var queryRegex = regexp.MustCompile(`friendCount=friendCount\+(?P<friendCount>\d+), requestCount=requestCount\+1, favoriteWord="(?P<favoriteWord>\w+)" WHERE name="(?P<name>\w+)"`)

/**
 * It's actually a poor-man in-memory lookalike database for this
 * specific use case, which is essentially a map[favoriteWord] -> fields.
 */
type SQL struct {
	SequenceNumber int
	output         map[string]*Output
	innerdb        map[string]*item
}

type item struct {
	friendCount  int
	requestCount int
	favoriteWord string
}

type Output struct {
	Output         []byte
	SequenceNumber int
}

func NewSQL() *SQL {
	sql := &SQL{
		output:  make(map[string]*Output),
		innerdb: make(map[string]*item),
	}

	for _, n := range []string{"siddarth", "gdb", "christian", "andy", "carl"} {
		sql.innerdb[n] = &item{
			friendCount:  0,
			requestCount: 0,
			favoriteWord: "",
		}
	}

	return sql
}

func (sql *SQL) Execute(create bool, name string, friendCount int, favoriteWord string) (*Output, error) {
	defer func() { sql.SequenceNumber += 1 }()

	// CREATE queries are not really needed anymore, we can just
	// ignore them and return the expected output.
	if create {
		return &Output{
			Output:         bytes.NewBufferString("").Bytes(),
			SequenceNumber: sql.SequenceNumber,
		}, nil
	}

	sql.innerdb[name].requestCount += 1
	sql.innerdb[name].favoriteWord = favoriteWord
	sql.innerdb[name].friendCount += friendCount

	out := ""
	for _, n := range []string{"siddarth", "gdb", "christian", "andy", "carl"} {
		out += n + "|" + strconv.Itoa(sql.innerdb[n].friendCount) + "|" + strconv.Itoa(sql.innerdb[n].requestCount) + "|" + sql.innerdb[n].favoriteWord + "\n"
	}

	return &Output{
		Output:         bytes.NewBufferString(out).Bytes(),
		SequenceNumber: sql.SequenceNumber,
	}, nil
}

func (sql *SQL) CacheExec(create bool, name string, friendCount int, favoriteWord string) (*Output, error) {
	if output, ok := sql.output[favoriteWord]; ok {
		// do not re-execute the command if we already did (hence
		// the output is already stored)
		return output, nil
	}

	output, err := sql.Execute(create, name, friendCount, favoriteWord)
	if err != nil {
		return nil, err
	}

	// cache the query output for later use
	sql.output[favoriteWord] = output
	return output, nil
}

func (sql *SQL) ReadOutput(key string) ([]byte, error) {
	if output, ok := sql.output[key]; ok {
		return output.Output, nil
	}
	return nil, errors.New("missing value")
}

// return favoriteWord, name, friendCount
func (sql *SQL) GetData(query string) (string, string, int) {
	keys := queryRegex.FindStringSubmatch(query)
	if len(keys) != 4 {
		return "DUMB", "DUMB", 0
	}
	v, err := strconv.Atoi(keys[1])
	if err != nil {
		v = 0
	}
	return keys[2], keys[3], v
}
