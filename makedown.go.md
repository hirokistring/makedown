# makedown.go

```go
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/russross/blackfriday/v2"
)
```

## Functions

### func `isHeadingWithColon`

- parameters
  - `node *blackfriday.Node`
- returns
  - `bool`
  - `string` : target

```go
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
```

### func `findSiblingWithType`

- parameters
  - `node *blackfriday.Node`
  - `nodeType blackfriday.NodeType`
- returns
  - `bool`
  - `*blackfriday.Node` : sibling

```go
	// Find sibling with the given node type
	var sibling *blackfriday.Node = node.Prev // finding backward
	for sibling != nil {
		if sibling.Type == nodeType {
			return true, sibling
		}
		sibling = sibling.Prev
	}
	return false, sibling
```

### func `findSiblingWithTypeBefore`

- parameters
  - `node *blackfriday.Node`
  - `nodeType blackfriday.NodeType`
  - `beforeType blackfriday.NodeType`
- returns
  - `bool`
  - `*blackfriday.Node` : sibling

```go
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
```

### func `findSiblingHeadingWithColon`

- parameters
  - `node *blackfriday.Node`
- returns
  - `bool`
  - `string` : target

```go
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
```

### func `getFirstLeaf`

- parameters
  - `node *blackfriday.Node`
- returns
  - `*blackfriday.Node`

```go
	firstChild := node.FirstChild
	for firstChild != nil {
		if firstChild.IsLeaf() {
			return firstChild
		}
		firstChild = firstChild.FirstChild
	}
	return firstChild
```

### func `hasParentType`

- parameters
  - `node *blackfriday.Node`
  - `nodeType blackfriday.NodeType`
- returns
  - `bool`
  - `*blackfriday.Node`

```go
	parent := node.Parent
	for parent != nil {
		if parent.Type == nodeType {
			return true, parent
		}
		parent = parent.Parent
	}
	return false, nil
```

### func `leafWalker`

- parameters
  - `buff *strings.Builder`
  - `fmterr *error`
- returns
  - `blackfriday.NodeVisitor`

```go
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
```

> // > : prerequisites

### func `isBlockQuoteWithColon`

- parameters
  - `node *blackfriday.Node`
- returns
  - `bool`
  - `string` : target
  - `string` : quotes

```go
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
```

### func `isCodeBlockUnderHeadingWithColon`

- parameters
  - `node *blackfriday.Node`
- returns
  - `bool` : headingFound
  - `bool` : firstCodeBlock
  - `string` : target
  - `string` : recipes

```go
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
```

### func `addIndents`

- parameters
  - `recipes string`
- returns
  - `string`

```go
	recipesWithIndent := ""
	lines := strings.Split(strings.ReplaceAll(recipes, "\r\n", "\n"), "\n")
	for _, line := range lines {
		recipesWithIndent += "\t" + line + "\n"
	}
	return recipesWithIndent
```

### func `makedownWalker`

- parameters
  - `buff *strings.Builder`
  - `targets *[]string`
  - `fmterr *error`
- returns
  - `blackfriday.NodeVisitor`

```go
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
```

## func `GenerateMakefileFromMarkdown`

- parameters
  - `input_filename string`
  - `md []byte`
- returns
  - `[]byte`
  - `[]string`
  - `error`

```go
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
```
