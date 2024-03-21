package converts

import "strconv"

func ToInt(numberString string) int {
    num, err := strconv.Atoi(numberString)

    if err != nil {
        return 0 
    }

    return num
}

func ToFloat64(numberString string) float64 {
    num, err := strconv.ParseFloat(numberString,2)

    if err != nil {
        return 0 
    }

    return num
}


