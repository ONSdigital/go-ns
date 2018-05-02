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
    {"name": "number_of_players", "type": "int"},
    {"name": "pay_per_week", "type": "long"}
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

var nestedObjectSchema = `{
	"type": "record",
	"name": "nested-object-example",
	"fields": [
		{
			"name": "team",
			"type": "string"
		},
		{
			"name": "footballer",
			"type": {
				"name": "footballer-name",
				"type": "record",
				"fields": [
					{
						"name": "surname",
						"type": "string",
						"default": ""
					},
					{
						"name": "forename",
						"type": "string",
						"default": ""
					},
                                	{
                                        	"name": "aka",
                                                "type": {
                                                        "type": "map",
                                                        "values": "string"
                                		}
                                        }
				]
			}
		},
		{
			"name": "aka",
			"type": {
					"type": "map",
					"values": "string"
				}
		},
		{
			"name": "silverware",
			"default": null,
			"type": [
				"null",
				{
					"type": "map",
					"values": "string"
				}
			]
		},
		{
			"name": "stats",
			"type": [
				"int",
				"null"
			]
		}
	]
}`

// NestedTestData represents an object nested within an object
type NestedTestData struct {
	Team       string            `avro:"team"`
	Footballer FootballerName    `avro:"footballer"`
	Stats      int32             `avro:"stats"`
	AKA        map[string]string `avro:"aka"`
	Silverware map[string]string `avro:"silverware"`
}

// FootballerName represents an object containing the footballers name
type FootballerName struct {
	Surname  string            `avro:"surname"`
	Forename string            `avro:"forename"`
	AKA      map[string]string `avro:"aka"`
}

type testData struct {
	Manager         string `avro:"manager"`
	TeamName        string `avro:"team_name"`
	Owner           string `avro:"ownerOfTeam"`
	Sport           string `avro:"kind-of-sport"`
	URI             string `avro:"uri"`
	HasChangedName  bool   `avro:"has_changed_name"`
	NumberOfPlayers int32  `avro:"number_of_players"`
	PayPerWeek      int64  `avro:"pay_per_week"`
}

type testData1 struct {
	Manager         string `avro:"manager"`
	TeamName        string `avro:"team_name"`
	Owner           string `avro:"ownerOfTeam"`
	Sport           string `avro:"kind-of-sport"`
	URI             string `avro:"-"`
	HasChangedName  bool   `avro:"has_changed_name"`
	NumberOfPlayers int32  `avro:"number_of_players"`
	PayPerWeek      int64  `avro:"pay_per_week"`
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

type testData5 struct {
	WinningYears []string `avro:"winning_years"`
}

// Footballer represents the details of a footballer
type Footballer struct {
	Email string `avro:"email"`
	Name  string `avro:"name"`
}
