package filterobject

import (
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/xafelium/filter"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type ConditionEvaluator func(obj any, condition filter.Condition) (bool, error)

var (
	conditionEvaluators = make(map[string]ConditionEvaluator)
)

func init() {
	conditionEvaluators[filter.AndConditionType] = applyAnd
	conditionEvaluators[filter.ArrayContainsConditionType] = applyArrayContains
	conditionEvaluators[filter.ArrayContainsArrayConditionType] = applyArrayContainsArray
	conditionEvaluators[filter.ArrayContainsConditionType] = applyArrayContains
	conditionEvaluators[filter.ArrayIsContainedConditionType] = applyArrayIsContained
	conditionEvaluators[filter.ArraysOverlapConditionType] = applyArraysOverlap
	conditionEvaluators[filter.ContainsConditionType] = applyContains
	conditionEvaluators[filter.EqualsConditionType] = applyEquals
	conditionEvaluators[filter.GreaterThanConditionType] = applyGreaterThan
	conditionEvaluators[filter.GreaterThanOrEqualConditionType] = applyGreaterThanOrEqual
	conditionEvaluators[filter.GroupConditionType] = applyGroup
	conditionEvaluators[filter.InConditionType] = applyIn
	conditionEvaluators[filter.LowerThanConditionType] = applyLowerThan
	conditionEvaluators[filter.LowerThanOrEqualConditionType] = applyLowerThanOrEqual
	conditionEvaluators[filter.IsNilConditionType] = applyIsNil
	conditionEvaluators[filter.NotConditionType] = applyNot
	conditionEvaluators[filter.NotEqualsConditionType] = applyNotEquals
	conditionEvaluators[filter.NotNilConditionType] = applyNotNil
	conditionEvaluators[filter.NotRegexConditionType] = applyNotRegex
	conditionEvaluators[filter.OrConditionType] = applyOr
	conditionEvaluators[filter.OverlapsConditionType] = applyOverlaps
	conditionEvaluators[filter.RegexConditionType] = applyRegex
	conditionEvaluators[filter.WhereConditionType] = applyWhere
}

func FilterApplies(obj any, condition filter.Condition) (bool, error) {
	if condition == nil {
		return true, nil
	}
	evaluate, ok := conditionEvaluators[condition.Type()]
	if !ok {
		return false, fmt.Errorf(fmt.Sprintf("unknown condition: %s", condition.Type()))
	}
	return evaluate(obj, condition)
}

func applyWhere(obj any, condition filter.Condition) (bool, error) {
	whereCondition, ok := condition.(*filter.WhereCondition)
	if !ok {
		return false, fmt.Errorf("condition is no WhereCondition")
	}
	if whereCondition.Condition == nil {
		return true, nil
	}
	return FilterApplies(obj, whereCondition.Condition)
}

func applyAnd(obj any, condition filter.Condition) (bool, error) {
	andCondition, ok := condition.(*filter.AndCondition)
	if !ok {
		return false, fmt.Errorf("conditio is no AndCondition")
	}
	if len(andCondition.Conditions) < 2 {
		return false, fmt.Errorf("AND condition must have at least two conditions")
	}

	for _, c := range andCondition.Conditions {
		applies, err := FilterApplies(obj, c)
		if err != nil {
			return false, err
		}
		if !applies {
			return false, err
		}
	}
	return true, nil
}

func applyOr(obj any, condition filter.Condition) (bool, error) {
	orCondition, ok := condition.(*filter.OrCondition)
	if !ok {
		return false, fmt.Errorf("conditio is no OrCondition")
	}
	if len(orCondition.Conditions) < 2 {
		return false, fmt.Errorf("OR condition must have at least two conditions")
	}

	for _, c := range orCondition.Conditions {
		applies, err := FilterApplies(obj, c)
		if err != nil {
			return false, err
		}
		if applies {
			return true, err
		}
	}
	return false, nil
}

func applyGroup(obj any, condition filter.Condition) (bool, error) {
	groupCondition, ok := condition.(*filter.GroupCondition)
	if !ok {
		return false, fmt.Errorf("conditio is no GroupCondition")
	}
	return FilterApplies(obj, groupCondition.Condition)
}

func applyArrayContains(obj any, condition filter.Condition) (bool, error) {
	containsCondition, ok := condition.(*filter.ArrayContainsCondition)
	if !ok {
		return false, fmt.Errorf("condition is no ArrayContainsCondition")
	}
	field, err := getField(obj, containsCondition.Field)
	if err != nil {
		return false, err
	}
	if field.Kind() == reflect.String {
		return applyContains(obj, filter.Contains(containsCondition.Field, fmt.Sprintf("%s", containsCondition.Value)))
	}
	if field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
		return false, fmt.Errorf("field must be of type slice/array but is of type %s", field.Kind())
	}
	for i := 0; i < field.Len(); i++ {
		value := field.Index(i)
		if value.Interface() == containsCondition.Value {
			return true, nil
		}
	}
	return false, nil
}

