---
name: wizard-step
description: Add a new step to a wizard flow
---

# Adding a Wizard Step

## Template

```go
wizard.Step{
    Name:        "field-name",              // unique key in collected map
    Description: "User-facing prompt text", // shown to user
    Prompt:      wizard.SelectPrompt,       // SelectPrompt | TextInputPrompt | ConfirmPrompt | PasswordPrompt | MultiSelectPrompt
    Required:    true,
    DependsOn:   []string{"prior-step"},    // invalidate cache when these change

    // Load choices from API (select/multi-select only).
    Loader: func(ctx context.Context, p tui.Prompter, s tui.Status, store *wizard.Store) ([]wizard.Choice, error) {
        c := store.Collected()
        region := c["region"].(string)

        sp, _ := s.Spinner(ctx, "Loading...")
        items, err := api.List(ctx, region)
        sp.Stop("Done")

        var choices []wizard.Choice
        for _, item := range items {
            choices = append(choices, wizard.Choice{
                Label: item.Name,
                Value: item.ID,
            })
        }
        return choices, err
    },

    // For fixed choices, use StaticChoices helper:
    // Loader: wizard.StaticChoices(
    //     wizard.Choice{Label: "Small", Value: "small"},
    //     wizard.Choice{Label: "Large", Value: "large"},
    // ),

    // Default value based on previous answers.
    Default: func(c map[string]any) any {
        if c["environment"] == "prod" { return "3" }
        return "1"
    },

    // Validate before accepting.
    Validate: func(v any) error {
        if v.(string) == "" { return fmt.Errorf("cannot be empty") }
        return nil
    },

    // Write to your options struct.
    Setter:   func(v any) { opts.Field = v.(string) },
    Resetter: func()      { opts.Field = "" },

    // Skip if already set via CLI flag.
    IsSet: func() bool { return opts.Field != "" },
    Value: func() any  { return opts.Field },

    // Conditional skip based on earlier answers.
    ShouldSkip: func(c map[string]any) bool {
        return c["environment"] == "dev"
    },
}
```

## Checklist

- [ ] `Name` is unique across all steps in the flow
- [ ] `DependsOn` lists steps whose changes should invalidate this step's cached choices
- [ ] `Setter` and `Resetter` are paired — Resetter clears what Setter writes
- [ ] `IsSet` + `Value` are paired — skip the prompt when value comes from CLI flags
- [ ] Loader uses `store.Collected()` (not a captured map) for current values
- [ ] Loader shows a spinner via `tui.Status` for API calls
