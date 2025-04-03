package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// HTMLNode represents a simplified HTML node structure for comparison
type HTMLNode struct {
	Type         string            `json:"type"`
	TagName      string            `json:"tagName,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
	TextContent  string            `json:"textContent,omitempty"`
	Children     []*HTMLNode       `json:"children,omitempty"`
	ClassList    []string          `json:"classList,omitempty"`
	ID           string            `json:"id,omitempty"`
	ComputedPath string            `json:"computedPath,omitempty"`
	SelectorPath string            `json:"selectorPath,omitempty"`
}

// ParserConfig holds configuration for the HTML parser
type ParserConfig struct {
	NormalizeWhitespace bool
}

func main() {
	// Command line flags
	inputFile := flag.String("file", "", "HTML file to parse")
	inputHTML := flag.String("html", "", "HTML string to parse")
	outputFile := flag.String("output", "", "Output file for rendered structure (defaults to stdout)")
	formatFlag := flag.String("format", "json", "Output format: json, tree")
	normalizeWhitespace := flag.Bool("normalize-ws", true, "Normalize whitespace in text nodes")
	flag.Parse()

	var r io.Reader

	// Determine input source
	if *inputFile != "" {
		f, err := os.Open(*inputFile)
		if err != nil {
			log.Fatalf("Failed to open file: %v", err)
		}
		defer f.Close()
		r = f
	} else if *inputHTML != "" {
		r = strings.NewReader(*inputHTML)
	} else {
		// Check if there's data from stdin
		stdinInfo, _ := os.Stdin.Stat()
		if (stdinInfo.Mode() & os.ModeCharDevice) == 0 {
			r = os.Stdin
		} else {
			log.Fatal("No HTML input provided. Use -file, -html, or pipe content to stdin.")
		}
	}

	// Parse HTML
	doc, err := html.Parse(r)
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	// Create parsing config
	config := ParserConfig{
		NormalizeWhitespace: *normalizeWhitespace,
	}

	// Convert to our structure
	rendered := renderNode(doc, config, "", "")

	// Determine output writer
	var out io.Writer
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer f.Close()
		out = f
	} else {
		out = os.Stdout
	}

	// Output based on selected format
	switch *formatFlag {
	case "json":
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(rendered); err != nil {
			log.Fatalf("Failed to encode as JSON: %v", err)
		}
	case "tree":
		printTree(out, rendered, 0)
	default:
		log.Fatalf("Unknown format: %s", *formatFlag)
	}
}

// renderNode converts an html.Node to our HTMLNode structure
func renderNode(n *html.Node, config ParserConfig, path string, selectorPath string) *HTMLNode {
	// Skip script and style nodes
	if shouldSkipNode(n) {
		return nil
	}

	result := &HTMLNode{}

	// Set node type
	switch n.Type {
	case html.ElementNode:
		result.Type = "element"
		result.TagName = n.DataAtom.String()
		if result.TagName == "" {
			// Custom elements will have empty DataAtom but have Data
			result.TagName = strings.ToLower(n.Data)
		}
	case html.TextNode:
		result.Type = "text"
		text := n.Data
		if config.NormalizeWhitespace {
			text = normalizeText(text)
		}
		if text == "" {
			return nil // Skip empty text nodes
		}
		result.TextContent = text
		return result // Text nodes don't have children or attributes
	case html.CommentNode:
		// Skip comments
		return nil
	case html.DocumentNode:
		result.Type = "document"
	default:
		// Skip other node types like doctypes
		return nil
	}

	// Process element attributes
	if n.Type == html.ElementNode {
		result.Attributes = make(map[string]string)
		for _, attr := range n.Attr {
			if !shouldSkipAttribute(attr.Key) {
				result.Attributes[attr.Key] = attr.Val

				// Track ID and classes separately for easy comparison
				if attr.Key == "id" {
					result.ID = attr.Val
				} else if attr.Key == "class" {
					result.ClassList = strings.Fields(attr.Val)
				}
			}
		}
	}

	// Update path for this node
	nodePath := path
	if n.Type == html.ElementNode {
		if path == "" {
			nodePath = n.DataAtom.String()
		} else {
			nodePath = path + " > " + n.DataAtom.String()
		}

		// Create a CSS selector-like path
		nodeSelector := n.DataAtom.String()
		if result.ID != "" {
			nodeSelector += "#" + result.ID
		} else if len(result.ClassList) > 0 {
			nodeSelector += "." + strings.Join(result.ClassList, ".")
		}

		if selectorPath == "" {
			result.SelectorPath = nodeSelector
		} else {
			result.SelectorPath = selectorPath + " > " + nodeSelector
		}
	}

	result.ComputedPath = nodePath

	// Process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		childNode := renderNode(c, config, nodePath, result.SelectorPath)
		if childNode != nil {
			if result.Children == nil {
				result.Children = make([]*HTMLNode, 0)
			}
			result.Children = append(result.Children, childNode)
		}
	}

	return result
}

// shouldSkipNode checks if a node should be skipped
func shouldSkipNode(n *html.Node) bool {
	if n.Type == html.ElementNode {
		tagName := strings.ToLower(n.DataAtom.String())
		if tagName == "" {
			tagName = strings.ToLower(n.Data)
		}
		return tagName == "script" || tagName == "style"
	}
	return false
}

// shouldSkipAttribute checks if an attribute should be skipped
func shouldSkipAttribute(attrName string) bool {
	attrName = strings.ToLower(attrName)
	return strings.HasPrefix(attrName, "data-") || strings.HasPrefix(attrName, "aria-")
}

// normalizeText removes extra whitespace from text
func normalizeText(s string) string {
	// Replace all whitespace sequences with a single space
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}

// printTree prints the HTML structure as a tree
func printTree(w io.Writer, node *HTMLNode, level int) {
	indent := strings.Repeat("  ", level)

	switch node.Type {
	case "element":
		fmt.Fprintf(w, "%s<%s", indent, node.TagName)

		// Print attributes
		if node.ID != "" {
			fmt.Fprintf(w, " id=\"%s\"", node.ID)
		}
		if len(node.ClassList) > 0 {
			fmt.Fprintf(w, " class=\"%s\"", strings.Join(node.ClassList, " "))
		}

		for k, v := range node.Attributes {
			if k != "id" && k != "class" {
				fmt.Fprintf(w, " %s=\"%s\"", k, v)
			}
		}

		if len(node.Children) == 0 {
			fmt.Fprintln(w, "/>")
		} else {
			fmt.Fprintln(w, ">")

			// Print children
			for _, child := range node.Children {
				printTree(w, child, level+1)
			}

			fmt.Fprintf(w, "%s</%s>\n", indent, node.TagName)
		}
	case "text":
		fmt.Fprintf(w, "%s\"%s\"\n", indent, node.TextContent)
	case "document":
		for _, child := range node.Children {
			printTree(w, child, level)
		}
	}
}

// CompareHTMLNodes compares two HTML node structures
func CompareHTMLNodes(a, b *HTMLNode) (bool, []string) {
	differences := []string{}

	// Check node type
	if a.Type != b.Type {
		differences = append(differences, fmt.Sprintf("Node type mismatch: %s vs %s at %s",
			a.Type, b.Type, a.ComputedPath))
		return false, differences
	}

	// For text nodes, compare content
	if a.Type == "text" {
		if a.TextContent != b.TextContent {
			differences = append(differences, fmt.Sprintf("Text content mismatch at %s: \"%s\" vs \"%s\"",
				a.ComputedPath, a.TextContent, b.TextContent))
		}
		return len(differences) == 0, differences
	}

	// For element nodes, compare tag name and attributes
	if a.Type == "element" {
		if a.TagName != b.TagName {
			differences = append(differences, fmt.Sprintf("Tag name mismatch at %s: %s vs %s",
				a.ComputedPath, a.TagName, b.TagName))
		}

		// Compare IDs
		if a.ID != b.ID {
			differences = append(differences, fmt.Sprintf("ID mismatch at %s: %s vs %s",
				a.ComputedPath, a.ID, b.ID))
		}

		// Compare classes (order-independent)
		if !equalStringSlices(a.ClassList, b.ClassList) {
			differences = append(differences, fmt.Sprintf("Class list mismatch at %s: %v vs %v",
				a.ComputedPath, a.ClassList, b.ClassList))
		}

		// Compare attributes
		for k, v := range a.Attributes {
			if k != "id" && k != "class" {
				if bv, ok := b.Attributes[k]; !ok {
					differences = append(differences, fmt.Sprintf("Missing attribute %s at %s",
						k, a.ComputedPath))
				} else if v != bv {
					differences = append(differences, fmt.Sprintf("Attribute %s mismatch at %s: %s vs %s",
						k, a.ComputedPath, v, bv))
				}
			}
		}

		for k := range b.Attributes {
			if k != "id" && k != "class" {
				if _, ok := a.Attributes[k]; !ok {
					differences = append(differences, fmt.Sprintf("Extra attribute %s at %s",
						k, b.ComputedPath))
				}
			}
		}

		// Compare children
		if len(a.Children) != len(b.Children) {
			differences = append(differences, fmt.Sprintf("Children count mismatch at %s: %d vs %d",
				a.ComputedPath, len(a.Children), len(b.Children)))
		} else {
			for i := range a.Children {
				if i < len(b.Children) {
					equal, childDiffs := CompareHTMLNodes(a.Children[i], b.Children[i])
					if !equal {
						differences = append(differences, childDiffs...)
					}
				}
			}
		}
	}

	return len(differences) == 0, differences
}

// equalStringSlices compares two string slices, ignoring order
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	countA := make(map[string]int)
	countB := make(map[string]int)

	for _, s := range a {
		countA[s]++
	}

	for _, s := range b {
		countB[s]++
	}

	for k, v := range countA {
		if countB[k] != v {
			return false
		}
	}

	return true
}

// GetHTMLNodeBySelector finds an HTML node by CSS selector
func GetHTMLNodeBySelector(root *HTMLNode, selector string) []*HTMLNode {
	parts := strings.Split(selector, " ")
	return findNodesBySelectorParts(root, parts, 0)
}

// findNodesBySelectorParts recursively searches for nodes matching a selector path
func findNodesBySelectorParts(node *HTMLNode, parts []string, depth int) []*HTMLNode {
	if depth >= len(parts) {
		return []*HTMLNode{}
	}

	currentPart := parts[depth]
	results := []*HTMLNode{}

	// Check if current node matches
	if matchesSelector(node, currentPart) {
		if depth == len(parts)-1 {
			// This is the final part of the selector path
			results = append(results, node)
		} else {
			// Continue matching with children
			for _, child := range node.Children {
				results = append(results, findNodesBySelectorParts(child, parts, depth+1)...)
			}
		}
	}

	// Continue searching in all children
	for _, child := range node.Children {
		results = append(results, findNodesBySelectorParts(child, parts, depth)...)
	}

	return results
}

// matchesSelector checks if a node matches a simple selector
func matchesSelector(node *HTMLNode, selector string) bool {
	if node.Type != "element" {
		return false
	}

	// Handle simple tag selector
	if selector == node.TagName {
		return true
	}

	// Handle ID selector (#id)
	if strings.HasPrefix(selector, "#") {
		id := strings.TrimPrefix(selector, "#")
		return node.ID == id
	}

	// Handle class selector (.class)
	if strings.HasPrefix(selector, ".") {
		class := strings.TrimPrefix(selector, ".")
		for _, c := range node.ClassList {
			if c == class {
				return true
			}
		}
	}

	return false
}
