package network

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/client"
	"github.com/docker/docker/cli"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/spf13/cobra"
)

type byNetworkName []types.NetworkResource

func (r byNetworkName) Len() int           { return len(r) }
func (r byNetworkName) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byNetworkName) Less(i, j int) bool { return r[i].Name < r[j].Name }

type listOptions struct {
	quiet   bool
	noTrunc bool
	filter  []string
}

func newListCommand(dockerCli *client.DockerCli) *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Aliases: []string{"list"},
		Short:   "罗列所有网络",
		Args:    cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(dockerCli, opts)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "仅显示网络ID")
	flags.BoolVar(&opts.noTrunc, "no-trunc", false, "不截断命令输出内容")
	flags.StringSliceVarP(&opts.filter, "filter", "f", []string{}, "提供一些过滤值(比如 'dangling=true')")

	return cmd
}

func runList(dockerCli *client.DockerCli, opts listOptions) error {
	client := dockerCli.Client()

	netFilterArgs := filters.NewArgs()
	for _, f := range opts.filter {
		var err error
		netFilterArgs, err = filters.ParseFlag(f, netFilterArgs)
		if err != nil {
			return err
		}
	}

	options := types.NetworkListOptions{
		Filters: netFilterArgs,
	}

	networkResources, err := client.NetworkList(context.Background(), options)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(dockerCli.Out(), 20, 1, 3, ' ', 0)
	if !opts.quiet {
		fmt.Fprintf(w, "网络ID\t名称\t驱动\t范围")
		fmt.Fprintf(w, "\n")
	}

	sort.Sort(byNetworkName(networkResources))
	for _, networkResource := range networkResources {
		ID := networkResource.ID
		netName := networkResource.Name
		driver := networkResource.Driver
		scope := networkResource.Scope
		if !opts.noTrunc {
			ID = stringid.TruncateID(ID)
		}
		if opts.quiet {
			fmt.Fprintln(w, ID)
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t",
			ID,
			netName,
			driver,
			scope)
		fmt.Fprint(w, "\n")
	}
	w.Flush()
	return nil
}
