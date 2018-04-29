package wordsfilter

import (
	"sync"
	"strings"
	"bytes"
	"os"
	"bufio"
	"io"
)

type WordsFilter struct {
	placeholder string
	stripSpace  bool
	node        *Node
	mutex       sync.RWMutex
}

// New creates a filter.
func NewWordsFilter(placeholder string, stripSpace bool) *WordsFilter {
	return &WordsFilter{
		placeholder: placeholder,
		stripSpace:  stripSpace,
		node:        NewNode(make(map[string]*Node), ""),
	}
}

// Convert sensitive text lists into sensitive word tree nodes
func (wf *WordsFilter) Generate(texts []string) map[string]*Node {
	root := make(map[string]*Node)
	for _, text := range texts {
		wf.Add(text, root)
	}
	return root
}

// Convert sensitive text from file into sensitive word tree nodes.
// File content format, please wrap every sensitive word.
func (wf *WordsFilter) GenerateWithFile(path string) (map[string]*Node, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	buf := bufio.NewReader(fd)
	var texts []string
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		text := strings.Trim(string(line), " \\t\\n\\r\\0\\x0B")
		if text == "" {
			continue
		}
		texts = append(texts, text)
	}

	root := wf.Generate(texts)
	return root, nil
}

// Add sensitive words to specified sensitive words Map.
func (wf *WordsFilter) Add(text string, root map[string]*Node) {
	if wf.stripSpace {
		text = stripSpace(text)
	}
	wf.mutex.Lock()
	defer wf.mutex.Unlock()
	wf.node.Add(text, root, wf.placeholder)
}

// Replace sensitive words in strings and return new strings.
func (wf *WordsFilter) Replace(text string, root map[string]*Node) string {
	if wf.stripSpace {
		text = stripSpace(text)
	}
	wf.mutex.RLock()
	defer wf.mutex.RUnlock()
	return wf.node.Replace(text, root)
}

// Whether the string contains sensitive words.
func (wf *WordsFilter) Contains(text string, root map[string]*Node) bool {
	if wf.stripSpace {
		text = stripSpace(text)
	}
	wf.mutex.RLock()
	defer wf.mutex.RUnlock()
	return wf.node.Contains(text, root)
}

// Remove specified sensitive words from sensitive word map.
func (wf *WordsFilter) Remove(text string, root map[string]*Node) {
	if wf.stripSpace {
		text = stripSpace(text)
	}
	wf.mutex.Lock()
	defer wf.mutex.Unlock()
	wf.node.Remove(text, root)
}

// Strip space
func stripSpace(str string) string {
	fields := strings.Fields(str)
	var bf bytes.Buffer
	for _, field := range fields {
		bf.WriteString(field)
	}
	return bf.String()
}