func applyArrayContainsArray(obj any, condition filter.Condition) (bool, error) {
	c, ok := condition.(*filter.ArrayContainsArrayCondition)
	if !ok {
		return false, fmt.Errorf("condition is no ArrayContainsArrayCondition")
	}
	return applyArrayContains(obj, filter.ArrayContains(c.Field, c.Value))
}

func applyContains(obj any, condition filter.Condition) (bool, error) {
	containsCondition, ok := condition.(*filter.ContainsCondition)
	if !ok {
		return false, fmt.Errorf("condition is no ContainsCondition")
	}
	field, err := getField(obj, containsCondition.Field)
	if err != nil {
		return false, err
	}

	return strings.Index(
		strings.ToLower(fmt.Sprintf("%s", field.Interface())),
		strings.ToLower(fmt.Sprintf("%s", containsCondition.Value)),
	) != -1, nil
}

func applyEquals(obj any, condition filter.Condition) (bool, error) {
	equalsCondition, ok := condition.(*filter.EqualsCondition)
	if !ok {
		return false, fmt.Errorf("condition is no EqualsCondition")
	}
	field, err := getField(obj, equalsCondition.Field)
	if err != nil {
		return false, err
	}

	return field.Interface() == equalsCondition.Value, nil
}

func applyNotEquals(obj any, condition filter.Condition) (bool, error) {
	notEqualsCondition, ok := condition.(*filter.NotEqualsCondition)
	if !ok {
		return false, fmt.Errorf("condition is no NotEqualsCondition")
	}
	applies, err := FilterApplies(obj, filter.Equals(notEqualsCondition.Field, notEqualsCondition.Value))
	if err != nil {
		return false, err
	}
	return !applies, nil
}

func applyGreaterThan(obj any, condition filter.Condition) (bool, error) {
	gtCondition, ok := condition.(*filter.GreaterThanCondition)
	if !ok {
		return false, fmt.Errorf("condition is no GreaterThanCondition")
	}
	field, err := getField(obj, gtCondition.Field)
	if err != nil {
		return false, err
	}
	value := reflect.ValueOf(gtCondition.Value)
	if field.CanInt() && value.CanInt() {
		return field.Int() > value.Int(), nil
	}
	if field.CanFloat() && value.CanFloat() {
		return field.Float() > value.Float(), nil
	}
	if field.CanUint() && value.CanUint() {
		return field.Uint() > value.Uint(), nil
	}
	if field.Kind() == reflect.String && field.Kind() == reflect.String {
		return field.String() > value.String(), nil
	}
	if reflect.TypeOf(field.Interface()).String() == "time.Time" &&
		reflect.TypeOf(value.Interface()).String() == "time.Time" {
		fieldValue := field.Interface().(time.Time)
		actualValue := value.Interface().(time.Time)
		return fieldValue.After(actualValue), nil
	}
	return false, fmt.Errorf("cannot compare variables of type %s and %s",
		field.Kind(), value.Kind())
}

func applyGreaterThanOrEqual(obj any, condition filter.Condition) (bool, error) {
	gteCondition, ok := condition.(*filter.GreaterThanOrEqualCondition)
	if !ok {
		return false, fmt.Errorf("condition is no GreaterThanOrEqualCondition")
	}
	isEq, err := applyEquals(obj, filter.Equals(gteCondition.Field, gteCondition.Value))
	if err != nil {
		return false, err
	}
	isGt, err := applyGreaterThan(obj, filter.GreaterThan(gteCondition.Field, gteCondition.Value))
	if err != nil {
		return false, err
	}
	return isEq || isGt, nil
}

