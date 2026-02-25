package application

import "context"

type App struct {
	publicServer *PublicServer
	container    *Container
}

func (a *App) Run(ctx context.Context) error {

}

func (a *App) ShutDown(ctx context.Context) error {

}
