package main

import (
	"bufio"
	"fmt"
	"github.com/fly-studio/envsubst/parse"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

type envOptions struct {
	shellFormat      string
	shellFormatIsSet bool
	files            []string
	showVariables    bool
	keepUnset        bool
	unsetFatal       bool
	emptyFatal       bool
}

func main() {
	var options envOptions
	rootCmd := &cobra.Command{
		Use:   "envsubst [SHELL-FORMAT] [-f --file IN:OUT] [--keep-unset] [--unset-fatal] [--empty-fatal]",
		Short: "Substitutes the values of environment variables.",
		Long: `In normal operation mode, standard input is copied to standard output,
with references to environment variables of the form $VARIABLE or ${VARIABLE}
being replaced with the corresponding values.  If a SHELL-FORMAT is given,
only those environment variables that are referenced in SHELL-FORMAT are
substituted; otherwise all environment variables references occurring in
standard input are substituted.

When --variables is used, standard input is ignored, and the output consists
of the environment variables that are referenced in SHELL-FORMAT, one per line.`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.PersistentFlags().Changed("version") {
				_ = cmd.Help()
				return
			}
			if len(args) > 0 {
				options.shellFormat = args[0]
				options.shellFormatIsSet = true
			}
			if err := envsubst(options); err != nil {
				panic(err)
			}
		},
	}
	rootCmd.PersistentFlags().StringArrayVarP(&options.files, "file", "f", []string{}, "the files of \"IN:OUT\", or in-place via \"IN\"")
	rootCmd.PersistentFlags().BoolVarP(&options.showVariables, "variables", "v", false, "output the variables occurring in SHELL-FORMAT")
	rootCmd.PersistentFlags().BoolVar(&options.keepUnset, "keep-unset", false, "keep raw \"$KEY\", if key of env is not set. always be true if SHELL-FORMAT is given")
	rootCmd.PersistentFlags().BoolVar(&options.keepUnset, "unset-fatal", false, "fatal if the key of env is not set")
	rootCmd.PersistentFlags().BoolVar(&options.keepUnset, "empty-fatal", false, "fatal if the value of env is empty")
	rootCmd.PersistentFlags().BoolP("version", "V", false, "output version information and exit")

	err := rootCmd.Execute()
	if err != nil {
		panic(err.Error())
	}
}

func envsubst(options envOptions) error {

	var envKeys []string
	envKeys = EnvKeys(options.shellFormat)

	// 只显示变量名
	if options.showVariables {
		for _, name := range envKeys {
			fmt.Println(name)
		}
		return nil
	}

	if options.shellFormatIsSet {
		// 按照envsubst的做法，只要Shell-Format有传递，哪怕为空，都会按照Shell-Format的格式来替换环境变量。
		// 此处特意添加一个不存在的KEY，是为了避免EnvSubstitute使用os.Environ()
		envKeys = append(envKeys, "EMPTY_STRING")
	}

	customEnv := GetEnvMap(envKeys)
	restrictions := &parse.Restrictions{
		ErrorOnUnset: options.unsetFatal,
		ErrorOnEmpty: options.emptyFatal,
		KeepUnset:    options.keepUnset || options.shellFormatIsSet,
	}

	if len(options.files) > 0 {
		fmt.Printf("Environment variables substitute:\n")
		for _, file := range options.files {
			segments := strings.SplitN(file, ":", 2)
			inFile := segments[0]
			outFile := inFile
			if len(segments) > 1 {
				outFile = segments[1]
			}

			if err := EnvSubstituteFile(inFile, outFile, customEnv, restrictions); err != nil {
				return err
			}
			fmt.Printf(" - \"%s\" to \"%s\"\n", inFile, outFile)
		}
		return nil
	} else {
		stat, err := os.Stdin.Stat()
		if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
			return fmt.Errorf("must input a valid file or content, \"envsubst < 1.txt\"")
		}
		content, err := io.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			return err
		}

		out, err := EnvSubstitute(string(content), customEnv, restrictions)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	}
}
