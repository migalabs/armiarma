package utils

import (
	"strings"
	"time"

	ma "github.com/multiformats/go-multiaddr"
)

// ReturnGreatestTime
// * This method return the latest time inside the given array
// @param input_array: the array of times to compare
// @return the latest time inside the array
func ReturnGreatestTime(input_array []time.Time) time.Time {
	latestTime := input_array[0]

	for _, time_tmp := range input_array {
		if time_tmp.After(latestTime) {
			latestTime = time_tmp
		}
	}
	return latestTime
}

func ReturnMaxInt(input_array []int) int {
	max_int := input_array[0]

	for _, int_tmp := range input_array {
		if int_tmp > max_int {
			max_int = int_tmp
		}
	}
	return max_int
}

func ParseInterfaceStringArray(input_arr []interface{}) []string {
	result := make([]string, 0)
	if input_arr != nil {
		// we will range over a slice of interfaces
		for _, v := range input_arr {
			result = append(result, v.(string))
		}
	}

	return result
}

func ParseInterfaceTimeArray(input_arr []interface{}) ([]time.Time, error) {
	result := make([]time.Time, 0)
	if input_arr != nil {
		// we will range over a slice of interfaces
		for _, v := range input_arr {
			new_time, err := time.Parse(time.RFC3339, v.(string))
			if err != nil {
				return result, err
			}
			result = append(result, new_time)
		}

	}
	return result, nil

}

func ParseInterfaceAddrArray(input_arr []interface{}) ([]ma.Multiaddr, error) {
	result := make([]ma.Multiaddr, 0)

	if input_arr != nil {
		// we will range over a slice of interfaces
		for _, v := range input_arr {
			new_addr, err := UnmarshalMaddr(v.(string))
			if err != nil {
				return result, err
			}
			result = append(result, new_addr)
		}
	}
	return result, nil

}

func ExistsInArray(inputList []string, inputValue string) bool {
	for _, value := range inputList {
		if strings.ToLower(inputValue) == strings.ToLower(value) {
			return true
		}
	}
	return false
}

func ExistsInMapValue(inputMap map[string]string, inputValue string) bool {
	for _, value := range inputMap {

		if strings.ToLower(inputValue) == strings.ToLower(value) {
			return true
		}
	}
	return false
}
