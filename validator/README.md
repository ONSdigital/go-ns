### Validator

A Go library for the validation of HTML forms on the server side. Requires user
to define validation logic of form fields in a JSON file to allow potential
validation sharing with the client side.

#### How to use the validator

Imagine you had an HTML form as described below:

```html
<form action="/search" method="post">
  <input type="text" name="search" value="search">
  <input type="submit" value="Submit">
</form>
```

and we wanted to validate that the text entered in the search box was an email
address and it was at least 3 characters long. To use the validator we first need
to define a JSON file which follows the following structure:

```json
[
  {
    "id": "search",
    "rules": [
      {
        "name": "min_length",
        "value": 3
      },
      {
        "name": "email"
      }
    ]
  }
]
```

The id field would match the input name from the HTML, and we then define a list of
rules which correspond to a key within the function map in `rules.go`. The value
variable within in rule is an optional field which can be passed into a validation
function such as the one above for the min_length rule. To define your own custom
rules please see the section further down.

To use the validator you should create a new validator, passing in an `io.Reader`
with the contents of your JSON rules. Ensure that your rules JSON contains the
rules for all form fields which will be validated in this service, so the FormValidator
can be re used for all requests.

```go
import "github.com/ONSdigital/go-ns/validator"

file, err := os.Open("../path/rules.json")
if err != nil {
  // handle error
}
defer file.Close()

fv := validator.New(file)
```

Create a struct which matches the structure of your HTML form, defining any id
selectors by adding a schema tag to your struct fields.

```go
type Form struct {
  Search string `schema:"search"`
}
```

Inside your request handler function you will then want to call the Validate()
function and handle any errors as follows:

```go
  var frm Form
  if err := fv.Validate(req, &frm); if err != nil {
    if _, ok := err.(ErrFormValidationFailed); !ok {
      // handle this error as you would with any normal error
    } else {
      // we now know we have field validation errors so we can work out why a
      // particular field failed validation

      fieldErrs := err.(ErrFormValidationFailed).GetFieldErrors() // returns a map of []errors
      if len(fieldErrs["search"]) > 0 {
        // we now know something with search failed validation so we can handle
        // the error appropriately, perhaps by displaying an warning message to
        // the user on screen that their input was invalid

      }
    }
  }
```

#### Customising and Extending the validator

It is possible to add custom validation to this package. This can be done by
extending the RulesList map from your own code base. Imagine we had a custom rule
`food`, which would only validate if the value of an input was equal to the word
`food`. The json for this could look like:

```json
[
  {
    "id": "search",
    "rules": [
      {
        "name": "min_length",
        "value": 3
      },
      {
        "name": "email"
      },
      {
        "name": "food"
      }
    ]
  }
]
```

To handle this new rule you would need to add a new function to the rules map,
with the key "food" and its corresponding validation function, which must follow
the signature:

```go
  func(...interface{}) error
```

For example:

```go
validator.RulesList["food"] = func(vars ...interface{}) error {
  var f string
  var ok bool

  if f, ok = vars[0].(string); !ok {
    return errors.New("first parameter to food must be a string")
  }

  if f != "food" {
    return validator.FieldValidationErr{errors.New("value must equal food")}
  }
  return nil
}
```

Make sure you always handle the  `interface{}` to your preferred type and throw
an error if the conversion fails. When performing your validation, if it fails
make sure you return a `FieldValidationErr` rather than just an `error` so the
validator knows that the field has genuinely failed validation.

If you need to directly extend the validator as you known there is logic which
will be reused across services, then extend the `rules.go` file directly and,
adding your function to the RulesList map, and submit a pull request.
