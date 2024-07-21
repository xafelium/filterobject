package filterobject

import (
	"github.com/stretchr/testify/require"
	"github.com/xafelium/filter"
	"sort"
	"testing"
	"time"
)

type TestObject struct {
	Id          int
	TaskType    string
	Name        string
	Nicknames   []string
	HouseIds    []int
	CreatedAt   time.Time
	ChildObject *TestObject
}

func TestImplementsAllConditionTypes(t *testing.T) {
	var actual []string
	for t := range conditionEvaluators {
		actual = append(actual, t)
	}
	sort.Strings(actual)
	expected := filter.AllConditionTypes()
	sort.Strings(expected)
	require.Equal(t, expected, actual)
}

func TestFilterApplies(t *testing.T) {
	tests := []struct {
		name     string
		obj      any
		filter   filter.Condition
		expected bool
		err      error
	}{
		{
			name:     "no condition",
			obj:      TestObject{},
			filter:   nil,
			expected: true,
		},
		{
			name:     "empty where",
			obj:      TestObject{},
			filter:   filter.Where(nil),
			expected: true,
		},
		{
			name: "single matching condition",
			obj: TestObject{
				Name: "Hello World",
			},
			filter:   filter.Equals("name", "Hello World"),
			expected: true,
		},
		{
			name: "single not matching condition",
			obj: TestObject{
				Name: "foo bar",
			},
			filter:   filter.Equals("name", "Hello World"),
			expected: false,
		},
		{
			name: "where with and conditions",
			obj: TestObject{
				Name:     "Harry Potter",
				TaskType: "magic",
			},
			filter: filter.Where(
				filter.And(
					filter.Equals("taskType", "magic"),
					filter.Contains("name", "otter"),
				),
			),
			expected: true,
		},
		{
			name: "where with or conditions",
			obj: TestObject{
				Name:     "Harry Potter",
				TaskType: "magic",
			},
			filter: filter.Where(
				filter.Or(
					filter.Equals("name", "Hermine Granger"),
					filter.Equals("taskType", "magic"),
				),
			),
			expected: true,
		},
		{
			name: "where with groups",
			obj: TestObject{
				Id:        42,
				TaskType:  "magic",
				Name:      "Willi",
				Nicknames: []string{"Will", "William"},
				HouseIds:  []int{1, 2, 3},
				CreatedAt: time.Now(),
			},
			filter: filter.Where(
				filter.And(
					filter.Group(
						filter.Or(
							filter.Equals("id", 42),
							filter.Equals("name", "Hans"),
						),
					),
					filter.Group(
						filter.And(
							filter.Group(
								filter.Or(
									filter.Equals("id", 999),
									filter.Equals("name", "Berta"),
								),
							),
							filter.Group(
								filter.ArrayContains("houseIds", 2),
							),
						),
					),
				),
			),
			expected: false,
		},
		{
			name: "where with not condition",
			obj: TestObject{
				Id: 9999,
			},
			filter: filter.Where(
				filter.Not(
					filter.Equals("id", 4711),
				)),
			expected: true,
		},
		{
			name: "where with regex condition",
			obj:  TestObject{Name: "Mustermann"},
			filter: filter.Where(
				filter.Regex("name", "mann$"),
			),
			expected: true,
		},
		{
			name: "where with not regex condition",
			obj:  TestObject{Name: "Mustermann"},
			filter: filter.Where(
				filter.NotRegex("name", "frau$"),
			),
			expected: true,
		},
		{
			name: "where with not equals condition",
			obj:  TestObject{Name: "Mustermann"},
			filter: filter.Where(
				filter.NotEquals("name", "Musterfrau"),
			),
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test := test

			applies, err := FilterApplies(test.obj, test.filter)

			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, applies)
		})
	}
}

