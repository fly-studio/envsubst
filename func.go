package main

import (
	"github.com/fly-studio/envsubst/parse"
	"os"
	"regexp"
	"strings"
)

var envRegexp = regexp.MustCompile(`\$\{?([\w\d_]*)\}?`)

func EnvKeys(shellFormat string) []string {
	if !strings.Contains(shellFormat, "$") {
		return nil
	}
	_envList := envRegexp.FindAllStringSubmatch(shellFormat, -1)

	var env []string
	for _, line := range _envList {
		env = append(env, line[1])
	}

	return env
}

func GetEnvMap(envKeys []string) map[string]string {
	if len(envKeys) <= 0 {
		return nil
	}

	env := map[string]string{}
	for _, name := range envKeys {
		env[name], _ = os.LookupEnv(name)
	}
	return env
}

func EnvSubstitute(content string, customEnv map[string]string, restrictions *parse.Restrictions) (string, error) {
	envLine := os.Environ()
	if len(customEnv) > 0 {
		restrictions.KeepUnset = true
		envLine = EnvToStrings(customEnv)
	}
	return parse.New("string", envLine, restrictions).Parse(content)
}

func EnvToStrings(env map[string]string) []string {
	var lines []string
	for k, v := range env {
		lines = append(lines, k+"="+v)
	}

	return lines
}

func EnvSubstituteFile(fromFile, toFile string, env map[string]string, restrictions *parse.Restrictions) error {
	stat, err := os.Stat(fromFile)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(fromFile)
	if err != nil {
		return err
	}

	out, err := EnvSubstitute(string(content), env, restrictions)
	if err != nil {
		return err
	}

	return os.WriteFile(toFile, []byte(out), stat.Mode())
}
