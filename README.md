# cfg

A convenient, type-safe configuration library for Go that populates your structs from multiple sources including environment variables, command-line flags, JSON files, and more.

## Features

- **Multiple Providers**: Load configuration from environment variables, flags, JSON, or static values
- **Type Safety**: Automatic type conversion for primitive types, slices, maps, and custom types
- **Struct Tags**: Simple tag-based configuration using `env`, `flag`, `json` tags
- **Composable**: Chain multiple providers with fallback behavior
- **Extensible**: Create custom providers for any configuration source
- **CLI Support**: Built-in CLI framework with command hierarchies and help generation

## Quick Start

```go
package main

import (
  "context"
  "fmt"
  "os"

  "github.com/upfluence/cfg"
)

type Config struct {
  Host     string `env:"HOST" flag:"host"`
  Port     int    `env:"PORT" flag:"port"`
  Debug    bool   `env:"DEBUG" flag:"debug"`
  Timeout  time.Duration `env:"TIMEOUT" flag:"timeout"`
}

func main() {
  var c Config
  ctx := context.Background()

  if err := cfg.NewDefaultConfigurator().Populate(ctx, &c); err != nil {
    fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
    os.Exit(1)
  }

  fmt.Printf("Server: %s:%d (debug=%v, timeout=%s)\n",
    c.Host, c.Port, c.Debug, c.Timeout)
}
```

The default configurator uses both environment variables and command-line flags:

```bash
# Using environment variables
$ HOST=localhost PORT=8080 DEBUG=true TIMEOUT=30s ./app
Server: localhost:8080 (debug=true, timeout=30s)

# Using flags
$ ./app --host localhost --port 8080 --debug --timeout 30s
Server: localhost:8080 (debug=true, timeout=30s)

# Boolean flags support --no- prefix
$ ./app --host localhost --no-debug
Server: localhost: (debug=false, timeout=0s)
```

## Providers

### Environment Variables

The `env` provider reads from environment variables with automatic uppercasing and underscore conversion.

```go
type Config struct {
  APIKey    string `env:"API_KEY"`
  BaseURL   string `env:"BASE_URL"`
}
```

You can also use prefixed environment variables:

```go
provider := env.NewProvider("MYAPP") // Looks for MYAPP_* variables
configurator := cfg.NewConfigurator(provider)
```

### Command-Line Flags

The `flags` provider parses command-line arguments with support for short flags, equals syntax, and boolean negation.

```go
type Config struct {
  Verbose  bool   `flag:"verbose,v"`  // --verbose or -v
  Output   string `flag:"output,o"`   // --output file or -o file
  Force    bool   `flag:"force"`      // --force or --no-force
}
```

Supported formats:
- `--flag value`
- `--flag=value`
- `-f value`
- `--flag` (boolean true)
- `--no-flag` (boolean false)

### JSON Files

Load configuration from JSON files:

```go
file, _ := os.Open("config.json")
defer file.Close()

jsonProvider := json.NewProviderFromReader(file)
configurator := cfg.NewConfigurator(jsonProvider)
```

The JSON provider supports nested structures using dot notation:

```json
{
  "database": {
    "host": "localhost",
    "port": 5432
  }
}
```

```go
type Config struct {
  DBHost string `json:"database.host"`
  DBPort int    `json:"database.port"`
}
```

### Static Provider

Provide configuration from Go values directly:

```go
staticProvider := static.NewProvider(map[string]interface{}{
  "host": "localhost",
  "port": 8080,
})
configurator := cfg.NewConfigurator(staticProvider)
```

### Custom Providers

Create your own provider by implementing the `Provider` interface:

```go
type Provider interface {
  StructTag() string
  Provide(context.Context, string) (string, bool, error)
}
```

Example custom provider:

```go
type ConsulProvider struct{}

func (p *ConsulProvider) StructTag() string { return "consul" }

func (p *ConsulProvider) Provide(ctx context.Context, key string) (string, bool, error) {
  // Fetch from Consul
  value, err := consulClient.Get(key)
  if err == ErrNotFound {
    return "", false, nil
  }
  return value, true, err
}
```

## Provider Chaining

Providers are evaluated in order. The first provider that returns a value wins:

```go
configurator := cfg.NewConfigurator(
  flags.NewDefaultProvider(),    // Check flags first
  env.NewDefaultProvider(),       // Then environment variables
  jsonProvider,                   // Then JSON file
  staticProvider,                 // Finally defaults
)
```

## Supported Types

The library automatically handles type conversion for:

- **Primitives**: `string`, `bool`, all `int` and `float` types
- **Time**: `time.Duration` (via `time.ParseDuration`), `time.Time` (RFC3339 format)
- **Slices**: Comma-separated values (`"a,b,c"` → `[]string{"a", "b", "c"}`)
- **Maps**: Key-value pairs (`"k1=v1,k2=v2"` → `map[string]string{"k1": "v1", "k2": "v2"}`)
- **Nested Structs**: Dot notation for nested fields
- **Custom Types**: Any type implementing:
  - `json.Unmarshaler`
  - `encoding.TextUnmarshaler`
  - `interface { Parse(string) error }`

## Advanced Usage

### Configuration Options

```go
configurator := cfg.NewConfiguratorWithOptions(
  cfg.WithProviders(myProvider),
  cfg.IgnoreMissingTag,  // Don't error on fields without tags
)
```

### Default Values

Use the `default` struct tag to provide fallback values when no other provider
supplies one:

```go
type Config struct {
  Host    string `env:"HOST" flag:"host" default:"localhost"`
  Port    int    `env:"PORT" flag:"port" default:"8080"`
  Verbose bool   `env:"VERBOSE"          default:"false"`
}
```

The default provider has the lowest priority — any value from environment
variables, flags, or other providers takes precedence. It is included
automatically in `NewDefaultConfigurator`. Nested structs work as expected:
only fields with an explicit `default` tag receive a value.

### Nested Structs

```go
type Database struct {
  Host string `env:"DB_HOST" flag:"db-host"`
  Port int    `env:"DB_PORT" flag:"db-port"`
}

type Config struct {
  Database Database  // Automatically traversed
  AppName  string `env:"APP_NAME" flag:"app-name"`
}
```

### Custom Configurator

For more control, create a configurator without the default providers:

```go
configurator := cfg.NewConfigurator(
  myProvider1,
  myProvider2,
)
```

## CLI Framework

The `x/cli` package provides a framework for building CLI applications with automatic help generation:

```go
import "github.com/upfluence/cfg/x/cli"

type RunConfig struct {
  Host string `flag:"host,h"`
  Port int    `flag:"port,p"`
}

cmd := cli.StaticCommand{
  Help:     cli.HelpWriter(&RunConfig{}),
  Synopsis: cli.SynopsisWriter(&RunConfig{}),
  Execute: func(ctx context.Context, cctx cli.CommandContext) error {
    var cfg RunConfig
    // Configuration is automatically populated
    fmt.Printf("Running on %s:%d\n", cfg.Host, cfg.Port)
    return nil
  },
}

app := cli.NewApp(
  cli.WithName("myapp"),
  cli.WithCommand(cmd),
)
app.Run(context.Background())
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.
