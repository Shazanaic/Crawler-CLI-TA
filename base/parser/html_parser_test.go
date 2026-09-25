package parser

import (
	"testing"
)

func TestHTMLParser_Parse(t *testing.T) {
	htmlData := []byte(`
	<html>
		<head>
			<title>Test Page</title>
		</head>
		<body>
			<h1>Welcome to the Test Page</h1>
			<p>This is a simple test page.</p>
			<a href="https://example.com">Example</a>
			<a href="https://test.com">Test</a>
		</body>
	</html>`)

	parser := NewHTMLParser()
	result, err := parser.Parse(htmlData)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	if result.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", result.Title)
	}
	expectedURLs := []string{"https://example.com", "https://test.com"}
	if len(result.URLs) != len(expectedURLs) {
		t.Fatalf("Expected %d URLs, got %d", len(expectedURLs), len(result.URLs))
	}
	for i, expectedURL := range expectedURLs {
		if result.URLs[i] != expectedURL {
			t.Errorf("Expected URL '%s', got '%s'", expectedURL, result.URLs[i])
		}
	}
}

func TestHTMLParser_Parse_Empty(t *testing.T) {
	htmlData := []byte(`
	<html>
		<head>
		</head>
		<body>
		</body>
	</html>`)

	parser := NewHTMLParser()
	result, err := parser.Parse(htmlData)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	if result.Title != "" {
		t.Errorf("Expected empty title, got '%s'", result.Title)
	}
	if len(result.URLs) != 0 {
		t.Errorf("Expected 0 URLs, got %d", len(result.URLs))
	}
}

func TestHTMLParser_IgnoresNotHrefLinks(t *testing.T) {
	htmlData := []byte(`
	<html>
		<head>
			<title>Test Page</title>
		</head>
		<body>
			<a>Link without href</a>
			<a href="">Empty href</a>
			<a href="https://example.com">Example</a>
			<a href="https://test.com">Test</a>
			<a>Link without href2</a>
		</body>
	</html>`)

	parser := NewHTMLParser()
	result, err := parser.Parse(htmlData)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	if result.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", result.Title)
	}
	expectedURLs := []string{"https://example.com", "https://test.com"}
	if len(result.URLs) != len(expectedURLs) {
		t.Fatalf("Expected %d URLs, got %d", len(expectedURLs), len(result.URLs))
	}
	for i, expectedURL := range expectedURLs {
		if result.URLs[i] != expectedURL {
			t.Errorf("Expected URL '%s', got '%s'", expectedURL, result.URLs[i])
		}
	}
}
