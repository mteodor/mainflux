// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	mfxsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/spf13/cobra"
)

var cmdGroups = []cobra.Command{
	cobra.Command{
		Use:   "create",
		Short: "create <name> <description> <user_auth_token>",
		Long:  `Creates new group`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Short)
				return
			}

			group := mfxsdk.Group{
				Name:        args[0],
				Description: args[1],
			}
			id, err := sdk.CreateGroup(group, args[2])
			if err != nil {
				logError(err)
				return
			}
			logCreated(id)
		},
	},
	cobra.Command{
		Use:   "get",
		Short: "get [all | <group_name>] <user_auth_token>",
		Long:  `Get all groups or group by name`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Short)
				return
			}

			if args[0] == "all" {
				l, err := sdk.Groups(args[1], uint64(Offset), uint64(Limit), Name)
				if err != nil {
					logError(err)
					return
				}
				logJSON(l)
				return
			}

			t, err := sdk.Group(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(t)
		},
	},
}

// NewGroupsCmd returns users command.
func NewGroupsCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "groups",
		Short: "Groups management",
		Long:  `Groups management: create accounts and tokens"`,
		Run: func(cmd *cobra.Command, args []string) {
			logUsage("Usage: Groups [create | get | delete]")
		},
	}

	for i := range cmdGroups {
		cmd.AddCommand(&cmdGroups[i])
	}

	return &cmd
}
