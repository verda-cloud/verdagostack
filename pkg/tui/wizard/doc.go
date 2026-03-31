// Package wizard provides a TUI wizard engine for interactive CLI flows.
//
// A wizard Flow is an ordered list of Steps. Each Step defines a prompt type,
// an optional data loader (for fetching choices from APIs), conditional skip
// logic, validation, and dependency tracking.
//
// The Engine executes the flow step by step, skipping steps already provided
// via CLI flags, caching loader results, and supporting back-navigation when
// a required step has no available options.
//
// Example usage:
//
//	flow := &wizard.Flow{
//	    Name: "vm-create",
//	    Steps: []wizard.Step{
//	        {
//	            Name:   "region",
//	            Prompt: wizard.SelectPrompt,
//	            Loader: wizard.StaticChoices(
//	                wizard.Choice{Label: "Finland", Value: "FIN-01"},
//	            ),
//	            Setter: func(v any) { opts.Region = v.(string) },
//	        },
//	    },
//	}
//
//	engine := wizard.NewEngine(prompter)
//	if err := engine.Run(ctx, flow); err != nil {
//	    return err
//	}
package wizard
