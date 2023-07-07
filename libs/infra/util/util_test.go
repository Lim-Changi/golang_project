package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetTypeName(t *testing.T) {
	assert.Equal(t, "T", GetTypeName(&testing.T{}))
	assert.Equal(t, "T", GetTypeName((*testing.T)(nil)))

	var integer int = 5
	assert.Equal(t, "int", GetTypeName(&integer))

	str := "string!"
	assert.Equal(t, "string", GetTypeName(&str))
}

func TestBoolPointer(t *testing.T) {
	assert.True(t, *BoolPointer(true))
	assert.False(t, *BoolPointer(false))
}

func TestToLowerCaseStringArray(t *testing.T) {
	mixedCaseStringArray := []string{
		"aBc",
		"DeF",
		"GHi",
	}
	lowerCaseStringArray := []string{
		"abc",
		"def",
		"ghi",
	}
	ToLowerCaseStringArray(mixedCaseStringArray)
	assert.Len(t, mixedCaseStringArray, len(lowerCaseStringArray))
	for i, v := range mixedCaseStringArray {
		assert.Equal(t, lowerCaseStringArray[i], v)
	}
}

func TestSplitCSVWithoutEmptyString(t *testing.T) {
	//
	assert.Len(t, SplitCSVWithoutEmptyString(",,,,"), 0)
	//
	csv := ",,aaa,,bbb,,,ccc,,ddd,eee,"
	expected := []string{
		"aaa",
		"bbb",
		"ccc",
		"ddd",
		"eee",
	}
	outArr := SplitCSVWithoutEmptyString(csv)
	assert.Len(t, outArr, len(expected))
	for i, v := range expected {
		assert.Equal(t, v, outArr[i])
	}
}

func TestSplitWithoutEmptyString(t *testing.T) {
	delimiters := []string{"!", "@", "#", "$", "^", "&", "*", "-", "|", "."}
	for _, d := range delimiters {
		//
		assert.Len(t, SplitWithoutEmptyString(fmt.Sprintf("%s%s%s%s", d, d, d, d), d), 0)
		//
		str := fmt.Sprintf(
			"%s%saaa%s%sbbb%s%s%sccc%s%sddd%seee%s",
			d, d,
			d, d,
			d, d, d,
			d, d,
			d,
			d)
		expected := []string{
			"aaa",
			"bbb",
			"ccc",
			"ddd",
			"eee",
		}
		outArr := SplitWithoutEmptyString(str, d)
		assert.Len(t, outArr, len(expected))
		for i, v := range expected {
			assert.Equal(t, v, outArr[i])
		}
	}
}

func TestContainsInStringList(t *testing.T) {
	list := []string{
		"aaa",
		"bbb",
		"ccc",
		"ddd",
		"eee",
	}
	assert.False(t, ContainsInStringList(list, ""))
	for _, v := range list {
		assert.True(t, ContainsInStringList(list, v))
		assert.False(t, ContainsInStringList(list, strings.ToUpper(v)))
	}
}

func TestRemoveStringFromStringList(t *testing.T) {
	list := []string{
		"aaa",
		"bbb",
		"ccc",
		"ddd",
		"eee",
		"ccc",
	}
	expected := []string{
		"aaa",
		"bbb",
		"ddd",
		"eee",
	}
	cnt := RemoveStringFromStringList("ccc", &list)
	assert.Equal(t, 2, cnt)
	assert.Equal(t, expected, list)
}

func TestRemoveDuplicatesFromStringList(t *testing.T) {
	list := []string{
		"aaa",
		"bbb",
		"ccc",
		"ddd",
		"eee",
		"ccc",
		"aaa",
		"eee",
	}
	expected := []string{
		"aaa",
		"bbb",
		"ccc",
		"ddd",
		"eee",
	}
	out := RemoveDuplicatesFromStringList(list)
	assert.ElementsMatch(t, expected, out)
}

func TestEqualStringListIgnoreOrder(t *testing.T) {
	src := []string{
		"aaa",
		"bbb",
		"ccc",
		"ddd",
		"eee",
	}
	match := []string{
		"bbb",
		"aaa",
		"eee",
		"ddd",
		"ccc",
	}
	unmatch := []string{
		"bbb",
		"aaa",
		"ddd",
		"ccc",
		"kkk",
	}
	assert.False(t, EqualStringListIgnoreOrder(src, unmatch))
	assert.True(t, EqualStringListIgnoreOrder(src, match))

}

