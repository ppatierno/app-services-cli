package list

import (
	"context"

	"github.com/redhat-developer/app-services-cli/pkg/cmd/registry/artifact/util"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/registry/registrycmdutil"

	"github.com/redhat-developer/app-services-cli/pkg/core/cmdutil/factory"
	"github.com/redhat-developer/app-services-cli/pkg/core/cmdutil/flagutil"
	"github.com/redhat-developer/app-services-cli/pkg/core/config"
	"github.com/redhat-developer/app-services-cli/pkg/core/connection"
	"github.com/redhat-developer/app-services-cli/pkg/core/ioutil/dump"
	"github.com/redhat-developer/app-services-cli/pkg/core/ioutil/iostreams"
	"github.com/redhat-developer/app-services-cli/pkg/core/localize"
	"github.com/redhat-developer/app-services-cli/pkg/core/logging"
	registryinstanceclient "github.com/redhat-developer/app-services-sdk-go/registryinstance/apiv1internal/client"

	"github.com/spf13/cobra"
)

// row is the details of a Service Registry instance needed to print to a table
type artifactRow struct {
	// The ID of a single artifact.
	Id string `json:"id" header:"ID"`

	Name string `json:"name,omitempty" header:"Name"`

	CreatedOn string `json:"createdOn" header:"Created on"`

	CreatedBy string `json:"createdBy" header:"Created By"`

	Type registryinstanceclient.ArtifactType `json:"type" header:"Type"`

	State registryinstanceclient.ArtifactState `json:"state" header:"State"`
}

type options struct {
	group string

	registryID   string
	outputFormat string
	name         string
	description  string
	labels       []string
	properties   []string

	page  int32
	limit int32

	IO         *iostreams.IOStreams
	Config     config.IConfig
	Connection factory.ConnectionFunc
	Logger     logging.Logger
	localizer  localize.Localizer
	Context    context.Context
}

// NewListCommand creates a new command for listing registry artifacts.
func NewListCommand(f *factory.Factory) *cobra.Command {
	opts := &options{
		Config:     f.Config,
		Connection: f.Connection,
		Logger:     f.Logger,
		IO:         f.IOStreams,
		localizer:  f.Localizer,
		Context:    f.Context,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   f.Localizer.MustLocalize("artifact.cmd.list.description.short"),
		Long:    f.Localizer.MustLocalize("artifact.cmd.list.description.long"),
		Example: f.Localizer.MustLocalize("artifact.cmd.list.example"),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.outputFormat != "" && !flagutil.IsValidInput(opts.outputFormat, flagutil.ValidOutputFormats...) {
				return flagutil.InvalidValueError("output", opts.outputFormat, flagutil.ValidOutputFormats...)
			}

			if opts.page < 1 || opts.limit < 1 {
				return opts.localizer.MustLocalizeError("artifact.common.error.page.and.limit.too.small")
			}

			if opts.registryID != "" {
				return runList(opts)
			}

			cfg, err := opts.Config.Load()
			if err != nil {
				return err
			}

			instanceID, ok := cfg.GetServiceRegistryIdOk()
			if !ok {
				return opts.localizer.MustLocalizeError("artifact.cmd.common.error.noServiceRegistrySelected")
			}

			opts.registryID = instanceID

			return runList(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.group, "group", "g", util.DefaultArtifactGroup, opts.localizer.MustLocalize("artifact.common.group"))
	cmd.Flags().Int32VarP(&opts.page, "page", "", 1, opts.localizer.MustLocalize("artifact.common.page.number"))
	cmd.Flags().Int32VarP(&opts.limit, "limit", "", 100, opts.localizer.MustLocalize("artifact.common.page.limit"))

	cmd.Flags().StringVar(&opts.name, "name", "", opts.localizer.MustLocalize("artifact.cmd.list.flag.name.description"))
	cmd.Flags().StringArrayVar(&opts.labels, "label", []string{}, opts.localizer.MustLocalize("artifact.cmd.list.flag.labels.description"))
	cmd.Flags().StringVar(&opts.description, "description", "", opts.localizer.MustLocalize("artifact.cmd.list.flag.description.description"))
	cmd.Flags().StringArrayVar(&opts.properties, "property", []string{}, opts.localizer.MustLocalize("artifact.cmd.list.flag.properties.description"))

	cmd.Flags().StringVar(&opts.registryID, "instance-id", "", opts.localizer.MustLocalize("artifact.common.instance.id"))
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "", opts.localizer.MustLocalize("artifact.common.message.output.format"))

	flagutil.EnableOutputFlagCompletion(cmd)

	return cmd
}

func runList(opts *options) error {
	if opts.group == util.DefaultArtifactGroup {
		opts.Logger.Info(opts.localizer.MustLocalize("artifact.common.message.no.group", localize.NewEntry("DefaultArtifactGroup", util.DefaultArtifactGroup)))
		opts.group = util.DefaultArtifactGroup
	}

	conn, err := opts.Connection(connection.DefaultConfigRequireMasAuth)
	if err != nil {
		return err
	}

	api := conn.API()

	a, _, err := api.ServiceRegistryInstance(opts.registryID)
	if err != nil {
		return err
	}
	request := a.ArtifactsApi.SearchArtifacts(opts.Context)

	request = request.Group(opts.group)
	request = request.Offset((opts.page - 1) * opts.limit)
	request = request.Limit(opts.limit)
	request = request.Orderby(registryinstanceclient.SORTBY_CREATED_ON)
	request = request.Order(registryinstanceclient.SORTORDER_ASC)

	if opts.name != "" {
		request = request.Name(opts.name)
	}

	if len(opts.labels) > 0 {
		request = request.Labels(opts.labels)
	}

	if opts.description != "" {
		request = request.Description(opts.description)
	}

	if len(opts.properties) > 0 {
		request = request.Properties(opts.properties)
	}

	response, _, err := request.Execute()
	if err != nil {
		return registrycmdutil.TransformInstanceError(err)
	}

	if len(response.Artifacts) == 0 && opts.outputFormat == "" {
		opts.Logger.Info(opts.localizer.MustLocalize("artifact.common.message.no.artifact.available.for.group.and.registry", localize.NewEntry("Group", opts.group), localize.NewEntry("Registry", opts.registryID)))
		return nil
	}

	switch opts.outputFormat {
	case dump.EmptyFormat:
		rows := mapResponseItemsToRows(response.Artifacts)
		dump.Table(opts.IO.Out, rows)
		opts.Logger.Info("")
	default:
		return dump.Formatted(opts.IO.Out, opts.outputFormat, response)
	}

	return nil
}

func mapResponseItemsToRows(artifacts []registryinstanceclient.SearchedArtifact) []artifactRow {
	rows := make([]artifactRow, len(artifacts))

	for i := range artifacts {
		k := (artifacts)[i]
		row := artifactRow{
			Id:        k.GetId(),
			Name:      k.GetName(),
			CreatedOn: k.GetCreatedOn(),
			CreatedBy: k.GetCreatedBy(),
			Type:      k.GetType(),
			State:     k.GetState(),
		}

		rows[i] = row
	}

	return rows
}
