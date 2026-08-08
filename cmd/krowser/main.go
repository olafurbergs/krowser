package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/olafurb/krowser/internal/k8s"
	"github.com/olafurb/krowser/internal/tui"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig (default: $KUBECONFIG or ~/.kube/config)")
	context := flag.String("context", "", "kubeconfig context to use")
	namespace := flag.String("namespace", "", "namespace to browse (default: context namespace, or all namespaces)")
	allNamespaces := flag.Bool("all-namespaces", false, "browse all namespaces")
	theme := flag.String("theme", "", "theme name (default: Monokai)")
	flag.Parse()

	kc := *kubeconfig
	if kc == "" {
		kc = k8s.DefaultKubeconfigPath()
	}

	contexts, err := k8s.Contexts(kc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "krowser: loading kubeconfig: %v\n", err)
		os.Exit(1)
	}
	if len(contexts) == 0 {
		fmt.Fprintln(os.Stderr, "krowser: no contexts found in kubeconfig")
		os.Exit(1)
	}

	ctxName := *context
	if ctxName == "" {
		for _, c := range contexts {
			if c.Current {
				ctxName = c.Name
				break
			}
		}
		if ctxName == "" {
			ctxName = contexts[0].Name
		}
	}

	client, err := k8s.New(kc, ctxName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "krowser: %v\n", err)
		os.Exit(1)
	}

	nsName := *namespace
	allNs := *allNamespaces
	if nsName == "" {
		nsName = client.Namespace()
		if nsName == "" {
			// No namespace on the context either: browse everything rather
			// than an often-empty default namespace.
			allNs = true
		}
	}

	cfg := tui.Config{
		Client:        client,
		Contexts:      contexts,
		Kubeconfig:    kc,
		Namespace:     nsName,
		AllNamespaces: allNs,
		ContextPicker: *context == "" && len(contexts) > 1,
		Theme:         *theme,
	}

	p := tea.NewProgram(tui.New(cfg))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
