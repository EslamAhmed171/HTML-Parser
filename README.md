# HTML Parser for Keploy

A lightweight, efficient HTML parser and comparison tool designed for testing HTML responses in Keploy.

## Overview

This tool parses HTML content and converts it into a structured JSON representation, making it easy to compare HTML responses during API testing. It's specifically designed to work with Keploy for comparing HTML responses during test verification.

Key features:
- Converts HTML documents to a normalized JSON structure
- Provides tree visualization for HTML structure
- Supports comparison between HTML structures
- Handles CSS selector queries for finding specific elements
- Normalizes whitespace for consistent comparisons

## Installation

```bash
go get github.com/yourusername/html-parser
```

Or clone the repository:

```bash
git clone https://github.com/yourusername/html-parser.git
cd html-parser
go build
```

## Usage

### Basic Usage

Parse an HTML file and output JSON:

```bash
./html-parser -file example.html
```

Parse HTML directly from a string:

```bash
./html-parser -html "<html><body><h1>Hello World</h1></body></html>"
```

Pipe HTML content:

```bash
cat example.html | ./html-parser
```

### Output Options

Save output to a file:

```bash
./html-parser -file example.html -output parsed.json
```

Display as a tree instead of JSON:

```bash
./html-parser -file example.html -format tree
```

### Whitespace Handling

Disable whitespace normalization:

```bash
./html-parser -file example.html -normalize-ws=false
```

## Structure Format

The HTML is converted to a structured format that looks like this:

```json
{
  "type": "element",
  "tagName": "html",
  "attributes": {},
  "children": [
    {
      "type": "element",
      "tagName": "body",
      "attributes": {},
      "children": [
        {
          "type": "element",
          "tagName": "h1",
          "attributes": {},
          "textContent": "Hello World",
          "children": []
        }
      ]
    }
  ]
}
```

## Example

### Input HTML
```html
<!DOCTYPE html>
<html>
<head>
    <title>Example Page</title>
</head>
<body>
    <div id="main" class="container">
        <h1>Hello World</h1>
        <p>This is a paragraph</p>
    </div>
</body>
</html>
```

### Output JSON
```json
{
  "type": "document",
  "children": [
    {
      "type": "element",
      "tagName": "html",
      "attributes": {},
      "children": [
        {
          "type": "element",
          "tagName": "head",
          "attributes": {},
          "children": [
            {
              "type": "element",
              "tagName": "title",
              "attributes": {},
              "children": [
                {
                  "type": "text",
                  "textContent": "Example Page"
                }
              ]
            }
          ]
        },
        {
          "type": "element",
          "tagName": "body",
          "attributes": {},
          "children": [
            {
              "type": "element",
              "tagName": "div",
              "attributes": {
                "id": "main",
                "class": "container"
              },
              "id": "main",
              "classList": [
                "container"
              ],
              "children": [
                {
                  "type": "element",
                  "tagName": "h1",
                  "attributes": {},
                  "children": [
                    {
                      "type": "text",
                      "textContent": "Hello World"
                    }
                  ]
                },
                {
                  "type": "element",
                  "tagName": "p",
                  "attributes": {},
                  "children": [
                    {
                      "type": "text",
                      "textContent": "This is a paragraph"
                    }
                  ]
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
