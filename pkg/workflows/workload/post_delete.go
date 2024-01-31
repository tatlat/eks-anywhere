package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type postDeleteWorkload struct{}

func (s *postDeleteWorkload) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.OriginalError != nil {
		collector := &workflows.CollectMgmtClusterDiagnosticsTask{}
		collector.Run(ctx, commandContext)
	}
	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster deleted!")
	}
	return nil
}

func (s *postDeleteWorkload) Name() string {
	return "validate-delete-workload-success"
}

func (s *postDeleteWorkload) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *postDeleteWorkload) Checkpoint() *task.CompletedTask {
	return nil
}
