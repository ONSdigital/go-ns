package neo4j

import (
	"context"
	"github.com/ONSdigital/go-ns/log"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/pkg/errors"
	"io"
)

// DBPool contains the methods to control access to the Neo4J
// database pool
type DBPool interface {
	OpenPool() (bolt.Conn, error)
	Close() error
}

type NeoDB struct {
	Pool DBPool
}

type QueryResult struct {
	Data     []interface{}
	Meta     map[string]interface{}
	RowIndex int
}

type ResultExtractorClosure func(result *QueryResult) error

func (n *NeoDB) Query(ctx context.Context, cypherQuery string, params map[string]interface{}, resultExtractor ResultExtractorClosure) error {
	conn, err := n.Pool.OpenPool()
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "error opening neo4j connection"), nil)
		return err
	}
	defer conn.Close()

	rows, err := conn.QueryNeo(cypherQuery, params)
	if err != nil {
		return errors.WithMessage(err, "error executing neo4j query")
	}
	defer rows.Close()

	if err := n.ExtractResults(ctx, rows, resultExtractor); err != nil {
		return errors.WithMessage(err, "error extracting row data")
	}
	return nil
}

func (n *NeoDB) ExtractResults(ctx context.Context, rows bolt.Rows, resultExtractor ResultExtractorClosure) error {
	index := 0
	for {
		data, meta, err := rows.NextNeo()
		if err != nil {
			if err == io.EOF {
				log.InfoCtx(ctx, "ExtractResults: reached end of result rows", nil)
				return nil
			} else {
				log.ErrorCtx(ctx, errors.WithMessage(err, "row error, breaking loop"), nil)
				return err
			}
		}
		if err := resultExtractor(&QueryResult{Data: data, Meta: meta, RowIndex: index}); err != nil {
			return err
		}
		index++
	}
	return nil
}
