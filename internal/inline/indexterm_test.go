// Package inline tests for index term functionality.
package inline

import (
	"testing"
)

func TestFlowIndexTerm(t *testing.T) {
	source := `The Lady of the Lake held aloft Excalibur, signifying that I, ((Arthur)), was to carry Excalibur.`

	p := NewParser(source)
	nodes := p.Parse()

	if len(nodes) < 3 {
		t.Fatalf("Expected at least 3 nodes, got %d", len(nodes))
	}

	// Find the index term node
	var indexTerm *Node
	for _, node := range nodes {
		if node.Type == NodeIndexTerm {
			indexTerm = node
			break
		}
	}

	if indexTerm == nil {
		t.Fatal("No index term node found")
	}

	if indexTerm.IndexTermPrimary != "Arthur" {
		t.Errorf("Expected primary term 'Arthur', got '%s'", indexTerm.IndexTermPrimary)
	}

	if indexTerm.IndexTermConcealed {
		t.Error("Expected flow index term to NOT be concealed")
	}

	if indexTerm.Text != "Arthur" {
		t.Errorf("Expected text 'Arthur', got '%s'", indexTerm.Text)
	}
}

func TestConcealedIndexTerm(t *testing.T) {
	source := `That is why I am your king.(((Sword, Broadsword, Excalibur)))That is why!`

	p := NewParser(source)
	nodes := p.Parse()

	if len(nodes) < 2 {
		t.Fatalf("Expected at least 2 nodes, got %d", len(nodes))
	}

	// Find the index term node
	var indexTerm *Node
	for _, node := range nodes {
		if node.Type == NodeIndexTerm {
			indexTerm = node
			break
		}
	}

	if indexTerm == nil {
		t.Fatal("No index term node found")
	}

	if indexTerm.IndexTermPrimary != "Sword" {
		t.Errorf("Expected primary term 'Sword', got '%s'", indexTerm.IndexTermPrimary)
	}

	if indexTerm.IndexTermSecondary != "Broadsword" {
		t.Errorf("Expected secondary term 'Broadsword', got '%s'", indexTerm.IndexTermSecondary)
	}

	if indexTerm.IndexTermTertiary != "Excalibur" {
		t.Errorf("Expected tertiary term 'Excalibur', got '%s'", indexTerm.IndexTermTertiary)
	}

	if !indexTerm.IndexTermConcealed {
		t.Error("Expected concealed index term to be concealed")
	}
}

func TestConcealedIndexTermSingle(t *testing.T) {
	source := `To create a new git repository,(((Repository, create)))use the git init command.`

	p := NewParser(source)
	nodes := p.Parse()

	// Find the index term node
	var indexTerm *Node
	for _, node := range nodes {
		if node.Type == NodeIndexTerm {
			indexTerm = node
			break
		}
	}

	if indexTerm == nil {
		t.Fatal("No index term node found")
	}

	if indexTerm.IndexTermPrimary != "Repository" {
		t.Errorf("Expected primary term 'Repository', got '%s'", indexTerm.IndexTermPrimary)
	}

	if indexTerm.IndexTermSecondary != "create" {
		t.Errorf("Expected secondary term 'create', got '%s'", indexTerm.IndexTermSecondary)
	}

	if indexTerm.IndexTermTertiary != "" {
		t.Errorf("Expected empty tertiary term, got '%s'", indexTerm.IndexTermTertiary)
	}
}

func TestConcealedIndexTermWithCommaInTerm(t *testing.T) {
	source := `I, King Arthur.(((knight, "Arthur, King")))`

	p := NewParser(source)
	nodes := p.Parse()

	// Find the index term node
	var indexTerm *Node
	for _, node := range nodes {
		if node.Type == NodeIndexTerm {
			indexTerm = node
			break
		}
	}

	if indexTerm == nil {
		t.Fatal("No index term node found")
	}

	if indexTerm.IndexTermPrimary != "knight" {
		t.Errorf("Expected primary term 'knight', got '%s'", indexTerm.IndexTermPrimary)
	}

	if indexTerm.IndexTermSecondary != "Arthur, King" {
		t.Errorf("Expected secondary term 'Arthur, King', got '%s'", indexTerm.IndexTermSecondary)
	}
}

func TestMultipleIndexTerms(t *testing.T) {
	source := `Discussing ((Lancelot)) and (((Galahad, Knight of the Round Table))) in this text.`

	p := NewParser(source)
	nodes := p.Parse()

	var indexTerms []*Node
	for _, node := range nodes {
		if node.Type == NodeIndexTerm {
			indexTerms = append(indexTerms, node)
		}
	}

	if len(indexTerms) != 2 {
		t.Fatalf("Expected 2 index term nodes, got %d", len(indexTerms))
	}

	// First term - flow index term
	if indexTerms[0].IndexTermPrimary != "Lancelot" {
		t.Errorf("Expected first primary term 'Lancelot', got '%s'", indexTerms[0].IndexTermPrimary)
	}

	// Second term - concealed index term
	if indexTerms[1].IndexTermPrimary != "Galahad" {
		t.Errorf("Expected second primary term 'Galahad', got '%s'", indexTerms[1].IndexTermPrimary)
	}

	if indexTerms[1].IndexTermSecondary != "Knight of the Round Table" {
		t.Errorf("Expected secondary term 'Knight of the Round Table', got '%s'", indexTerms[1].IndexTermSecondary)
	}
}

func TestIndexTermStringMethod(t *testing.T) {
	tests := []struct {
		name     string
		term     *Node
		expected string
	}{
		{
			name: "flow index term",
			term: &Node{
				Type:              NodeIndexTerm,
				IndexTermPrimary:  "Arthur",
				IndexTermConcealed: false,
			},
			expected: "(Arthur)",
		},
		{
			name: "concealed single term",
			term: &Node{
				Type:              NodeIndexTerm,
				IndexTermPrimary:  "Sword",
				IndexTermConcealed: true,
			},
			expected: "((Sword))",
		},
		{
			name: "concealed two-level term",
			term: &Node{
				Type:               NodeIndexTerm,
				IndexTermPrimary:   "Sword",
				IndexTermSecondary: "Broadsword",
				IndexTermConcealed:  true,
			},
			expected: "((Sword, Broadsword))",
		},
		{
			name: "concealed three-level term",
			term: &Node{
				Type:               NodeIndexTerm,
				IndexTermPrimary:   "Sword",
				IndexTermSecondary: "Broadsword",
				IndexTermTertiary:  "Excalibur",
				IndexTermConcealed:  true,
			},
			expected: "((Sword, Broadsword, Excalibur))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.term.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
