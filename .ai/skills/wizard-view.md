---
name: wizard-view
description: Create a custom View for the wizard engine
---

# Creating a Wizard View

A View is an actor that receives messages and renders output in the wizard layout.

## When to Use

- Display dynamic information alongside wizard prompts (cost, quota, status)
- React to step completion or store changes
- Communicate data between display components

## Template

```go
package mypackage

import (
    "fmt"
    "reflect"

    "github.com/verda-cloud/verdagostack/pkg/tui/wizard"
)

// MyDataMsg is published by this view for other views to consume.
// Message types are defined by the producer.
type MyDataMsg struct {
    Value string
}

// MyView displays [describe what it shows].
type MyView struct {
    last string
}

func (v *MyView) Update(msg any) (string, []any) {
    switch m := msg.(type) {
    case wizard.CollectedChangedMsg:
        // React to step completion.
        val := m.Collected["step-name"]
        v.last = fmt.Sprintf("  Info: %v\n", val)
        // Optionally publish for other views.
        return v.last, []any{MyDataMsg{Value: val.(string)}}
    case wizard.StoreChangedMsg:
        // React to store.Set() calls from loaders.
    }
    return v.last, nil
}

// Subscribe returns nil for engine broadcasts only.
// Add specific types to receive inter-view messages:
//   return []reflect.Type{reflect.TypeFor[OtherViewMsg]()}
func (v *MyView) Subscribe() []reflect.Type {
    return nil
}
```

## Register in Flow Layout

```go
flow := &wizard.Flow{
    Layout: []wizard.ViewDef{
        {ID: "progress", View: wizard.NewProgressView()},
        {ID: "myview", View: &MyView{}},
    },
    Steps: []wizard.Step{...},
}
```

## Key Rules

1. **Return unchanged output** when the message isn't relevant — the engine skips printing unchanged views
2. **Message types are owned by producers** — define your msg struct where you publish it
3. **Subscribe only for inter-view messages** — engine broadcasts (StepChanged, CollectedChanged, StoreChanged) always arrive regardless of Subscribe
4. **No queues** — Update is called synchronously, process and return immediately
