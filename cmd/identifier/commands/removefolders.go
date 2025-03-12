package commands

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/identifier/identifier"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"regexp"
)

var removeFoldersRemoveFlag bool
var removeFoldersRegexpFlag string

var removeFoldersCmd = &cobra.Command{
	Use:     "removefolders [path to data]",
	Aliases: []string{},
	Short:   "removes folders including files and subfolders based on go regular expression",
	Long: `removes folders including files and subfolders based on go regular expression (https://pkg.go.dev/regexp/syntax)
If there are multiple folders in one hierarchy matching the regular expression, only the first one with lowest depth will be listed 
or removed, which inherently includes the rest.

Caveat: dry-run (no --remove flag) is always recommended before removing files from filesystem.
`,
	Example: `Find all folders with more or equal 8 characters in name

` + appname + ` removefolders C:/daten/aiptest --regexp ".{8,}"
2025-03-12T12:18:02+01:00 INF dry-run: no files will be removed timestamp="2025-03-12 12:18:02.2662753 +0100 CET m=+0.164232301"
2025-03-12T12:18:02+01:00 INF working on folder 'C:/daten/aiptest0' timestamp="2025-03-12 12:18:02.2662753 +0100 CET m=+0.164232301"
2025-03-12T12:18:02+01:00 INF using regexp ".{8,}" timestamp="2025-03-12 12:18:02.2662753 +0100 CET m=+0.164232301"
payload/#1    audio
payload/empty folder

list all folders starting with 'empty' and remove them

` + appname + ` removefolders C:/daten/aiptest --regexp "^empty" --remove
2025-03-12T12:22:07+01:00 INF working on folder 'C:/daten/aiptest0' timestamp="2025-03-12 12:22:07.4433671 +0100 CET m=+0.113916501"
2025-03-12T12:22:07+01:00 INF using regexp "^empty" timestamp="2025-03-12 12:22:07.4433671 +0100 CET m=+0.113916501"
payload/empty folder
2025-03-12T12:22:07+01:00 INF removing 'C:\daten\aiptest0\payload\empty folder' timestamp="2025-03-12 12:22:07.4433671 +0100 CET m=+0.113916501"`,
	Args: cobra.ExactArgs(1),
	Run:  doRemoveFolders,
}

func removeFoldersInit() {
	removeFoldersCmd.Flags().StringVar(&removeFoldersRegexpFlag, "regexp", "", "[required] regular expression to match files")
	removeFoldersCmd.MarkFlagRequired("regexp")
	removeFoldersCmd.Flags().BoolVar(&removeFoldersRemoveFlag, "remove", false, "removes (deletes) the folders including files and subfolders from filesystem (if not set it's just a dry run)")
}

func doRemoveFolders(cmd *cobra.Command, args []string) {
	dataPath, err := identifier.Fullpath(args[0])
	cobra.CheckErr(err)
	if fi, err := os.Stat(dataPath); err != nil || !fi.IsDir() {
		cobra.CheckErr(errors.Errorf("'%s' is not a directory", dataPath))
	}

	folderRegexp, err := regexp.Compile(removeFoldersRegexpFlag)
	cobra.CheckErr(errors.Wrapf(err, "cannot compile '%s'", removeFoldersRegexpFlag))

	if !removeFoldersRemoveFlag {
		logger.Info().Msg("dry-run: no files will be removed")
	}
	logger.Info().Msgf("working on folder '%s'", dataPath)
	logger.Info().Msgf("using regexp \"%s\"", removeFoldersRegexpFlag)
	dirFS := os.DirFS(dataPath)
	pathElements, err := identifier.BuildPath(dirFS)
	cobra.CheckErr(errors.Wrapf(err, "cannot build paths from '%s'", dataPath))

	for name := range pathElements.FindDirname(folderRegexp) {
		fmt.Printf("%s\n", name)
		if removeFoldersRemoveFlag {
			fullpath := filepath.Join(dataPath, name)
			logger.Info().Msgf("removing '%s'", fullpath)
			if err := os.RemoveAll(fullpath); err != nil {
				logger.Fatal().Err(err).Msgf("cannot remove '%s'", fullpath)
			}
		}
	}
	return
}
