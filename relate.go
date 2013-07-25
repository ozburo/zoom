package zoom

import (
	"fmt"
	"reflect"
)

type Relation struct {
	Name, Id string
}

// Return an interface{} as a result of an association lookup
func (m *Model) Fetch(relationName string) (interface{}, error) {
	// Get the id for the corresponding relation
	relation, err := findRelationByName(m.Parent, relationName)
	if err != nil {
		return nil, err
	}

	// find the result
	return FindById(relation.Name, relation.Id)
}

// TODO: return a slice of interface{} for has many relations
func (m *Model) FetchAll(relationName string) ([]interface{}, error) {
	fmt.Println("TODO: implement FetchAll")
	return nil, nil
}

func fieldIsRelational(field reflect.StructField) bool {
	n := field.Name
	if n[len(n)-2:] == "Id" || n[len(n)-3:] == "Ids" {
		tag := field.Tag
		if tag.Get("refersTo") != "" {
			return true
		}
	}
	return false
}

func relationalModelName(field reflect.StructField) string {
	return field.Tag.Get("refersTo")
}

func findRelationByName(in interface{}, relationName string) (*Relation, error) {
	// get the number of fields
	elem := reflect.ValueOf(in).Elem().Interface() // Get the actual element from the pointer
	val := reflect.ValueOf(elem)                   // for getting the actual field value
	typ := reflect.TypeOf(elem)                    // for name/type/kind information
	numFields := val.NumField()

	// we wish to iterate through the fields and find the one with the proper tags
	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		// skip the embedded Model struct
		// that's used internally and doesn't belong in redis
		if field.Name == "Model" {
			continue
		}
		// there's a special case for relational attributes
		// a.k.a. those which include Id in the name and are
		// tagged with `refersTo:*`
		if fieldIsRelational(field) {
			if field.Tag.Get("as") == relationName {
				relation := &Relation{
					Name: field.Tag.Get("refersTo"),
					Id:   val.Field(i).String(),
				}
				return relation, nil
			}
		}
	}

	return nil, nil
}

// Takes as arguments: a field, the reflect.Value of the struct which contains field,
// and the index of that field.
// Verifies that:
// 		(1) the refersTo tag is a valid model name and has been registered
// 		(2) the value of the field is a valid id (the key exists)
// A return value of nil means that the relational field is valid
// Any other return value will be the error that was caused
func validateRelationalField(field reflect.StructField, val reflect.Value, i int) error {
	fieldVal := val.Field(i)
	relateName := relationalModelName(field)
	if !alreadyRegisteredName(relateName) {
		return NewModelNameNotRegisteredError(relateName)
	}
	if fieldVal.String() != "" {
		key := relateName + ":" + fieldVal.String()
		exists, err := keyExists(key)
		if err != nil {
			return err
		}
		if !exists {
			msg := fmt.Sprintf("Couldn't find %s with id = %s\n", relateName, fieldVal.String())
			return NewKeyNotFoundError(msg)
		}
	} else {
		// fmt.Println("relational field was empty")
		return nil
	}
	return nil
}