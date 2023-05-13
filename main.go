package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/abdfnx/gosh"
	"github.com/nxadm/tail"
)

// regex pattern
type Pattern struct {
	Name    string
	Regex   string // actual regex
	Command string
	Alert   string
}

type Config struct {
	Logging  string
	Log_file string
	Files    []string
	Regex    []Pattern
}

func canAccessFile(filepath string) error {
	_, err := os.Stat(filepath)

	if err == nil {
		return nil
	} else if errors.Is(err, os.ErrNotExist) {
		return err
	}

	// other error (e.g. insufficient permissions)
	return err
}

func main() {
	// Help screen
	if len(os.Args) != 2 {
		fmt.Printf("\nRun commands based on RegEx matches in continuously read files\n\nUsage: logspark [configuration file]\n\n")
		return
	}

	if err := canAccessFile(os.Args[1]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var config Config

	if _, err := toml.DecodeFile(os.Args[1], &config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// check for empty logging configuration
	if config.Logging == "" {
		fmt.Printf("Warning! No logging method specified, defaulting to verbose.\n\n")
	} else if config.Logging != "verbose" && config.Logging != "minimal" && config.Logging != "none" {
		fmt.Printf("Warning! Invalid logging method, defaulting to verbose.\n\n")
	}

	// check for duplicate values
	var duplicates []string
	for i := 0; i+1 < len(config.Regex); i++ {
		// name
		if config.Regex[i].Name == config.Regex[i+1].Name {
			duplicates = append([]string(duplicates), config.Regex[i].Name)
		}

		// regex
		if config.Regex[i].Regex == config.Regex[i+1].Regex {
			duplicates = append([]string(duplicates), config.Regex[i].Regex)
		}

		// command
		if config.Regex[i].Command == config.Regex[i+1].Command {
			duplicates = append([]string(duplicates), config.Regex[i].Command)
		}

		// alert
		if config.Regex[i].Alert == config.Regex[i+1].Alert {
			duplicates = append([]string(duplicates), config.Regex[i].Alert)
		}
	}

	for _, element := range duplicates {
		fmt.Printf("Error! Duplicate value: %s\n", element)
		os.Exit(1)
	}

	// check for empty regex
	for i := 0; i < len(config.Regex); i++ {
		if len(config.Regex[i].Regex) == 0 {
			var input string
			for input == "" || input != "y" && input != "n" {
				fmt.Printf("Regex patterns empty, are you sure you want to continue? [y/n] ")

				fmt.Scanf("%s", &input)
			}

			if input == "n" {
				os.Exit(0)
			}
		}
	}

	// tails array (tail for each file)
	tails := make([]*tail.Tail, len(config.Files))

	// for each file
	for i := 0; i < len(config.Files); i++ {
		// check if the file exists
		if err := canAccessFile(config.Files[i]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			// if it does, tail the file
			t, err := tail.TailFile(config.Files[i], tail.Config{
				Follow: true,
				ReOpen: true,
				Poll:   true,
				Logger: tail.DiscardingLogger,
			})

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			tails[i] = t // add the tail to the array
		}

		if config.Log_file == config.Files[i] {
			fmt.Println("Error: cannot tail log file (possible loop)")
			os.Exit(1)
		}
	}

	for i := 0; i < len(tails); i++ {
		// tail each file concurrently
		go func(i int) {
			for line := range tails[i].Lines {
				// ignore empty
				if line.Text == "\n" {
					continue
				}
				// for each [[regex]] in the TOML file
				for j := 0; j < len(config.Regex); j++ {
					regex, err := regexp.Compile(config.Regex[j].Regex)

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					// check for the regex
					if regex.MatchString(line.Text) {
						var output string
						// log verbosity
						switch config.Logging {
						default:
							fallthrough
						case "verbose":
							output = fmt.Sprintf("%s | %s | %s | %s | %s | %s | %s", time.Now().Format(time.RFC1123), tails[i].Filename, config.Regex[j].Name, config.Regex[j].Regex, config.Regex[j].Command, config.Regex[j].Alert, line.Text)

						case "minimal":
							output = fmt.Sprintf("%s | %s | %s | %s", time.Now().Format(time.DateTime), tails[i].Filename, config.Regex[j].Name, line.Text)
						}

						// log accordingly
						switch config.Log_file {
						default:
							file, err := os.OpenFile(config.Log_file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

							if err != nil {
								fmt.Println(err)
								os.Exit(1)
							}

							fmt.Fprintf(file, "%s\n", output)
						case "stdout":
							fmt.Println(output)
						case "":
						case "none":
						}

						// run the command
						gosh.Run(config.Regex[j].Command)
						// run the "alert" command
						gosh.Run(config.Regex[j].Alert)
					}
				}
			}

		}(i)
	}

	select {}
}
