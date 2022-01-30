package stew

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/marwanhawari/stew/constants"
	"github.com/mholt/archiver"
	progressbar "github.com/schollz/progressbar/v3"
)

func IsArchiveFile(filePath string) bool {
	fileExtension := filepath.Ext(filePath)
	if fileExtension == ".br" || fileExtension == ".bz2" || fileExtension == ".zip" || fileExtension == ".gz" || fileExtension == ".lz4" || fileExtension == ".sz" || fileExtension == ".xz" || fileExtension == ".zst" || fileExtension == ".tar" || fileExtension == ".rar" {
		return true
	}
	return false
}

func IsExecutableFile(filePath string) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	filePerm := fileInfo.Mode()
	isExecutable := filePerm&0111 != 0

	return isExecutable, nil
}

func CatchAndExit(err error) {
	if err != nil {
		fmt.Println(err)
		stewPath, _ := GetStewPath()
		stewTmpPath := path.Join(stewPath, "tmp")
		err = os.RemoveAll(stewTmpPath)
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func GetStewPath() (string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	stewPath := filepath.Join(homeDir, ".stew")

	exists, err := PathExists(stewPath)
	if err != nil {
		return "", err
	} else if !exists {
		return "", StewPathNotFoundError{StewPath: stewPath}
	}

	return stewPath, nil
}

func DownloadFile(downloadPath string, url string) error {
	sp := constants.LoadingSpinner
	sp.Start()
	resp, err := http.Get(url)
	sp.Stop()

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NonZeroStatusCodeDownloadError{StatusCode: resp.StatusCode}
	}

	outputFile, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"⬇️  Downloading asset:",
	)
	_, err = io.Copy(io.MultiWriter(outputFile, bar), resp.Body)
	if err != nil {
		return err
	}

	_, err = io.Copy(outputFile, resp.Body)
	if err != nil {
		return err
	}

	return nil

}

func CopyFile(srcFile, destFile string) error {
	srcContents, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer srcContents.Close()

	destContents, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer destContents.Close()

	_, err = io.Copy(destContents, srcContents)
	if err != nil {
		return err
	}

	err = os.Chmod(destFile, 0755)
	if err != nil {
		return err
	}

	return nil
}

func WalkDir(rootDir string) ([]string, error) {
	allFilePaths := []string{}
	err := filepath.Walk(rootDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if !fileInfo.IsDir() {
			allFilePaths = append(allFilePaths, filePath)
		}
		return nil
	})
	return allFilePaths, err
}

func GetBinary(filePaths []string, repo string) (string, string, error) {
	binaryFile := ""
	binaryName := ""
	var err error
	executableFiles := []string{}
	for _, fullPath := range filePaths {
		fileNameBase := filepath.Base(fullPath)
		fileIsExecutable, err := IsExecutableFile(fullPath)
		if err != nil {
			return "", "", err
		}
		if fileNameBase == repo && fileIsExecutable {
			binaryFile = fullPath
			binaryName = repo
			executableFiles = append(executableFiles, fullPath)
		} else if filepath.Ext(fullPath) == ".exe" {
			binaryFile = fullPath
			binaryName = filepath.Base(fullPath)
			executableFiles = append(executableFiles, fullPath)
		} else if fileIsExecutable {
			executableFiles = append(executableFiles, fullPath)
		}
	}

	if binaryFile == "" {
		if len(executableFiles) == 1 {
			binaryFile = executableFiles[0]
			binaryName = filepath.Base(binaryFile)
		} else if len(executableFiles) != 1 {
			binaryFile, err = WarningPromptSelect("Could not automatically detect the binary. Please select it manually:", filePaths)
			if err != nil {
				return "", "", err
			}
			binaryName = filepath.Base(binaryFile)
		}
	}

	return binaryFile, binaryName, nil
}

func ValidateCLIInput(cliInput string) error {
	if cliInput == "" {
		return EmptyCLIInputError{}
	}

	return nil
}

type CLIInput struct {
	IsGithubInput bool
	Owner         string
	Repo          string
	Tag           string
	Asset         string
	DownloadURL   string
}

