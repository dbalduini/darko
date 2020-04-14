package dotenv

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Matches a string surrounded by single or double quotes
	quoteSurroundedRegexp = regexp.MustCompile(`^['"]|['"]$`)
)

type Env string

func (e Env) Int() int {
	i, err := strconv.Atoi(string(e))
	if err != nil {
		panic(err)
	}
	return i
}

func (e Env) String() string {
	return string(e)
}

// LoadEnvFromConfigFile loads dotenv from "CONFIG_FILE" environment variable if it is set.
func LoadEnvFromConfigFile() error {
	if f, ok := os.LookupEnv("CONFIG_FILE"); ok {
		return LoadFile(f)
	}
	return nil
}

// LoadFile loads the dotenv file.
func LoadFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return Load(file)
}

// Load loads the reader into os environment variables.
func Load(r io.Reader) error {
	// Read lines
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		k, v := getKeyValue(line)
		setEnvVar(k, v)
	}
	return scanner.Err()
}

// SetEnvVar dont replace env vars already set on the OS
func setEnvVar(k, v string) {
	if _, ok := os.LookupEnv(k); !ok {
		os.Setenv(k, unquoteAndTrim(v))
	}
}

// GetKeyValue returns the key, value splitted from the given line.
func getKeyValue(line string) (k string, v string) {
	// skip commetary line
	if line == "" || strings.HasPrefix(line, "#") {
		return k, v
	}
	i := strings.Index(line, "=")
	if i > -1 {
		k = line[:i]
		v = line[i+1:]
	}
	return k, v
}

// UnquoteAndTrim remove single and double quotes and trim the string.
func unquoteAndTrim(s string) (v string) {
	v = quoteSurroundedRegexp.ReplaceAllLiteralString(s, "")
	return strings.Trim(v, " ")
}

// Validate checks if vars are in the os env.
func Validate(vars []string) (string, bool) {
	for _, k := range vars {
		if _, ok := os.LookupEnv(k); !ok {
			return k, false
		}
	}
	return "", true
}

// GetOrElse returns the value for the environment value or the default value.
func GetOrElse(key, defaultValue string) Env {
	if v, ok := os.LookupEnv(key); ok {
		return Env(v)
	}
	return Env(defaultValue)
}

// MustGet returns the value for the key if present or panic otherwise
func MustGet(key string) Env {
	if v, ok := os.LookupEnv(key); ok {
		return Env(v)
	}
	panic(1)
}

// GetBool returns the value for the key as bool if present or false otherwise
func GetBool(key string) bool {
	return GetOrElse(key, "false").String() == "true"
}
