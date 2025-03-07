package acl

import (
	"github.com/redhat-developer/app-services-cli/pkg/cmd/kafka/acl/admin"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/kafka/acl/create"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/kafka/acl/delete"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/kafka/acl/grant"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/kafka/acl/list"
	"github.com/redhat-developer/app-services-cli/pkg/core/cmdutil/factory"
	"github.com/spf13/cobra"
)

// NewAclCommand creates a new command sub-group for Kafka ACL operations
func NewAclCommand(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "acl",
		Short:   f.Localizer.MustLocalize("kafka.acl.cmd.shortDescription"),
		Long:    f.Localizer.MustLocalize("kafka.acl.cmd.longDescription"),
		Example: f.Localizer.MustLocalize("kafka.acl.cmd.example"),
		Args:    cobra.ExactArgs(1),
	}

	cmd.AddCommand(
		list.NewListACLCommand(f),
		grant.NewGrantPermissionsACLCommand(f),
		delete.NewDeleteCommand(f),
		admin.NewAdminACLCommand(f),
		create.NewCreateCommand(f),
	)

	return cmd
}
