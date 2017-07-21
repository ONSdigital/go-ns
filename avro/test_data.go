package avro

var testSchema = `{ "type": "record",
 "name": "example",
 "fields": [
    {"name": "manager", "type": "string"},
    {"name": "team_name", "type": "string"},
    {"name": "ownerOfTeam", "type": "string"},
    {"name": "kind-of-sport", "type": "string"},
    {"name": "uri", "type": "string", "default": ""},
    {"name": "has_changed_name", "type": "boolean"},
    {"name": "number_of_players", "type": "int"}
 ]
}`

var testArraySchema = `{ "type": "record",
 "name": "example",
 "fields": [
      {"name": "winning_years","type":["null",{"type":"array","items":"string"}]},
 ]
}`

var testNestedArraySchema = `{
  "type": "record",
  "name": "example",
  "fields": [
        {
            "name" : "team",
            "type" : "string"
        },
        {
            "name" : "footballers",
            "type" : {
                "type" : "array",
                "items" : {
                    "name" : "footballer",
                    "type" : "record",
                    "fields" : [
												{
														"name" : "email",
														"type" : "string"
													},
												{
														"name": "name",
														"type": "string"
												}
                    ]
                }
            }
        }
    ]
}
`

type testData struct {
	Manager         string `avro:"manager"`
	TeamName        string `avro:"team_name"`
	Owner           string `avro:"ownerOfTeam"`
	Sport           string `avro:"kind-of-sport"`
	URI             string `avro:"uri"`
	HasChangedName  bool   `avro:"has_changed_name"`
	NumberOfPlayers int32  `avro:"number_of_players"`
}

type testData1 struct {
	Manager         string `avro:"manager"`
	TeamName        string `avro:"team_name"`
	Owner           string `avro:"ownerOfTeam"`
	Sport           string `avro:"kind-of-sport"`
	URI             string `avro:"-"`
	HasChangedName  bool   `avro:"has_changed_name"`
	NumberOfPlayers int32  `avro:"number_of_players"`
}

type testData2 struct {
	Manager         string `avro:"manager"`
	URI             string `avro:"-"`
	HasChangedName  bool   `avro:"has_changed_name"`
	NumberOfPlayers int32  `avro:"number_of_players"`
	NumberOfYouths  int    `avro:"number_of_youths"`
}

type testData3 struct {
	WinningYears []string `avro:"winning_years"`
}

type testData4 struct {
	Team        string       `avro:"team"`
	Footballers []Footballer `avro:"footballers"`
}

type Footballer struct {
	Email string `avro:"email"`
	Name  string `avro:"name"`
}
