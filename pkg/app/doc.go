// Package app provides a framework for building CLI applications with
// Cobra commands, Viper configuration, and pflag flag management.
//
// A typical application is created with NewApp and run with Run:
//
//	app := app.NewApp("myapp", "My application",
//	    app.WithOptions(&myOptions{}),
//	    app.WithRunFunc(func(ctx context.Context) error { return startServer(ctx) }),
//	)
//	app.Run()
package app