func applyIn(obj any, condition filter.Condition) (bool, error) {
	inCondition, ok := condition.(*filter.InCondition)
	if !ok {
		return false, fmt.Errorf("condition is no InCondition")
	}
	field, err := getField(obj, inCondition.Field)
	if err != nil {
		return false, err
	}
	valueType := reflect.ValueOf(inCondition.Value)
	if valueType.Kind() != reflect.Slice && valueType.Kind() != reflect.Array {
		return false, fmt.Errorf("field must be of type slice/array but is of type %s", valueType.Kind())
	}
	for i := 0; i < valueType.Len(); i++ {
		value := valueType.Index(i)
		if value.Interface() == field.Interface() {
			return true, nil
		}
	}
	return false, nil
}

func applyLowerThan(obj any, condition filter.Condition) (bool, error) {
	ltCondition, ok := condition.(*filter.LowerThanCondition)
	if !ok {
		return false, fmt.Errorf("condition is no LowerThanCondition")
	}
	field, err := getField(obj, ltCondition.Field)
	if err != nil {
		return false, err
	}
	value := reflect.ValueOf(ltCondition.Value)
	if field.CanInt() && value.CanInt() {
		return field.Int() < value.Int(), nil
	}
	if field.CanFloat() && value.CanFloat() {
		return field.Float() < value.Float(), nil
	}
	if field.CanUint() && value.CanUint() {
		return field.Uint() < value.Uint(), nil
	}
	if field.Kind() == reflect.String && field.Kind() == reflect.String {
		return field.String() < value.String(), nil
	}
	if reflect.TypeOf(field.Interface()).String() == "time.Time" &&
		reflect.TypeOf(value.Interface()).String() == "time.Time" {
		fieldValue := field.Interface().(time.Time)
		actualValue := value.Interface().(time.Time)
		return fieldValue.Before(actualValue), nil
	}
	return false, fmt.Errorf("cannot compare variables of type %s and %s",
		field.Kind(), value.Kind())
}

func applyLowerThanOrEqual(obj any, condition filter.Condition) (bool, error) {
	lteCondition, ok := condition.(*filter.LowerThanOrEqualCondition)
	if !ok {
		return false, fmt.Errorf("condition is no LowerThanOrEqualCondition")
	}
	isEq, err := applyEquals(obj, filter.Equals(lteCondition.Field, lteCondition.Value))
	if err != nil {
		return false, err
	}
	isLt, err := applyLowerThan(obj, filter.LowerThan(lteCondition.Field, lteCondition.Value))
	if err != nil {
		return false, err
	}
	return isEq || isLt, nil
}

func applyIsNil(obj any, condition filter.Condition) (bool, error) {
	isNilCondition, ok := condition.(*filter.IsNilCondition)
	if !ok {
		return false, fmt.Errorf("condition is no IsNilCondition")
	}
	field, err := getField(obj, isNilCondition.Field)
	if err != nil {
		return false, err
	}
	switch field.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return field.IsNil(), nil
	default:
		break
	}
	return false, nil
}

func applyNot(obj any, condition filter.Condition) (bool, error) {
	notCondition, ok := condition.(*filter.NotCondition)
	if !ok {
		return false, fmt.Errorf("condition is no NotCondition")
	}

	applies, err := FilterApplies(obj, notCondition.Condition)
	return !applies, err
}

func applyNotNil(obj any, condition filter.Condition) (bool, error) {
	notNilCondition, ok := condition.(*filter.NotNilCondition)
	if !ok {
		return false, fmt.Errorf("condition is no NotNilCondition")
	}
	field, err := getField(obj, notNilCondition.Field)
	if err != nil {
		return false, err
	}
	switch field.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return !field.IsNil(), nil
	default:
		break
	}
	return true, nil
}

