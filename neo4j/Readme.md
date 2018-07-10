## NeoDB

NeoDB provides a wrapper around the low level bolt library handling the boilerplate Neo4j query code for you.
Simply provide a query, parameters and a `ResultExtractorClosure`

### ResultExtractorClosure
`Query()` attempts to execute the provided cypher statement, if successful it will iterated over the result rows passing
 each to the supplied  `ResultExtractorClosure`. `ResultExtractorClosure` enables you to specify how to handle the row 
 data. `ResultExtractorClosure` accepts a `QueryResult` which contains the row `data`, `metadata` and row `index` giving 
 you access to all information returned from bolt.

##### Example: _Return a node count_ 
In this example we execute a count query of the nodes matching the specified criteria.

If more than 1 row is returned or if the data cannot be cast to the expected type we return an error. Otherwise we assign
the data to the `count` variable  **NOTE:** that `count` is declared outside of `ResultExtractorClosure`.

```go
	driver, err := bolt.NewClosableDriverPool("$bolt_url$", 1)
	if err != nil {
		// handle error
	}

	neo := NeoDB{Pool: driver}

	query := "MATCH (n:MyNode) WHERE n.SomeProperty = {propertyValue} RETURN count(*)"
	params := map[string]interface{}{"propertyValue": "xyz"}

	var count int64
	rowExtractor := func(r *QueryResult) error {
		// we expect a single result, if r.RowIndex is > 0 return an error
		if r.RowIndex != 0 {
			return errors.New("extract row result error: expected single result but was not")
		}
		var ok bool

		// Note: count is declared outside of function.
		count, ok = r.Data[0].(int64)
		if !ok {
			return errors.New("extract row result error: failed to cast result to int64")
		}
		return nil
	}

	err = neo.Query(context.Background(), query, params, rowExtractor)
	if err != nil {
		// handle error
	}
```