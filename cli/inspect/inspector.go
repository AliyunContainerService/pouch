package inspect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/pkg/utils/templates"

	"github.com/pkg/errors"
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
	if strings.Contains(tmplStr, ".Id") {
		tmplStr = strings.Replace(tmplStr, ".Id", ".ID", -1)
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
func Inspect(out io.Writer, refs []string, tmplStr string, getRef GetRefFunc) error {
	var errs []error

	inspector, err := NewTemplateInspectorFromString(out, tmplStr)
	if err != nil {
		return err
	}

	for _, ref := range refs {
		element, err := getRef(ref)
		if err != nil {
			errs = append(errs, errors.Errorf("Fetch object error: %v", err))
			continue
		}

		if err := inspector.Inspect(element); err != nil {
			errs = append(errs, err)
		}
	}

	if err := inspector.Flush(); err != nil {
		return err
	}

	if len(errs) == 0 {
		return nil
	}

	formatErrMsg := func(idx int, err error) (string, error) {
		errMsg := err.Error()
		errMsg = strings.TrimRight(errMsg, "\n")
		if idx != 0 {
			errMsg = fmt.Sprintf("Error: %s", errMsg)
		}
		return errMsg, nil
	}
	return utils.CombineErrors(errs, formatErrMsg)
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
	elements     []interface{}
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
	i.elements = append(i.elements, typedElement)
	return nil
}

// Flush writes the result of inspecting all elements into the output stream.
func (i *IndentedInspector) Flush() error {
	// TODO handle raw elements
	if i.elements == nil {
		_, err := io.WriteString(i.outputStream, "[]\n")
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
