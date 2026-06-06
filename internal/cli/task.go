package cli

import (
	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/spf13/cobra"
)

func newTaskCommand(opts *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks on the agent (get, list, cancel)",
	}
	cmd.AddCommand(
		newTaskGetCommand(opts),
		newTaskListCommand(opts),
		newTaskCancelCommand(opts),
		newTaskSubscribeCommand(opts),
	)
	return cmd
}

func newTaskGetCommand(opts *globalOptions) *cobra.Command {
	var historyLength int
	cmd := &cobra.Command{
		Use:   "get [task-id]",
		Short: "Fetch a task by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, _, err := dial(ctx, opts, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer client.Destroy()

			req := &a2a.GetTaskRequest{Tenant: opts.tenant, ID: a2a.TaskID(args[0])}
			if cmd.Flags().Changed("history-length") {
				req.HistoryLength = &historyLength
			}
			task, err := client.GetTask(ctx, req)
			if err != nil {
				return err
			}
			return printJSON(cmd.OutOrStdout(), task)
		},
	}
	cmd.Flags().IntVar(&historyLength, "history-length", 0, "Maximum number of history messages to retrieve")
	return cmd
}

func newTaskListCommand(opts *globalOptions) *cobra.Command {
	var (
		contextID        string
		status           string
		pageSize         int
		pageToken        string
		historyLength    int
		includeArtifacts bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks on the agent",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			client, _, err := dial(ctx, opts, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer client.Destroy()

			req := &a2a.ListTasksRequest{
				Tenant:           opts.tenant,
				ContextID:        contextID,
				Status:           a2a.TaskState(status),
				PageSize:         pageSize,
				PageToken:        pageToken,
				IncludeArtifacts: includeArtifacts,
			}
			if cmd.Flags().Changed("history-length") {
				req.HistoryLength = &historyLength
			}
			resp, err := client.ListTasks(ctx, req)
			if err != nil {
				return err
			}
			return printJSON(cmd.OutOrStdout(), resp)
		},
	}
	f := cmd.Flags()
	f.StringVar(&contextID, "context-id", "", "Filter by context ID")
	f.StringVar(&status, "status", "", "Filter by task state (e.g. submitted, working, completed)")
	f.IntVar(&pageSize, "page-size", 0, "Max tasks per page (1-100, server default if 0)")
	f.StringVar(&pageToken, "page-token", "", "Page token from a previous response")
	f.IntVar(&historyLength, "history-length", 0, "History messages to include per task")
	f.BoolVar(&includeArtifacts, "include-artifacts", false, "Include task artifacts in the response")
	return cmd
}

func newTaskCancelCommand(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel [task-id]",
		Short: "Cancel a task by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, _, err := dial(ctx, opts, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer client.Destroy()

			task, err := client.CancelTask(ctx, &a2a.CancelTaskRequest{Tenant: opts.tenant, ID: a2a.TaskID(args[0])})
			if err != nil {
				return err
			}
			return printJSON(cmd.OutOrStdout(), task)
		},
	}
}

func newTaskSubscribeCommand(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "subscribe [task-id]",
		Short: "Re-subscribe to an existing task and stream its events",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, _, err := dial(ctx, opts, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer client.Destroy()

			req := &a2a.SubscribeToTaskRequest{Tenant: opts.tenant, ID: a2a.TaskID(args[0])}
			out := cmd.OutOrStdout()
			for event, iterErr := range client.SubscribeToTask(ctx, req) {
				if iterErr != nil {
					return iterErr
				}
				if err := printJSON(out, event); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
