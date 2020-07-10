package pipeliner

import (
	"bytes"
	"expvar"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/shima-park/lotus/pkg/core/common/inject"
	"io"
	"io/ioutil"
	"net/url"
	"os/exec"
)

var (
	visualizers = map[string]Visualizer{
		"svg":  DotVisualizer("svg"),
		"png":  DotVisualizer("png"),
		"dot":  DotGrgphVisualizer,
		"term": TermVisualizer,
	}
	supportedVisualizerTypes []string
)

func init() {
	for t := range visualizers {
		supportedVisualizerTypes = append(supportedVisualizerTypes, t)
	}
}

type Visualizer func(w io.Writer, pipeline *pipeliner) error

func AddVisualizer(name string, v Visualizer) error {
	if _, ok := visualizers[name]; ok {
		return fmt.Errorf("%s is a already register visualizer", name)
	}

	visualizers[name] = v
	supportedVisualizerTypes = append(supportedVisualizerTypes, name)

	return nil
}

func ListVisualizer() map[string]Visualizer {
	return visualizers
}

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
		p.name,
	))

	first := true
	p.monitor.Do(func(root, namespace string, kv expvar.KeyValue) {
		if root != p.name {
			return
		}
		if first {
			first = false
			buffer.WriteString("<tr><td align=\"left\"><b>" + p.name + "</b></td></tr>\n")
		}

		buffer.WriteString("<tr><td align=\"left\">" + kv.Key + ":" + kv.Value.String() + "</td></tr>\n")
	})
	buffer.WriteString("</table>>];\n")
	buffer.WriteString("\n")

	for _, proc := range p.processors {
		first := true
		hasContent := false
		p.monitor.Do(func(root, namespace string, kv expvar.KeyValue) {
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
	if p.stream != nil {
		sc = p.stream.config
	}
	buildRefRalationship(sc, &buffer)

	buffer.WriteString("}")
	_, err := w.Write(buffer.Bytes())
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

func TermVisualizer(w io.Writer, pipe *pipeliner) error {
	printExecutorComponents(w, pipe)
	printExecutorProcessors(w, pipe)
	return nil
}

func printExecutorComponents(w io.Writer, p *pipeliner) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{
		"Component Name", "Reflect Type", "Inject Name", "Raw Config", "Description",
	})
	table.SetRowLine(true)
	for _, c := range p.components {
		arr := []string{
			c.Name,
			c.Component.Instance().Type().String(),
			c.Component.Instance().Name(),
			c.RawConfig,
			c.Factory.Description(),
		}

		table.Rich(arr, []tablewriter.Colors{
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{tablewriter.Normal, tablewriter.FgGreenColor},
			tablewriter.Colors{},
			tablewriter.Colors{},
		})

	}
	table.Render()
}

func printExecutorProcessors(w io.Writer, p *pipeliner) {
	mdeErrs := filterMissingDependencyError(p.CheckDependence())

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Processor Name", //"Config",
		"Struct name", "Field", "Reflect type", "Inject name"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	for _, p := range p.processors {
		requests, responses := inject.GetFuncReqAndRespReceptorList(p.Processor)

		for _, req := range requests {
			mdeErr := matchError(mdeErrs, req)
			if mdeErr != nil {
				table.Rich(
					[]string{p.Name, req.StructName, req.StructFieldName, req.ReflectType, req.InjectName},
					[]tablewriter.Colors{
						tablewriter.Colors{},
						tablewriter.Colors{},
						tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
						tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
						tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
					})
			} else {
				table.Rich(
					[]string{p.Name, req.StructName, req.StructFieldName, req.ReflectType, req.InjectName},
					[]tablewriter.Colors{
						tablewriter.Colors{},
						tablewriter.Colors{},
						tablewriter.Colors{},
						tablewriter.Colors{},
						tablewriter.Colors{tablewriter.Normal, tablewriter.FgCyanColor},
					})
			}
		}

		for _, resp := range responses {
			table.Rich(
				[]string{p.Name, resp.StructName, resp.StructFieldName, resp.ReflectType, resp.InjectName},
				[]tablewriter.Colors{
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{tablewriter.Normal, tablewriter.FgGreenColor},
				})
		}
	}
	table.Render()
}

func filterMissingDependencyError(errs []error) []MissingDependencyError {
	var mdeErrs []MissingDependencyError
	for _, err := range errs {
		cause, ok := errors.Cause(err).(MissingDependencyError)
		if ok {
			mdeErrs = append(mdeErrs, cause)
		}
	}
	return mdeErrs
}

func matchError(mdeErrs []MissingDependencyError, r inject.Receptor) *MissingDependencyError {
	for _, mdeErr := range mdeErrs {
		if mdeErr.Field == r.StructFieldName &&
			mdeErr.ReflectType == r.ReflectType &&
			mdeErr.InjectName == r.InjectName {
			return &mdeErr
		}
	}
	return nil
}
