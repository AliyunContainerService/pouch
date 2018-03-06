package inspect

import (
	"bytes"
	"encoding/json"
	"io"
	"text/template"

	"github.com/alibaba/pouch/pkg/utils/templates"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Inspector defines an interface to implement to process elements.
type Inspector interface {
	Inspect(typedElement interface{}) error
	Flush() error
}

// TemplateInspector uses a template to inspect elements.
type TemplateInspector struct {
	outputStream io.Writer
	buffer       *bytes.Buffer
	tmpl         *template.Template
}

// NewTemplateInspector creates a new inspector with a template.
func NewTemplateInspector(out io.Writer, tmpl *template.Template) Inspector {
	return &TemplateInspector{
		outputStream: out,
		buffer:       new(bytes.Buffer),
		tmpl:         tmpl,
	}
}

// NewTemplateInspectorFromString creates a new TemplateInspector from a string.
func NewTemplateInspectorFromString(out io.Writer, tmplStr string) (Inspector, error) {
	if tmplStr == "" {
		return NewIndentedInspector(out), nil
	}

	tmpl, err := templates.Parse(tmplStr)
	if err != nil {
		return nil, errors.Errorf("Parse template String error: %s", err)
	}
	return NewTemplateInspector(out, tmpl), nil
}

// GetRefFunc is a function which used by Inspect to fetch an object from
// a reference
type GetRefFunc func(ref string) (interface{}, error)

// Inspect fetches objects by reference and writes the json representation
// to the output writer.
func Inspect(out io.Writer, ref string, tmplStr string, getRef GetRefFunc) error {
	inspector, err := NewTemplateInspectorFromString(out, tmplStr)
	if err != nil {
		return err
	}

	element, err := getRef(ref)
	if err != nil {
		return errors.Errorf("Template parsing error: %v", err)
	}

	if err := inspector.Inspect(element); err != nil {
		return err
	}

	if err := inspector.Flush(); err != nil {
		logrus.Errorf("%s\n", err)
	}

	return nil
}

// Inspect executes the inspect template.
func (i *TemplateInspector) Inspect(typedElement interface{}) error {
	buf := new(bytes.Buffer)
	if err := i.tmpl.Execute(buf, typedElement); err != nil {
		return errors.Errorf("Template parsing error: %v", err)
	}
	i.buffer.Write(buf.Bytes())
	i.buffer.WriteByte('\n')
	return nil
}

// Flush writes the result of inspecting all elements into output stream.
func (i *TemplateInspector) Flush() error {
	if i.buffer.Len() == 0 {
		_, err := io.WriteString(i.outputStream, "\n")
		return err
	}

	_, err := io.Copy(i.outputStream, i.buffer)
	return err
}

// IndentedInspector uses a buffer to stop the indented representation of an element.
type IndentedInspector struct {
	outputStream io.Writer
	elements     interface{}
	rawElements  [][]byte
}

// NewIndentedInspector generates a new IndentedInspector.
func NewIndentedInspector(outputStream io.Writer) Inspector {
	return &IndentedInspector{
		outputStream: outputStream,
	}
}

// Inspect writes the raw element with an indented json format.
func (i *IndentedInspector) Inspect(typedElement interface{}) error {
	// TODO handle raw elements
	i.elements = typedElement
	return nil
}

// Flush writes the result of inspecting all elements into the output stream.
func (i *IndentedInspector) Flush() error {
	// TODO handle raw elements
	if i.elements == nil {
		_, err := io.WriteString(i.outputStream, "\n")
		return err
	}

	var buffer io.Reader
	b, err := json.MarshalIndent(i.elements, "", "    ")
	if err != nil {
		return err
	}
	buffer = bytes.NewReader(b)

	if _, err := io.Copy(i.outputStream, buffer); err != nil {
		return err
	}
	_, err = io.WriteString(i.outputStream, "\n")
	return err
}
