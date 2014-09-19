package mongo

import (
	"errors"
	"sort"
	"strings"

	"github.com/ottemo/foundation/utils"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

func (it *MongoDBCollection) convertValueToType(columnType string, value interface{}) interface{} {

	switch typedValue := value.(type) {
	case string:
		switch {
		case columnType == "string" || columnType == "text" || strings.Contains(columnType, "char"):
			return utils.InterfaceToString(value)
		case columnType == "int" || columnType == "integer":
			return utils.InterfaceToInt(value)
		case columnType == "real" || columnType == "float":
			return utils.InterfaceToFloat64(value)
		case strings.Contains(columnType, "numeric") || strings.Contains(columnType, "decimal") || columnType == "money":
			return utils.InterfaceToFloat64(value)
		case strings.Contains(columnType, "time") || strings.Contains(columnType, "date"):
			return utils.InterfaceToTime(value)
		}
	case []string:
		result := make([]interface{}, len(typedValue))
		for idx, listValue := range typedValue {
			result[idx] = it.convertValueToType(columnType, listValue)
		}
		value = result
	}

	return value
}

// converts known SQL filter operator to mongoDB one, also modifies value if needed
func (it *MongoDBCollection) getMongoOperator(columnName string, operator string, value interface{}) (string, interface{}, error) {
	operator = strings.ToLower(operator)

	columnType := it.GetColumnType(columnName)
	value = it.convertValueToType(columnType, value)

	switch operator {
	case "=":
		return "", value, nil
	case "!=", "<>":
		return "$ne", value, nil
	case ">":
		return "$gt", value, nil
	case ">=":
		return "$gte", value, nil
	case "<":
		return "$lt", value, nil
	case "<=":
		return "$lte", value, nil
	case "like":
		stringValue := utils.InterfaceToString(value)
		stringValue = strings.Replace(stringValue, "%", ".*", -1)
		return "$regex", stringValue, nil

	case "in", "nin":
		newOperator := "$" + operator

		switch typedValue := value.(type) {
		case *MongoDBCollection:
			refValue := new(bson.Raw)

			if len(typedValue.ResultAttributes) != 1 {
				typedValue.ResultAttributes = []string{"_id"}
			}

			if it.subcollections == nil {
				it.subcollections = make([]*MongoDBCollection, 0)
			}

			if it.subresults == nil {
				it.subresults = make([]*bson.Raw, 0)
			}

			it.subcollections = append(it.subcollections, typedValue)
			it.subresults = append(it.subresults, refValue)

			return newOperator, refValue, nil
		default:
			return newOperator, value, nil
		}
	}

	return "?", "?", errors.New("Unknown operator '" + operator + "'")
}

// returns filter group, creates new one if not exists
func (it *MongoDBCollection) getFilterGroup(groupName string) *T_DBFilterGroup {
	filterGroup, present := it.FilterGroups[groupName]
	if !present {
		filterGroup = &T_DBFilterGroup{Name: groupName, FilterValues: make([]bson.D, 0)}
		it.FilterGroups[groupName] = filterGroup
	}
	return filterGroup
}

// adds filter(combination of [column, operator, value]) in named filter group
func (it *MongoDBCollection) updateFilterGroup(groupName string, columnName string, operator string, value interface{}) error {

	/*if !it.HasColumn(columnName) {
		return errors.New("not existing column " + columnName)
	}*/

	// converting operator and value for mongoDB usage
	//-------------------------------------------------
	newOperator, newValue, err := it.getMongoOperator(columnName, operator, value)
	if err != nil {
		return err
	}

	if newOperator != "" {
		newValue = bson.D{bson.DocElem{Name: newOperator, Value: newValue}}
	}

	// adding filter with converted operator/value to filter group
	//------------------------------------------------------------
	newFilter := bson.D{bson.DocElem{Name: columnName, Value: newValue}}

	filterGroup := it.getFilterGroup(groupName)
	filterGroup.FilterValues = append(filterGroup.FilterValues, newFilter)

	return nil
}

// joins filters groups in one selector
func (it *MongoDBCollection) makeSelector() bson.D {

	// making sorted array of filter groups
	//-------------------------------------
	sortedFilterGroupsNames := make([]string, len(it.FilterGroups))
	idx := 0
	for groupName, _ := range it.FilterGroups {
		sortedFilterGroupsNames[idx] = groupName
		idx += 1
	}
	sort.Strings(sortedFilterGroupsNames)

	// making recursive groups injects, based on Parent field
	//-------------------------------------------------------
	topLevelGroup := &T_DBFilterGroup{Name: "", FilterValues: make([]bson.D, 0)}
	groupsStack := make([]*T_DBFilterGroup, 0)
	currentGroup := topLevelGroup

	for {

		childFound := false
		// loop over sorted filter group names
		for idx, filterGroupName := range sortedFilterGroupsNames {
			if filterGroupName == "" {
				continue
			}

			iterationFilterGroup := it.FilterGroups[filterGroupName]

			// looking for child groups, making stack on them
			//-----------------------------------------------
			if iterationFilterGroup.ParentGroup == currentGroup.Name {
				groupsStack = append(groupsStack, currentGroup)
				currentGroup = iterationFilterGroup

				// excluding group filter from our list
				sortedFilterGroupsNames[idx] = ""

				childFound = true
				break
			}
		}

		// no child found for currentGroup, collapsing stack for one level
		//----------------------------------------------------------------
		if childFound == false {

			// making document from T_DBFilterGroup before pop stack
			joinOperator := "$and"
			if currentGroup.OrSequence {
				joinOperator = "$or"
			}
			bsonDoc := bson.D{bson.DocElem{Name: joinOperator, Value: currentGroup.FilterValues}}

			// popping stack - moving level down for one level, if possible
			lastIndex := len(groupsStack) - 1
			if lastIndex >= 0 {
				currentGroup = groupsStack[lastIndex]
				groupsStack = groupsStack[0:lastIndex]
			} else {
				break
			}

			// appending top level child to parent
			currentGroup.FilterValues = append(currentGroup.FilterValues, bsonDoc)
		}
	}

	if len(topLevelGroup.FilterValues) > 0 {
		return bson.D{bson.DocElem{Name: "$and", Value: topLevelGroup.FilterValues}}
	} else {
		return bson.D{}
	}
}

// returns bson.Query struct with applied Sort, Offset, Limit parameters, and executed subqueries
func (it *MongoDBCollection) prepareQuery() *mgo.Query {
	query := it.collection.Find(it.makeSelector())

	if len(it.Sort) > 0 {
		query.Sort(it.Sort...)
	}

	if it.Offset > 0 {
		query = query.Skip(it.Offset)
	}
	if it.Limit > 0 {
		query = query.Limit(it.Limit)
	}

	for idx, subCollection := range it.subcollections {
		subCollection.prepareQuery().Distinct(subCollection.ResultAttributes[0], it.subresults[idx])
	}

	return query
}
