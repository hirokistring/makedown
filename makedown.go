package makedown

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/russross/blackfriday/v2"
)

func isHeadingWithColon(node *blackfriday.Node) (bool, string) { // bool, target
	if node.Type == blackfriday.Heading {
		lastWordNode := node.LastChild
		if lastWordNode != nil {
			lastWord := string(lastWordNode.Literal)
			if strings.HasSuffix(lastWord, ":") {
				return true, lastWord
			}
		}
	}
	return false, ""
}

func findSiblingWithType(node *blackfriday.Node, nodeType blackfriday.NodeType) (bool, *blackfriday.Node) { // bool, sibling
	// Find sibling with the given node type
	var sibling *blackfriday.Node = node.Prev // finding backward
	for sibling != nil {
		if sibling.Type == nodeType {
			return true, sibling
		}
		sibling = sibling.Prev
	}
	return false, sibling
}

func findSiblingWithTypeBefore(node *blackfriday.Node, nodeType blackfriday.NodeType, beforeType blackfriday.NodeType) (bool, *blackfriday.Node) { // bool, sibling
	// Find sibling with the given node type before the given another node type
	var sibling *blackfriday.Node = node.Prev // finding backward
	for sibling != nil {
		if sibling.Type == beforeType {
			return false, sibling
		}
		if sibling.Type == nodeType {
			return true, sibling
		}
		sibling = sibling.Prev
	}
	return false, sibling
}

func findSiblingHeadingWithColon(node *blackfriday.Node) (bool, string) { // bool, target
	// Find sibling of Heading with colon
	found, heading := findSiblingWithType(node, blackfriday.Heading)
	if found {
		isHeadingWithColon, target := isHeadingWithColon(heading)
		if isHeadingWithColon {
			return true, target
		}
	}
	// fallback. header not found or header without colon is found.
	return false, ""
}

func getFirstLeaf(node *blackfriday.Node) *blackfriday.Node {
	firstChild := node.FirstChild
	for firstChild != nil {
		if firstChild.IsLeaf() {
			return firstChild
		}
		firstChild = firstChild.FirstChild
	}
	return firstChild
}

func hasParentType(node *blackfriday.Node, nodeType blackfriday.NodeType) (bool, *blackfriday.Node) {
	parent := node.Parent
	for parent != nil {
		if parent.Type == nodeType {
			return true, parent
		}
		parent = parent.Parent
	}
	return false, nil
}

func leafWalker(buff *strings.Builder, fmterr *error) blackfriday.NodeVisitor {
	return func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering {
			if node.IsLeaf() {
				literal := string(node.Literal)
				//log.Printf("[makedown] leaf %s has '%s'\n", node.String(), literal)
				fmt.Fprintf(buff, "%s", literal)
			}
		}
		return blackfriday.GoToNext
	}
}

// > : prerequisites
func isBlockQuoteWithColon(node *blackfriday.Node) (bool, string, string) { // bool, target, quotes
	if node.Type == blackfriday.BlockQuote {
		var buff strings.Builder
		var err error

		// concatenates leaf literals
		node.Walk(leafWalker(&buff, &err))
		if err != nil {
			err = fmt.Errorf("error during concatenating leaf nodes: %s", err)
			fmt.Fprintln(os.Stderr, err)
			log.Fatal(err)
		}
		prerequisites := buff.String()
		log.Printf("prerequisites: '%s'", prerequisites[1:])

		if strings.HasPrefix(prerequisites, ":") {
			// Find sibling of Heading with colon
			found, target := findSiblingHeadingWithColon(node)
			return found, target, prerequisites[1:] // trim colon at the beginning
		}
	}
	return false, "", ""
}

// ```
// recipes
// ```
func isCodeBlockUnderHeadingWithColon(node *blackfriday.Node) (bool, bool, string, string) { // headingFound, firstCodeBlock, target, recipes
	if node.Type == blackfriday.CodeBlock {
		recipes := string(node.Literal)
		//log.Printf("[makedown] recipes: '%s'", recipes)

		// Find sibling of Heading with colon
		headingFound, target := findSiblingHeadingWithColon(node)
		if headingFound {
			codeBlockFound, _ := findSiblingWithTypeBefore(node, blackfriday.CodeBlock, blackfriday.Heading)
			return true, !codeBlockFound, target, recipes
		}
	}
	return false, false, "", ""
}

func addIndents(recipes string) string {
	recipesWithIndent := ""
	lines := strings.Split(strings.ReplaceAll(recipes, "\r\n", "\n"), "\n")
	for _, line := range lines {
		recipesWithIndent += "\t" + line + "\n"
	}
	return recipesWithIndent
}

func makedownWalker(buff *strings.Builder, targets *[]string, fmterr *error) blackfriday.NodeVisitor {
	return func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering {
			// write targets and the prerequisites.
			isBlockQuoteWithColon, target, quotes := isBlockQuoteWithColon(node)
			if isBlockQuoteWithColon {
				fmt.Fprintf(buff, "%s%s\n", target, quotes)
			}

			// write targets and the recipes.
			isCodeBlockUnderHeadingWithColon, isFirstCodeBlock, target, recipes := isCodeBlockUnderHeadingWithColon(node)
			if isCodeBlockUnderHeadingWithColon {
				if target == "variables:" || target == "makedown:" {
					fmt.Fprintf(buff, "%s\n", recipes)
				} else {
					if isFirstCodeBlock {
						*targets = append(*targets, target)
						fmt.Fprintf(buff, "%s\n", target)
					}
					fmt.Fprintf(buff, "%s", addIndents(recipes))
				}
			}
		}
		return blackfriday.GoToNext
	}
}

func GenerateMakefileFromMarkdown(input_filename string, md []byte) ([]byte, []string, error) {
	var err error
	n := blackfriday.New(blackfriday.WithExtensions(blackfriday.FencedCode)).Parse(md)

	var buff strings.Builder
	var targets []string

	n.Walk(makedownWalker(&buff, &targets, &err))
	if err != nil {
		return nil, targets, fmt.Errorf("%w", err)
	}

	// Print the footer
	fmt.Fprintf(&buff, "\n# This makefile is generated by makedown from %q\n", input_filename)

	bs := bytes.NewBufferString(buff.String())
	return bs.Bytes(), targets, nil
}