func applyArraysOverlap(obj any, condition filter.Condition) (bool, error) {
	overlapsCondition, ok := condition.(*filter.ArraysOverlapCondition)
	if !ok {
		return false, errors.New("condition is no ArraysOverlapCondition")
	}
	field, err := getField(obj, overlapsCondition.Field)
	if err != nil {
		return false, err
	}
	if field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
		return false, fmt.Errorf("field must be of type slice/array but is of type %s", field.Kind())
	}
	if field.Len() == 0 {
		return false, err
	}
	if overlapsCondition.Value == nil {
		return false, err
	}

	v := reflect.ValueOf(overlapsCondition.Value)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return false, fmt.Errorf("value must be of type slice/array but is of type %s", field.Kind())
	}
	if v.Len() == 0 {
		return false, nil
	}

	fieldElemType := field.Type().Elem()
	valueElemType := v.Type().Elem()
	if fieldElemType != valueElemType {
		return false, fmt.Errorf("type mismatch: cannot compare %s (field) and %s (value)", fieldElemType.String(), valueElemType.String())
	}

	valueMap := make(map[any]struct{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		valueMap[v.Index(i).Interface()] = struct{}{}
	}

	for i := 0; i < field.Len(); i++ {
		if _, found := valueMap[field.Index(i).Interface()]; found {
			return true, nil
		}
	}
	return false, nil
}

func applyOverlaps(obj any, condition filter.Condition) (bool, error) {
	c, ok := condition.(*filter.OverlapsCondition)
	if !ok {
		return false, errors.New("condition is no ArraysOverlapCondition")
	}
	return applyArraysOverlap(obj, filter.ArraysOverlap(c.Field, c.Value))
}

func getField(obj any, name string) (reflect.Value, error) {
	var v reflect.Value
	kind := reflect.ValueOf(obj).Kind()
	switch kind {
	case reflect.Ptr:
		v = reflect.ValueOf(obj).Elem()
	case reflect.Struct:
		v = reflect.ValueOf(obj)
	default:
		break
	}
	if !v.IsValid() {
		return reflect.Value{}, fmt.Errorf("invalid object type: %s", kind)
	}

	var field reflect.Value
	fieldName := strcase.ToCamel(name)
	for i := 0; i < v.NumField(); i++ {
		if strcase.ToCamel(v.Type().Field(i).Name) == fieldName {
			field = v.Field(i)
			break
		}
	}
	if !field.IsValid() {
		return field, fmt.Errorf("field '%s' was not found on object", name)
	}
	return field, nil
}

func applyArrayIsContained(obj any, condition filter.Condition) (bool, error) {
	containsCondition, ok := condition.(*filter.ArrayIsContainedCondition)
	if !ok {
		return false, errors.New("condition is no ArrayIsContainedCondition")
	}
	field, err := getField(obj, containsCondition.Field)
	if err != nil {
		return false, err
	}
	if field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
		return false, fmt.Errorf("field must be of type slice/array but is of type %s", field.Kind())
	}
	if field.Len() == 0 {
		return true, nil
	}
	if containsCondition.Value == nil {
		return false, nil
	}

	v := reflect.ValueOf(containsCondition.Value)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return false, fmt.Errorf("value must be of type slice/array but is of type %s", field.Kind())
	}
	if v.Len() == 0 {
		return false, nil
	}

	fieldElemType := field.Type().Elem()
	valueElemType := v.Type().Elem()
	if fieldElemType != valueElemType {
		return false, fmt.Errorf("type mismatch: cannot compare %s (field) and %s (value)", fieldElemType.String(), valueElemType.String())
	}

	valueMap := make(map[any]struct{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		valueMap[v.Index(i).Interface()] = struct{}{}
	}

	for i := 0; i < field.Len(); i++ {
		if _, found := valueMap[field.Index(i).Interface()]; !found {
			return false, nil
		}
	}
	return true, nil
}

func applyRegex(obj any, condition filter.Condition) (bool, error) {
	regexCondition, ok := condition.(*filter.RegexCondition)
	if !ok {
		return false, fmt.Errorf("condition is no RegexCondition")
	}
	field, err := getField(obj, regexCondition.Field)
	if err != nil {
		return false, err
	}

	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}
	return regexp.MatchString(regexCondition.Expression, field.String())
}

func applyNotRegex(obj any, condition filter.Condition) (bool, error) {
	notRegexCondition, ok := condition.(*filter.NotRegexCondition)
	if !ok {
		return false, fmt.Errorf("condition is no NotRegexCondition")
	}
	applies, err := FilterApplies(obj, filter.Regex(notRegexCondition.Field, notRegexCondition.Expression))
	if err != nil {
		return false, err
	}
	return !applies, nil
}
