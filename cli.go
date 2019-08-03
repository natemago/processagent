package processagent

import "flag"

// Config holds the program arguments values as configuration.
type Config struct {
	Port       *int
	Command    *string
	MaxWorkers *int
}

// RunCommand runs a CLI command with the given Config.
type RunCommand func(*Config) error

func configureFlags() *Config {
	cfg := Config{}

	cfg.Port = flag.Int("p", 8080, "Expose on port. Default 8080.")
	cfg.MaxWorkers = flag.Int("max-workers", 0, "Maximal number of parallel workers. Set 0 for unlimited.")
	cfg.Command = flag.String("c", "", "Command to execute.")

	return &cfg
}

// RunCLI configures the flags, parses the program arguments then runs the given
// command with the Config extracted from those arguments.
func RunCLI(command RunCommand) error {
	config := configureFlags()
	flag.Parse()
	return command(config)
}