func TestFindDuplicateItemsBetweenStringLists(t *testing.T) {
	sampleStringList1 := []string{
		"AAA",
		"BBB",
		"CCC",
		"DDD",
		"EEE",
	}
	sampleStringList2 := []string{
		"CCC",
		"DDD",
		"EEE",
		"FFF",
		"GGG",
		"HHH",
		"III",
	}
	sampleStringList3 := []string{
		"AAA",
		"BBB",
		"FFF",
		"GGG",
		"HHH",
		"III",
	}

	type args struct {
		list1 []string
		list2 []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"sampleStringList1, sampleStringList2",
			args{sampleStringList1, sampleStringList2},
			[]string{
				"CCC",
				"DDD",
				"EEE",
			},
		},
		{
			"sampleStringList2, sampleStringList3",
			args{sampleStringList2, sampleStringList3},
			[]string{
				"FFF",
				"GGG",
				"HHH",
				"III",
			},
		},
		{
			"empty, sampleStringList3",
			args{[]string{}, sampleStringList3},
			[]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindDuplicateItemsBetweenStringLists(tt.args.list1, tt.args.list2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindDuplicateItemsBetweenStringLists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindNotExistItemsInSrcStringList(t *testing.T) {
	sampleStringList1 := []string{
		"AAA",
		"BBB",
		"CCC",
		"DDD",
		"EEE",
	}
	sampleStringList2 := []string{
		"CCC",
		"DDD",
		"EEE",
		"FFF",
		"GGG",
		"HHH",
		"III",
	}
	sampleStringList3 := []string{
		"AAA",
		"BBB",
		"FFF",
		"GGG",
		"HHH",
		"III",
	}
	type args struct {
		src []string
		dst []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"sampleStringList1, sampleStringList2",
			args{sampleStringList1, sampleStringList2},
			[]string{
				"FFF",
				"GGG",
				"HHH",
				"III",
			},
		},
		{
			"sampleStringList2, sampleStringList3",
			args{sampleStringList2, sampleStringList3},
			[]string{
				"AAA",
				"BBB",
			},
		},
		{
			"empty, sampleStringList3",
			args{[]string{}, sampleStringList3},
			sampleStringList3,
		},
		{
			"sampleStringList1, empty",
			args{sampleStringList1, []string{}},
			[]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindNotExistItemsInSrcStringList(tt.args.src, tt.args.dst); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindNotExistItemsInSrcStringList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindDifferenceItemsBetweenStringLists(t *testing.T) {
	sampleStringList1 := []string{
		"AAA",
		"BBB",
		"CCC",
		"DDD",
		"EEE",
	}
	sampleStringList2 := []string{
		"CCC",
		"DDD",
		"EEE",
		"FFF",
		"GGG",
		"HHH",
		"III",
	}
	sampleStringList3 := []string{
		"AAA",
		"BBB",
		"FFF",
		"GGG",
		"HHH",
		"III",
	}
	type args struct {
		list1 []string
		list2 []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"sampleStringList1, sampleStringList2",
			args{sampleStringList1, sampleStringList2},
			[]string{
				"AAA",
				"BBB",
				"FFF",
				"GGG",
				"HHH",
				"III",
			},
		},
		{
			"sampleStringList2, sampleStringList3",
			args{sampleStringList2, sampleStringList3},
			[]string{
				"AAA",
				"BBB",
				"CCC",
				"DDD",
				"EEE",
			},
		},
		{
			"empty, sampleStringList3",
			args{[]string{}, sampleStringList3},
			sampleStringList3,
		},
		{
			"sampleStringList1, empty",
			args{sampleStringList1, []string{}},
			sampleStringList1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindDifferenceItemsBetweenStringLists(tt.args.list1, tt.args.list2); !assert.ElementsMatch(t, got, tt.want) {
				t.Errorf("FindDifferenceItemsBetweenStringLists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ObjectToMapInterface(t *testing.T) {
	type TestObj struct {
		CreatedAt time.Time
		Name      string
	}
	obj := TestObj{
		Name:      "333",
		CreatedAt: time.Now(),
	}
	mapIf, err := ObjectToMapInterface(&obj)
	assert.NoError(t, err)
	objBytes, err := json.Marshal(&obj)
	assert.NoError(t, err)
	mapIfBytes, err := json.Marshal(mapIf)
	assert.NoError(t, err)
	assert.Equal(t, string(objBytes), string(mapIfBytes))

}

func TestTimeToISOString(t *testing.T) {
	now := time.Now()
	str := TimeToISOString(now)

	fmt.Println(str)
	assert.Regexp(t, `\d{4}(.\d{2}){2}(\s|T)(\d{2}.){2}\d{2}`, str)
}

func TestTimeFromIsoString(t *testing.T) {
	t.Run("정상", func(t *testing.T) {
		str := "2023-01-27T19:16:07+09:00"
		ti, _ := TimeFromIsoString(str)

		assert.IsType(t, &time.Time{}, ti)
	})

	t.Run("비정상", func(t *testing.T) {
		str := "2023-01-27"
		ti, err := TimeFromIsoString(str)

		assert.Nil(t, ti)
		assert.Error(t, err)
	})
}
