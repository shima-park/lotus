package lotusctl

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/shima-park/lotus/pkg/rpc/client"
)

func newClient(hosts ...string) *client.Client {
	var host string
	if len(host) > 0 {
		host = hosts[0]
	} else if envHosts := os.Getenv("LOTUS_SERVER_ENV"); envHosts != "" {
		host = envHosts
	} else {
		host = "localhost:8080"
	}

	return client.NewClient(host)
}

func renderTable(header []string, rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func stringInSlice(t string, ss []string) bool {
	for _, s := range ss {
		if t == s {
			return true
		}
	}
	return false
}