func ParseCLIInput(cliInput string) (CLIInput, error) {
	err := ValidateCLIInput(cliInput)
	if err != nil {
		return CLIInput{}, err
	}

	reGithub, err := regexp.Compile(constants.RegexGithub)
	if err != nil {
		return CLIInput{}, err
	}
	reURL, err := regexp.Compile(constants.RegexURL)
	if err != nil {
		return CLIInput{}, err
	}
	var parsedInput CLIInput
	if reGithub.MatchString(cliInput) {
		parsedInput, err = ParseGithubInput(cliInput)
	} else if reURL.MatchString(cliInput) {
		parsedInput, err = ParseURLInput(cliInput)
	} else {
		return CLIInput{}, UnrecognizedInputError{}
	}
	if err != nil {
		return CLIInput{}, err
	}

	return parsedInput, nil

}

func ParseGithubInput(cliInput string) (CLIInput, error) {
	parsedInput := CLIInput{}
	parsedInput.IsGithubInput = true
	trimmedString := strings.Trim(strings.Trim(strings.Trim(strings.TrimSpace(cliInput), "/"), "@"), "::")
	splitInput := strings.SplitN(trimmedString, "@", 2)

	ownerAndRepo := splitInput[0]
	splitOwnerAndRepo := strings.SplitN(ownerAndRepo, "/", 2)
	parsedInput.Owner = splitOwnerAndRepo[0]
	parsedInput.Repo = splitOwnerAndRepo[1]

	if len(splitInput) == 2 {
		tagAndAsset := splitInput[1]
		splitTagAndAsset := strings.SplitN(tagAndAsset, "::", 2)
		parsedInput.Tag = splitTagAndAsset[0]
		if len(splitTagAndAsset) == 2 {
			parsedInput.Asset = splitTagAndAsset[1]
		}
	}

	return parsedInput, nil

}

func ParseURLInput(cliInput string) (CLIInput, error) {
	return CLIInput{IsGithubInput: false, Asset: path.Base(cliInput), DownloadURL: cliInput}, nil
}

func Contains(slice []string, target string) (int, bool) {
	for index, element := range slice {
		if target == element {
			return index, true
		}
	}
	return -1, false
}

func GetOS() string {
	return runtime.GOOS
}

func GetArch() string {
	return runtime.GOARCH
}

func ExtractBinary(downloadedFilePath, tmpExtractionPath string) error {
	var err error
	isArchive := IsArchiveFile(downloadedFilePath)
	if isArchive {
		err = archiver.Unarchive(downloadedFilePath, tmpExtractionPath)
	} else {
		err = CopyFile(downloadedFilePath, path.Join(tmpExtractionPath, filepath.Base(downloadedFilePath)))
	}
	return err
}

func InstallBinary(downloadedFilePath string, repo string, systemInfo SystemInfo, lockFile *LockFile, overwriteFromUpgrade bool) (string, error) {

	tmpExtractionPath := systemInfo.StewTmpPath
	assetDownloadPath := systemInfo.StewPkgPath
	binaryInstallPath := systemInfo.StewBinPath

	err := ExtractBinary(downloadedFilePath, tmpExtractionPath)
	if err != nil {
		return "", err
	}

	allFilePaths, err := WalkDir(tmpExtractionPath)
	if err != nil {
		return "", err
	}

	binaryFile, binaryName, err := GetBinary(allFilePaths, repo)
	if err != nil {
		return "", err
	}

	for index, pkg := range lockFile.Packages {
		previousAssetPath := path.Join(assetDownloadPath, pkg.Asset)
		newAssetPath := downloadedFilePath
		var overwrite bool
		if pkg.Binary == binaryName {
			if !overwriteFromUpgrade {
				overwrite, err = WarningPromptConfirm(fmt.Sprintf("The binary %v version: %v is already installed, would you like to overwrite it?", constants.YellowColor(binaryName), constants.YellowColor(pkg.Tag)))
				if err != nil {
					os.RemoveAll(newAssetPath)
					return "", err
				}
			} else {
				overwrite = true
			}

			if overwrite {
				err := os.RemoveAll(previousAssetPath)
				if err != nil {
					return "", err
				}

				if !overwriteFromUpgrade {
					lockFile.Packages, err = RemovePackage(lockFile.Packages, index)
					if err != nil {
						return "", err
					}
				}

			} else {
				err = os.RemoveAll(newAssetPath)
				if err != nil {
					return "", err
				}

				err = os.RemoveAll(tmpExtractionPath)
				if err != nil {
					return "", err
				}

				return "", AbortBinaryOverwriteError{Binary: pkg.Binary}
			}
		}
	}

	err = CopyFile(binaryFile, path.Join(binaryInstallPath, binaryName))
	if err != nil {
		return "", err
	}

	err = os.RemoveAll(tmpExtractionPath)
	if err != nil {
		return "", err
	}

	return binaryName, nil
}
