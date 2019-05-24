package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func readKVStrings(files []string, override []string) ([]string, error) {
	envVariables :=[]string{}
	for _, ef:=range files{
		parsedVars, err := ParseEnvFile(ef)
		if err != nil{
			return nil, err
		}
		envVariables = append(envVariables, parsedVars...)
	}
	//parse the '-e' and '--env' after, to allow override
	envVariables = append(envVariables, override...)

	return envVariables, nil
}

func ParseEnvFile(filename string) ([]string, error){
	fh, err := os.Open(filename)
	if err!=nil{
		return []string{}, err
	}
	defer fh.Close()

	lines := []string{}
	scanner := bufio.NewScanner(fh)
	for scanner.Scan(){
		line:=strings.TrimLeft(scanner.Text(), whiteSpaces)
		if len(line)>0&&!strings.HasPrefix(line, "#"){
			data := strings.SplitN(line, "=", 2)

			variable:=strings.TrimLeft(data[0], whiteSpaces)
			if strings.ContainsAny(variable, whiteSpaces){
				return []string{}, ErrBadEnvVariable{fmt.Sprintf("variable '%s' has white spaces", variable)}
			}

			if len(data)>1{
				lines=append(lines, fmt.Sprintf("%s=%s",variable, data[1]))
			}else{
				lines=append(lines, fmt.Sprintf("%s=%s",strings.TrimSpace(line),os.Getenv(line)))
			}
		}
	}
	return lines, scanner.Err()
}

var whiteSpaces=" \t"

type ErrBadEnvVariable struct{
	msg string
}

func (e ErrBadEnvVariable) Error() string {
	return fmt.Sprintf("poorly formatted environment: %s", e.msg)
}
