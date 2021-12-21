package utils

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_ReturnGreatestTime(t *testing.T) {

	tNow := time.Now()
	timeArray := make([]time.Time, 0)

	greatestTime := ReturnGreatestTime(timeArray)
	require.Equal(t, time.Time{}, greatestTime)

	timeArray = append(timeArray, tNow.Add(60*time.Second))
	timeArray = append(timeArray, tNow.Add(30*time.Second))
	timeArray = append(timeArray, tNow.Add(45*time.Second))

	greatestTime = ReturnGreatestTime(timeArray)

	require.Equal(t, tNow.Add(60*time.Second), greatestTime)
}

func Test_ReturnMaxInt(t *testing.T) {

	intArray := make([]int, 0)

	greatestInt := ReturnMaxInt(intArray)
	require.Equal(t, math.MinInt, greatestInt)

	intArray = append(intArray, 40)
	intArray = append(intArray, 90)
	intArray = append(intArray, 50)

	greatestInt = ReturnMaxInt(intArray)

	require.Equal(t, 90, greatestInt)

	intArray = append(intArray, -50)
	greatestInt = ReturnMaxInt(intArray)
	require.Equal(t, 90, greatestInt)

	intArray = append(intArray, 91)
	greatestInt = ReturnMaxInt(intArray)
	require.Equal(t, 91, greatestInt)

	intArray = append(intArray, 91)
	greatestInt = ReturnMaxInt(intArray)
	require.Equal(t, 91, greatestInt)
}

func Test_ExistsinArray(t *testing.T) {

	stringArray := make([]string, 0)
	searchValue := "hello"
	require.Equal(t, false, ExistsInArray(stringArray, searchValue))

	stringArray = append(stringArray, "hello")
	require.Equal(t, true, ExistsInArray(stringArray, searchValue))

}

func Test_ExistsinMapValue(t *testing.T) {

	stringMap := make(map[string]string)
	searchValue := "hello"
	require.Equal(t, false, ExistsInMapValue(stringMap, searchValue))

	stringMap["first"] = "hello"
	require.Equal(t, true, ExistsInMapValue(stringMap, searchValue))

	searchValue = "he"
	require.Equal(t, false, ExistsInMapValue(stringMap, searchValue))

	searchValue = "helloo"
	require.Equal(t, false, ExistsInMapValue(stringMap, searchValue))

}
