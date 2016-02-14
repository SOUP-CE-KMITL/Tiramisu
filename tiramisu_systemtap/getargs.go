package main

import (
    "fmt"
    "strconv"
    "strings"

    "gopkg.in/pipe.v2"
)

func GetArguments(pid int) []string {
    filename := "/proc/" + strconv.Itoa(pid) + "/cmdline"
    p := pipe.Line(
        pipe.ReadFile(filename),
        pipe.Exec("strings", "-1"),
    )
    output, err := pipe.CombinedOutput(p)
    if err != nil {
        fmt.Printf("error:[%v]\n", err)
    }
    return strings.Fields(string(output))
}

func main() {
    wordlist := GetArguments(19351)
    for index, element := range wordlist {
        fmt.Printf("%v: %v\n", index, element)
    }
}
