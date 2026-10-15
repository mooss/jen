package highlight

import (
	"io"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type MarkdownStreamer struct {
	writer       io.Writer
	styleName    string
	pending      strings.Builder
	inFence      bool
	language     string
	fence        string
	formatter    chroma.Formatter
	style        *chroma.Style
	lexerCache   map[string]chroma.Lexer
	colorEnabled bool
}

func NewMarkdownStreamer(writer io.Writer, styleName string, colorEnabled bool) *MarkdownStreamer {
	return &MarkdownStreamer{
		writer:       writer,
		styleName:    styleName,
		colorEnabled: colorEnabled,
	}
}

func (s *MarkdownStreamer) ColorizeTag(tag string) string {
	if !s.colorEnabled || formatters.Get("terminal256") == nil || styles.Get(s.styleName) == nil {
		return tag
	}
	return "\x1b[1;91m" + tag + "\x1b[0m"
}

func (s *MarkdownStreamer) Write(text string) error {
	s.pending.WriteString(text)
	for {
		pending := s.pending.String()
		newline := strings.IndexByte(pending, '\n')
		if newline < 0 {
			return nil
		}

		line := pending[:newline+1]
		s.pending.Reset()
		s.pending.WriteString(pending[newline+1:])
		if err := s.writeLine(line); err != nil {
			return err
		}
	}
}

func (s *MarkdownStreamer) Flush() error {
	if s.pending.Len() == 0 {
		return nil
	}
	line := s.pending.String()
	s.pending.Reset()
	return s.writeLine(line)
}

func (s *MarkdownStreamer) writeLine(line string) error {
	trimmed := strings.TrimSpace(line)
	if marker, language, ok := markdownFence(trimmed); ok {
		if !s.inFence {
			s.inFence = true
			s.fence = marker
			s.language = language
		} else if marker == s.fence && strings.Trim(trimmed, marker[:1]) == "" {
			s.inFence = false
			s.fence = ""
			s.language = ""
		}
		return s.highlight(line, "markdown")
	}

	if s.inFence {
		return s.highlight(line, s.language)
	}
	return s.highlight(line, "markdown")
}

func (s *MarkdownStreamer) highlight(text, language string) error {
	if !s.colorEnabled {
		_, err := io.WriteString(s.writer, text)
		return err
	}

	if s.formatter == nil {
		s.formatter = formatters.Get("terminal256")
	}
	if s.style == nil {
		s.style = styles.Get(s.styleName)
	}
	if s.lexerCache == nil {
		s.lexerCache = make(map[string]chroma.Lexer)
	}
	lexer, cached := s.lexerCache[language]
	if !cached {
		lexer = lexers.Get(language)
		s.lexerCache[language] = lexer
	}
	if lexer == nil || s.formatter == nil || s.style == nil {
		_, err := io.WriteString(s.writer, text)
		return err
	}

	tokens, err := lexer.Tokenise(nil, text)
	if err != nil {
		_, fallbackErr := io.WriteString(s.writer, text)
		return fallbackErr
	}
	if err := s.formatter.Format(s.writer, s.style, tokens); err != nil {
		_, fallbackErr := io.WriteString(s.writer, text)
		return fallbackErr
	}
	return nil
}

/////////////////////
// Local functions //

func markdownFence(line string) (marker, language string, ok bool) {
	if len(line) < 3 || (line[0] != '`' && line[0] != '~') {
		return "", "", false
	}
	char := line[:1]
	count := 0
	for count < len(line) && line[count:count+1] == char {
		count++
	}
	if count < 3 {
		return "", "", false
	}

	marker = strings.Repeat(char, count)
	info := strings.TrimSpace(line[count:])
	if fields := strings.Fields(info); len(fields) > 0 {
		language = fields[0]
	}
	return marker, language, true
}
