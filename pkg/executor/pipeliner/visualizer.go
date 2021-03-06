package pipeliner

import (
	"bytes"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os/exec"

	"gopkg.in/yaml.v2"
)

var (
	visualizers = map[string]Visualizer{
		"svg": DotVisualizer("svg"),
		"png": DotVisualizer("png"),
		"dot": DotGrgphVisualizer,
	}
	supportedVisualizerTypes []string
)

func init() {
	for t := range visualizers {
		supportedVisualizerTypes = append(supportedVisualizerTypes, t)
	}
}

func AddVisualizer(name string, v Visualizer) error {
	if _, ok := visualizers[name]; ok {
		return fmt.Errorf("%s is a already register visualizer", name)
	}

	visualizers[name] = v

	return nil
}

func ListVisualizer() map[string]Visualizer {
	return visualizers
}

type Visualizer func(w io.Writer, pipeline *pipeliner) error

func DotVisualizer(format string) Visualizer {
	return func(w io.Writer, pipeline *pipeliner) error {
		dotFile, err := ioutil.TempFile("", "dot")
		if err != nil {
			return err
		}
		defer dotFile.Close()

		err = DotGrgphVisualizer(dotFile, pipeline)
		if err != nil {
			return err
		}

		outputFile, err := ioutil.TempFile("", format)
		if err != nil {
			return err
		}
		defer outputFile.Close()

		cmd := exec.Command("dot", "-T"+format, dotFile.Name(), "-o", outputFile.Name())
		buff := bytes.NewBuffer(nil)
		cmd.Stderr = buff
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("error: %s, stderr: %s", err, buff.String())
		}

		b, err := ioutil.ReadAll(outputFile)
		if err != nil {
			return err
		}

		_, err = w.Write(b)
		return err
	}
}

func DotGrgphVisualizer(w io.Writer, p *pipeliner) error {
	var buffer bytes.Buffer
	buffer.WriteString("digraph {\n")

	buffer.WriteString(`node [shape=plaintext fontname="Sans serif" fontsize="24"];` + "\n")

	buffer.WriteString(fmt.Sprintf(`%s [ label=<
   <table border="1" cellborder="0" cellspacing="1">`+"\n",
		p.Name(),
	))

	first := true
	p.Monitor().Do(func(root, namespace string, kv expvar.KeyValue) {
		if root != p.Name() {
			return
		}
		if first {
			first = false
			buffer.WriteString("<tr><td align=\"left\"><b>" + p.Name() + "</b></td></tr>\n")
		}

		buffer.WriteString("<tr><td align=\"left\">" + kv.Key + ":" + kv.Value.String() + "</td></tr>\n")
	})
	buffer.WriteString("</table>>];\n")
	buffer.WriteString("\n")

	for _, proc := range p.ListProcessors() {
		first := true
		hasContent := false
		p.Monitor().Do(func(root, namespace string, kv expvar.KeyValue) {
			if namespace != proc.Name {
				return
			}

			if !hasContent {
				hasContent = true
			}
			if first {
				first = false
				buffer.WriteString(fmt.Sprintf(`%s [ label=<
   <table border="1" cellborder="0" cellspacing="1">`+"\n",
					proc.Name,
				))
				buffer.WriteString("<tr><td align=\"left\"><b>" + proc.Name + "</b></td></tr>\n")
			}

			buffer.WriteString("<tr><td align=\"left\">" + kv.Key + ":" + url.QueryEscape(kv.Value.String()) + "</td></tr>\n")
		})

		if hasContent {
			buffer.WriteString("</table>>];\n")
			buffer.WriteString("\n")
		}
	}

	var sc StreamConfig
	err := yaml.Unmarshal([]byte(p.Config()), &sc)
	if err != nil {
		return err
	}
	buildRefRalationship(sc, &buffer)

	buffer.WriteString("}")
	_, err = w.Write(buffer.Bytes())
	return err
}

func buildRefRalationship(c StreamConfig, w io.Writer) {
	if c.Name == "" {
		return
	}

	for _, x := range c.Childs {
		_, _ = w.Write([]byte(fmt.Sprintf("  %s %s %s;\n", c.Name, "->", x.Name)))
		buildRefRalationship(x, w)
	}
}
