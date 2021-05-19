package src

import (
	"regexp"

	"./fileparser"
	"./util"
)

//------------------------------------------------------------------------------

const constPrivateToFileComment = "//!PARANO__PRIVATE_TO_FILE"
const constLocalPrivateStuffLineRegexp1 = "\n//\\s+LOCAL PRIVATE STUFF\\s+\n"
const constLocalPrivateStuffLineRegexp2 = "\n//\\s+PRIVATE LOCAL STUFF\\s+\n"

//------------------------------------------------------------------------------

type featurePrivateToFile struct {
	locationLocalPrivateStuff int
	privateToFileDecl         map[string]bool
}

//------------------------------------------------------------------------------

func ParanoPrivateToFileInit(fileBytes []byte) *featurePrivateToFile {

	var locationLocalPrivateStuff = -1
	var loc = regexp.MustCompile(constLocalPrivateStuffLineRegexp1).FindIndex(fileBytes)
	if len(loc) > 0 {
		locationLocalPrivateStuff = loc[1]
	} else {
		loc = regexp.MustCompile(constLocalPrivateStuffLineRegexp2).FindIndex(fileBytes)
		if len(loc) > 0 {
			locationLocalPrivateStuff = loc[1]
		}
	}
	if util.IsDebug() {
		util.DebugPrintf("  locationLocalPrivateStuff: %d", locationLocalPrivateStuff)
	}
	return &featurePrivateToFile{
		locationLocalPrivateStuff: locationLocalPrivateStuff,
		privateToFileDecl:         make(map[string]bool),
	}
}

//------------------------------------------------------------------------------

func ParanoPrivateToFileVisit(n *fileparser.Node, feat *featurePrivateToFile) {

	if n.TypeStr == "Ident" && n.DepthLevel <= 4 && feat.locationLocalPrivateStuff != -1 && n.BytesIndexBegin > feat.locationLocalPrivateStuff {
		feat.privateToFileDecl[n.Name] = true
	}
	if n.IsCommentGroupWithComment(constPrivateToFileComment) && n.Father != nil {
		if n.Father.TypeStr == "GenDecl" {
			for _, n2 := range n.Father.Children {
				if n2.TypeStr == "ValueSpec" {
					if len(n2.Children) >= 2 {
						var name = n2.Children[0].Bytes
						if util.IsDebug() {
							util.DebugPrintf("....... PrivateToFile: ValueSpec: >=%s <=", name)
						}
						feat.privateToFileDecl[name] = true
						break
					}
				}
			}
		} else if n.Father.TypeStr == "FuncDecl" {
			if util.IsDebug() {
				util.DebugPrintf("....... PrivateToFile: FuncDecl: >=%s %s<=", n.Father.Name, n.Father.TypeStr)
			}
			feat.privateToFileDecl[n.Father.Name] = true
		} else {
			var nextNode = n.NextNode()
			if nextNode != nil && nextNode.TypeStr == "TypeSpec" {
				if util.IsDebug() {
					util.DebugPrintf("....... PrivateToFile: TypeSpec: >=%s %s<=", nextNode.Name, nextNode.TypeStr)
				}
				feat.privateToFileDecl[nextNode.Name] = true
			}
		}
	}
}

//------------------------------------------------------------------------------

func ParanoPrivateToFileCheck(n *fileparser.Node, featurePrivateToFile *featurePrivateToFile, filename1 string, filename2 string) (failedAtLeastOnce bool) {

	if filename1 != filename2 {
		if _, ok := featurePrivateToFile.privateToFileDecl[n.Name]; ok {
			util.NotPass("Cannot use %s in %s, declared as private to file in %s", n.Name, filename1, filename2)
			failedAtLeastOnce = true
		}
	}
	return
}

//------------------------------------------------------------------------------
