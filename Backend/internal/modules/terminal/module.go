package terminal

import (
	"context"

	"ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/modules/terminal/infrastructure/docker"
	terminalgithub "ai-agent/internal/modules/terminal/infrastructure/github"
	terminalpostgres "ai-agent/internal/modules/terminal/infrastructure/postgres"
	"ai-agent/internal/modules/terminal/policy"
	"ai-agent/internal/modules/terminal/worker"
	"ai-agent/internal/shared/logger"

	"github.com/jmoiron/sqlx"
)

type ModuleConfig struct {
	Database     *sqlx.DB
	Logger       logger.Logger
	DockerBinary string
	PolicyPath   string
	ProjectRepo  worker.ProjectRepository
}

type Module struct {
	Service *application.Service
	Worker  *worker.SandboxWorker
	Handler *Handler
}

func LoadModule(config ModuleConfig) (*Module, error) {
	sandboxPolicy, err := policy.Load(config.PolicyPath)
	if err != nil {
		return nil, err
	}

	dockerCLI := docker.NewDockerCLI(config.DockerBinary)
	runtime := docker.NewDockerRuntimeWithCLI(dockerCLI)
	files := docker.NewDockerFileStore(dockerCLI)
	repositoryCloner := terminalgithub.NewGitHubRepositoryCloner(dockerCLI)
	repositoryPusher := terminalgithub.NewGitHubRepositoryPusher(dockerCLI)
	repo := terminalpostgres.NewRepository(config.Database)
	service := application.NewService(repo, runtime, files, repositoryCloner, sandboxPolicy, repositoryPusher)
	sandboxWorker := worker.NewSandboxWorker(service, config.ProjectRepo, config.Logger)

	module := &Module{
		Service: service,
		Worker:  sandboxWorker,
	}
	module.Handler = NewHandler(module)

	return module, nil
}

func (m *Module) Start(ctx context.Context, workerCount int) {
	if m.Worker != nil {
		m.Worker.Start(ctx, workerCount)
	}
}

func (m *Module) Stop() {
	if m.Worker != nil {
		m.Worker.Stop()
	}
}

func (m *Module) OnProjectCreated(ctx context.Context, projectID, userID string) {
	if m.Worker != nil {
		m.Worker.Publish(worker.ProjectEvent{
			Type:      "project.created",
			ProjectID: projectID,
			UserID:    userID,
		})
	}
}

func (m *Module) OnProjectDeleted(ctx context.Context, projectID string) error {
	if m.Worker == nil {
		return nil
	}
	return m.Worker.DeleteProject(ctx, projectID)
}

func (m *Module) OnProjectBranchUpdated(ctx context.Context, projectID string) error {
	if m.Worker == nil {
		return nil
	}
	return m.Worker.RefreshProjectBranch(ctx, projectID)
}
