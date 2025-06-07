package cmd

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    cfgFile   string
    namespace string
    output    string
    verbose   bool
    quiet     bool
    dryRun    bool
    context   string
)

var rootCmd = &cobra.Command{
    Use:   "lhcli",
    Short: "Longhorn CLI - A command-line interface for Longhorn",
    Long: `lhcli is a comprehensive CLI tool for managing Longhorn storage system.
It provides functionality equivalent to the Longhorn WebUI and more.`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    cobra.OnInitialize(initConfig)
    
    // Global flags
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.lhcli/config.yaml)")
    rootCmd.PersistentFlags().StringVar(&context, "context", "", "context to use from the config file")
    rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "longhorn-system", "Longhorn namespace")
    rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format (table|json|yaml|wide)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
    rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Minimal output")
    rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview actions without executing")
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        
        viper.AddConfigPath(home + "/.lhcli")
        viper.SetConfigType("yaml")
        viper.SetConfigName("config")
    }
    
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err == nil {
        if verbose {
            fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
        }
    }
}
