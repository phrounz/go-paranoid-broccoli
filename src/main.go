package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"./node"
)

//------------------------------------------------------------------------------

type packageInfos struct {
	packageName      string
	packageDir       string
	subPackagesInfos []*packageInfos
	infosByFile      map[string]infosFile
}

type infosFile struct {
	packageName              string
	rootNode                 *node.Node
	featurePrivateToFile     *featurePrivateToFile
	featureExhaustiveFilling *featureExhaustiveFilling
}

var colorRed = "\033[31m"
var colorDefault = "\033[39m"

var verbose bool
var noColor bool
var debugInfo = false

//------------------------------------------------------------------------------

func main() {

	if len(os.Args) == 1 {
		usage()
	}

	var noColorPtr = flag.Bool("nocolor", false, "disable color")
	var pkgDir = flag.String("dir", "", "source directory")
	var pkg = flag.String("pkg", "", "source package")
	var verbosePtr = flag.Bool("v", false, "verbose info")
	var debug = flag.Bool("debug", false, "debug info")
	flag.Parse()
	noColor = *noColorPtr
	verbose = *verbosePtr

	debugInfo = *debug
	node.DebugInfo = debugInfo

	if *pkg != "" {
		*pkgDir = os.Getenv("GOPATH") + "/src/" + *pkg
	}

	_, err := os.Stat(*pkgDir)
	if os.IsNotExist(err) {
		panic("Folder " + *pkgDir + " does not exist.")
	}

	var rootPkg = recurseDir(*pkgDir)

	var mInfosByPackageName = make(map[string]*packageInfos)
	processPkgRecursiveAndMakeMap(rootPkg, mInfosByPackageName)

	if verbose {
		fmt.Println("\"Fourth\" pass")
	}

	processPkgAgain(mInfosByPackageName)
}

//------------------------------------------------------------------------------

func processPkgRecursiveAndMakeMap(pkgInfos *packageInfos, mInfosByPackageName map[string]*packageInfos) {
	mInfosByPackageName[pkgInfos.packageName] = pkgInfos
	for _, subPackageInfos := range pkgInfos.subPackagesInfos {
		processPkgRecursiveAndMakeMap(subPackageInfos, mInfosByPackageName)
	}
}

//------------------------------------------------------------------------------

func recurseDir(pkgDir string) *packageInfos {

	var subPackagesInfos = make([]*packageInfos, 0)

	var items, err = filepath.Glob(pkgDir + "/*")
	if err != nil {
		panic(err)
	}
	var srcFiles []string
	for _, item := range items {
		if len(item) >= 3 && item[len(item)-3:] == ".go" {
			srcFiles = append(srcFiles, item)
		} else {
			info, err := os.Stat(item)
			if os.IsNotExist(err) {
				panic("File does not exist: " + item)
			}
			if info.IsDir() {
				subPackagesInfos = append(subPackagesInfos, recurseDir(item))
			}
		}
	}
	// if len(srcFiles) == 0 {
	// 	fmt.Println("WARNING: no sources files in " + pkgDir)
	// }
	if verbose {
		fmt.Println("Processing package: " + pkgDir)
	}

	var infosByFile = processPkgFiles(srcFiles)
	var packageName string // note: remains empty string if no source files
	for _, infosFile := range infosByFile {
		packageName = infosFile.packageName
		break
	}
	return &packageInfos{
		packageName:      packageName,
		packageDir:       pkgDir,
		subPackagesInfos: subPackagesInfos,
		infosByFile:      infosByFile,
	}
}

//------------------------------------------------------------------------------

func processPkgFiles(files []string) (infosByFile map[string]infosFile) {

	infosByFile = make(map[string]infosFile)
	for _, filename := range files {
		infosByFile[filename] = processFile(filename)
	}

	if debugInfo || verbose {
		fmt.Println("  Checking ...")
	}

	if noColor {
		colorRed = ""
		colorDefault = ""
	}

	//----
	// third pass => check

	var failedAtLeastOnce = false
	for filename1, fileInfos := range infosByFile { // for each input file
		fileInfos.rootNode.Visit(func(n *node.Node) {
			if n.Name != "" {
				for filename2, fileInfos2 := range infosByFile { // for each file
					failedAtLeastOnce = paranoPrivateToFileCheck(n, fileInfos2.featurePrivateToFile, filename1, filename2) && failedAtLeastOnce
					failedAtLeastOnce = paranoExhaustiveFillingCheck(n, fileInfos2.packageName, fileInfos2.featureExhaustiveFilling, filename1, filename2) && failedAtLeastOnce
				}
			}
		})
	}
	if failedAtLeastOnce {
		fmt.Println("Stopping program now.")
		os.Exit(1)
	}

	return
}

//------------------------------------------------------------------------------

func processFile(filename string) infosFile {
	if debugInfo {
		fmt.Println("===============================")
	}
	if verbose {
		fmt.Println("  Scanning: " + filepath.Base(filename) + " ...")
	}

	//----
	// first pass => load file and get nodes

	var rootNode, fileBytes = node.ReadFile(filename)

	//----
	// init

	var featurePrivateToFile = paranoPrivateToFileInit(fileBytes)
	var featureExhaustiveFilling = paranoExhaustiveFillingInit()

	//----
	// second pass => gather informations about nodes of this file

	var packageName string
	rootNode.Visit(func(n *node.Node) {
		if debugInfo {
			fmt.Printf("==> ")
			n.Display()
		}
		if n.TypeStr == "Ident" && n.Father.TypeStr == "File" {
			packageName = n.Name
		}
		paranoPrivateToFileVisit(n, featurePrivateToFile)
		paranoExhaustiveFillingVisit(n, featureExhaustiveFilling)
	})

	var infosf = infosFile{
		packageName:              packageName,
		rootNode:                 rootNode,
		featurePrivateToFile:     featurePrivateToFile,
		featureExhaustiveFilling: featureExhaustiveFilling,
	}

	if debugInfo {
		fmt.Printf("\n\n\n%+v\n", infosf)
	}

	return infosf
}

//------------------------------------------------------------------------------

func usage() {
	fmt.Println("usage: " + os.Args[0] + " [ -dir <dir> | -pkg <pkg> ]")
	os.Exit(1)
}

//------------------------------------------------------------------------------

func processPkgAgain(mInfosByPackageName map[string]*packageInfos) {

	//----
	// fourth pass => check

	paranoExhaustiveFillingCheckGlobal(mInfosByPackageName)
}

//------------------------------------------------------------------------------