func TestApplyEquals(t *testing.T) {
	now := time.Now()
	obj := TestObject{
		Id:        1,
		TaskType:  "someType",
		CreatedAt: now,
	}
	applies, err := applyEquals(&obj, filter.Equals("id", 1))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyEquals(obj, filter.Equals("taskType", "someType"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyEquals(&obj, filter.Equals("id", 2))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyEquals(obj, filter.Equals("id", "3"))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyEquals(obj, filter.Equals("createdAt", now.Add(-1*time.Millisecond)))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyEquals(obj, filter.Equals("createdAt", now))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyEquals(obj, filter.Equals("createdAt", now.Add(1*time.Millisecond)))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyEquals(obj, filter.Equals("unknownField", 1))
	require.Error(t, err)
	require.False(t, applies)
}

func TestApplyArrayContains(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		HouseIds:  []int{1, 2, 3},
		Nicknames: []string{"foo", "bar", "baz"},
	}

	// Slice of strings
	applies, err = applyArrayContains(obj, filter.ArrayContains("nicknames", "foo"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContains(obj, filter.ArrayContains("nicknames", "bar"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContains(obj, filter.ArrayContains("nicknames", "baz"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContains(obj, filter.ArrayContains("nicknames", "test"))
	require.NoError(t, err)
	require.False(t, applies)

	// Slice of numbers
	applies, err = applyArrayContains(obj, filter.ArrayContains("houseIds", 1))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContains(obj, filter.ArrayContains("houseIds", 2))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContains(obj, filter.ArrayContains("houseIds", 3))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContains(obj, filter.ArrayContains("houseIds", 4))
	require.NoError(t, err)
	require.False(t, applies)

	// Errors
	applies, err = applyArrayContains(obj, filter.ArrayContains("unknownField", 1))
	require.Error(t, err)
	require.False(t, applies)
}

func TestApplyArrayContainsArray(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		HouseIds:  []int{1, 2, 3},
		Nicknames: []string{"foo", "bar", "baz"},
	}

	// Slice of strings
	applies, err = applyArrayContainsArray(obj, filter.ArrayContainsArray("nicknames", "foo"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContainsArray(obj, filter.ArrayContainsArray("nicknames", "bar"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContainsArray(obj, filter.ArrayContainsArray("nicknames", "baz"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContainsArray(obj, filter.ArrayContainsArray("nicknames", "test"))
	require.NoError(t, err)
	require.False(t, applies)

	// Slice of numbers
	applies, err = applyArrayContainsArray(obj, filter.ArrayContainsArray("houseIds", 1))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContainsArray(obj, filter.ArrayContainsArray("houseIds", 2))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContainsArray(obj, filter.ArrayContainsArray("houseIds", 3))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayContainsArray(obj, filter.ArrayContainsArray("houseIds", 4))
	require.NoError(t, err)
	require.False(t, applies)

	// Errors
	applies, err = applyArrayContainsArray(obj, filter.ArrayContainsArray("unknownField", 1))
	require.Error(t, err)
	require.False(t, applies)
}

func TestApplyContains(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		TaskType: "the sun is shining",
		Name:     "foo bar baz",
	}

	// String
	applies, err = applyContains(obj, filter.Contains("taskType", "SUN"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyContains(obj, filter.Contains("name", "foo"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyContains(obj, filter.Contains("name", "bar"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyContains(obj, filter.Contains("name", "baz"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyContains(obj, filter.Contains("name", "test"))
	require.NoError(t, err)
	require.False(t, applies)

	// Errors
	applies, err = applyEquals(obj, filter.Contains("unknownField", "some value"))
	require.Error(t, err)
	require.False(t, applies)
}

func TestApplyGreaterThan(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		Id:        15,
		Name:      "felix",
		CreatedAt: time.Now(),
	}

	applies, err = applyGreaterThan(obj, filter.GreaterThan("id", 14))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyGreaterThan(obj, filter.GreaterThan("id", 15))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyGreaterThan(obj, filter.GreaterThan("id", 16))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyGreaterThan(obj, filter.GreaterThan("name", "berta"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyGreaterThan(obj, filter.GreaterThan("name", "felix"))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyGreaterThan(obj, filter.GreaterThan("name", "hans"))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyGreaterThan(obj, filter.GreaterThan("createdAt", obj.CreatedAt.Add(-5*time.Hour)))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyGreaterThan(obj, filter.GreaterThan("createdAt", obj.CreatedAt))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyGreaterThan(obj, filter.GreaterThan("createdAt", obj.CreatedAt.Add(3*time.Hour)))
	require.NoError(t, err)
	require.False(t, applies)
}

func TestApplyGreaterThanOrEqual(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		Id:        15,
		Name:      "felix",
		CreatedAt: time.Now(),
	}

	applies, err = applyGreaterThanOrEqual(obj, filter.GreaterThanOrEqual("id", 14))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyGreaterThanOrEqual(obj, filter.GreaterThanOrEqual("id", 15))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyGreaterThanOrEqual(obj, filter.GreaterThanOrEqual("id", 16))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyGreaterThanOrEqual(obj, filter.GreaterThanOrEqual("name", "berta"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyGreaterThanOrEqual(obj, filter.GreaterThanOrEqual("name", "felix"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyGreaterThanOrEqual(obj, filter.GreaterThanOrEqual("name", "hans"))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyGreaterThanOrEqual(obj, filter.GreaterThanOrEqual("createdAt", obj.CreatedAt.Add(-5*time.Hour)))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyGreaterThanOrEqual(obj, filter.GreaterThanOrEqual("createdAt", obj.CreatedAt))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyGreaterThanOrEqual(obj, filter.GreaterThanOrEqual("createdAt", obj.CreatedAt.Add(3*time.Hour)))
	require.NoError(t, err)
	require.False(t, applies)
}

func TestApplyIn(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		Id:   42,
		Name: "Hans",
	}

	applies, err = applyIn(obj, filter.In("id", []int{1, 42, 99}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyIn(obj, filter.In("id", []int{1, 50, 99}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyIn(obj, filter.In("name", []string{"Berta", "Hans", "Fred"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyIn(obj, filter.In("name", []string{"Berta", "Charles", "Fred"}))
	require.NoError(t, err)
	require.False(t, applies)

	// Errors
	applies, err = applyIn(obj, filter.In("unknownField", 1))
	require.Error(t, err)
	require.False(t, applies)
}

func TestApplyLowerThan(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		Id:        15,
		Name:      "felix",
		CreatedAt: time.Now(),
	}

	applies, err = applyLowerThan(obj, filter.LowerThan("id", 14))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyLowerThan(obj, filter.LowerThan("id", 15))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyLowerThan(obj, filter.LowerThan("id", 16))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyLowerThan(obj, filter.LowerThan("name", "berta"))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyLowerThan(obj, filter.LowerThan("name", "felix"))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyLowerThan(obj, filter.LowerThan("name", "hans"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyLowerThan(obj, filter.LowerThan("createdAt", obj.CreatedAt.Add(-5*time.Hour)))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyLowerThan(obj, filter.LowerThan("createdAt", obj.CreatedAt))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyLowerThan(obj, filter.LowerThan("createdAt", obj.CreatedAt.Add(3*time.Hour)))
	require.NoError(t, err)
	require.True(t, applies)
}

func TestApplyLowerThanOrEqual(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		Id:        15,
		Name:      "felix",
		CreatedAt: time.Now(),
	}

	applies, err = applyLowerThanOrEqual(obj, filter.LowerThanOrEqual("id", 14))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyLowerThanOrEqual(obj, filter.LowerThanOrEqual("id", 15))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyLowerThanOrEqual(obj, filter.LowerThanOrEqual("id", 16))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyLowerThanOrEqual(obj, filter.LowerThanOrEqual("name", "berta"))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyLowerThanOrEqual(obj, filter.LowerThanOrEqual("name", "felix"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyLowerThanOrEqual(obj, filter.LowerThanOrEqual("name", "hans"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyLowerThanOrEqual(obj, filter.LowerThanOrEqual("createdAt", obj.CreatedAt.Add(-5*time.Hour)))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyLowerThanOrEqual(obj, filter.LowerThanOrEqual("createdAt", obj.CreatedAt))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyLowerThanOrEqual(obj, filter.LowerThanOrEqual("createdAt", obj.CreatedAt.Add(3*time.Hour)))
	require.NoError(t, err)
	require.True(t, applies)
}

func TestApplyIsNil(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		Id:          11,
		Name:        "max",
		ChildObject: nil,
	}

	applies, err = applyIsNil(obj, filter.IsNil("id"))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyIsNil(obj, filter.IsNil("name"))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyIsNil(obj, filter.IsNil("childObject"))
	require.NoError(t, err)
	require.True(t, applies)

	obj.ChildObject = new(TestObject)
	applies, err = applyIsNil(obj, filter.IsNil("childObject"))
	require.NoError(t, err)
	require.False(t, applies)
}

func TestApplyNotNil(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		Id:          11,
		Name:        "max",
		ChildObject: nil,
	}

	applies, err = applyNotNil(obj, filter.NotNil("id"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyNotNil(obj, filter.NotNil("name"))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyNotNil(obj, filter.NotNil("childObject"))
	require.NoError(t, err)
	require.False(t, applies)

	obj.ChildObject = new(TestObject)
	applies, err = applyNotNil(obj, filter.NotNil("childObject"))
	require.NoError(t, err)
	require.True(t, applies)
}

func TestApplyArraysOverlap(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		Nicknames: []string{"foo", "bar"},
		HouseIds:  []int{2, 4},
	}

	// Empty slices
	applies, err = applyArraysOverlap(TestObject{}, filter.ArraysOverlap("nicknames", nil))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(TestObject{}, filter.ArraysOverlap("nicknames", []string{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(TestObject{}, filter.ArraysOverlap("nicknames", []int{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(TestObject{}, filter.ArraysOverlap("nicknames", []string{"test"}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(TestObject{}, filter.ArraysOverlap("nicknames", []int{1}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("nicknames", nil))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("nicknames", []string{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("nicknames", []int{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("nicknames", []string{"test"}))
	require.NoError(t, err)
	require.False(t, applies)

	// String slice
	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("nicknames", []string{"foo"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("nicknames", []string{"bar", "test"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("nicknames", []string{"foo", "bar"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("nicknames", []string{"baz"}))
	require.NoError(t, err)
	require.False(t, applies)

	// Integer slice
	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("houseIds", []int{2}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("houseIds", []int{4, 6}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("houseIds", []int{2, 4}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("houseIds", []int{1}))
	require.NoError(t, err)
	require.False(t, applies)

	// Errors
	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("nicknames", []int{1}))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("houseIds", []string{"a"}))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("unknownField", nil))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("unknownField", []string{}))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyArraysOverlap(obj, filter.ArraysOverlap("unknownField", []int{}))
	require.Error(t, err)
	require.False(t, applies)
}

func TestApplyArrayIsContained(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		HouseIds:  []int{2, 4},
		Nicknames: []string{"foo", "bar"},
	}

	// Empty field slices
	applies, err = applyArrayIsContained(TestObject{}, filter.ArrayIsContained("nicknames", nil))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayIsContained(TestObject{}, filter.ArrayIsContained("nicknames", []string{"test"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayIsContained(TestObject{}, filter.ArrayIsContained("nicknames", []int{1}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayIsContained(TestObject{}, filter.ArrayIsContained("houseIds", []string{"test"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayIsContained(TestObject{}, filter.ArrayIsContained("houseIds", []int{1}))
	require.NoError(t, err)
	require.True(t, applies)

	// Empty value slices
	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("nicknames", nil))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("nicknames", []string{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("nicknames", []int{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("nicknames", [0]string{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("nicknames", [0]int{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("houseIds", nil))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("houseIds", []string{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("houseIds", []int{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("houseIds", [0]string{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("houseIds", [0]int{}))
	require.NoError(t, err)
	require.False(t, applies)

	// String slices
	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("nicknames", []string{"foo", "bar"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("nicknames", []string{"bar", "baz", "foo"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("nicknames", []string{"foo", "bar", "baz"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("nicknames", []string{"foo", "baz"}))
	require.NoError(t, err)
	require.False(t, applies)

	// Int slices
	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("houseIds", []int{2, 4}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("houseIds", []int{4, 2}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("houseIds", []int{2, 4, 6}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("houseIds", []int{1, 2, 3}))
	require.NoError(t, err)
	require.False(t, applies)

	// Errors
	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("nicknames", []int{1}))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("houseIds", []string{"a"}))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("unknownField", nil))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("unknownField", []string{}))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyArrayIsContained(obj, filter.ArrayIsContained("unknownField", []int{}))
	require.Error(t, err)
	require.False(t, applies)
}

func TestApplyOverlaps(t *testing.T) {
	var applies bool
	var err error
	obj := TestObject{
		Nicknames: []string{"foo", "bar"},
		HouseIds:  []int{2, 4},
	}

	// Empty slices
	applies, err = applyOverlaps(TestObject{}, filter.Overlaps("nicknames", nil))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(TestObject{}, filter.Overlaps("nicknames", []string{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(TestObject{}, filter.Overlaps("nicknames", []int{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(TestObject{}, filter.Overlaps("nicknames", []string{"test"}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(TestObject{}, filter.Overlaps("nicknames", []int{1}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("nicknames", nil))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("nicknames", []string{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("nicknames", []int{}))
	require.NoError(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("nicknames", []string{"test"}))
	require.NoError(t, err)
	require.False(t, applies)

	// String slice
	applies, err = applyOverlaps(obj, filter.Overlaps("nicknames", []string{"foo"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("nicknames", []string{"bar", "test"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("nicknames", []string{"foo", "bar"}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("nicknames", []string{"baz"}))
	require.NoError(t, err)
	require.False(t, applies)

	// Integer slice
	applies, err = applyOverlaps(obj, filter.Overlaps("houseIds", []int{2}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("houseIds", []int{4, 6}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("houseIds", []int{2, 4}))
	require.NoError(t, err)
	require.True(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("houseIds", []int{1}))
	require.NoError(t, err)
	require.False(t, applies)

	// Errors
	applies, err = applyOverlaps(obj, filter.Overlaps("nicknames", []int{1}))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("houseIds", []string{"a"}))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("unknownField", nil))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("unknownField", []string{}))
	require.Error(t, err)
	require.False(t, applies)

	applies, err = applyOverlaps(obj, filter.Overlaps("unknownField", []int{}))
	require.Error(t, err)
	require.False(t, applies)
}
