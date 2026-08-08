package tui

import (
	"time"

	"github.com/olafurb/krowser/internal/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// --- screen navigation messages ---

// setResourceMsg navigates to a resource list view.
type setResourceMsg struct{ res *k8s.Resource }

// setNamespaceMsg changes the active namespace ("" = all namespaces).
type setNamespaceMsg struct {
	ns  string
	all bool
}

// openDetailMsg opens the YAML/describe view for an object.
type openDetailMsg struct {
	res  *k8s.Resource
	obj  *unstructured.Unstructured
	kind string // "yaml" or "describe"
}

// openLogsMsg opens the log view for an object.
type openLogsMsg struct {
	res *k8s.Resource
	obj *unstructured.Unstructured
}

// openTopMsg opens the live usage view for a resource list.
type openTopMsg struct {
	res   *k8s.Resource
	ns    string
	allNs bool
}

// backMsg returns to the previous screen.
type backMsg struct{}

// helpToggleMsg toggles the help overlay.
type helpToggleMsg struct{}

// --- picker selection messages ---

type selectedContextMsg struct{ name string }

type selectedNamespaceMsg struct{ name string }

type selectedResourceMsg struct{ res *k8s.Resource }

type selectedContainerMsg struct{ name string }

type selectedThemeMsg struct{ name string }

// loadedNamespacesMsg carries the namespace list for the picker.
type loadedNamespacesMsg struct{ names []string }

// loadedKindsMsg carries the cluster's discovered resource kinds.
type loadedKindsMsg struct{ kinds []k8s.Kind }

// setScreenMsg forces navigation to an explicit screen.
type setScreenMsg struct{ screen screen }

// --- data loading messages ---

// loadedRowsMsg carries the result of a resource list.
type loadedRowsMsg struct {
	res  *k8s.Resource
	rows []k8s.Row
}

// loadErrorMsg carries a failed list/detail load.
type loadErrorMsg struct{ err error }

// loadedDetailMsg carries the YAML/describe payload for an object.
type loadedDetailMsg struct {
	kind string // "yaml" or "describe"
	data string
}

// --- background / streaming messages ---

// logLineMsg is emitted for each log line from a streaming pod log.
type logLineMsg struct{ line string }

// logStreamEndMsg is emitted when a log stream ends.
type logStreamEndMsg struct{ err error }

// loadedContainersMsg carries the container list of a pod.
type loadedContainersMsg struct{ containers []string }

// loadedTopMsg carries a fresh usage sample for the top screen.
type loadedTopMsg struct{ entries []k8s.TopEntry }

// topTickMsg drives the one-second refresh of the top screen.
type topTickMsg struct{ t time.Time }

// logReconnectMsg schedules a log stream reconnect while following.
type logReconnectMsg struct{}

// --- notification messages ---

// toastKind enumerates toast severities.
type toastKind int

// Toast severities.
const (
	ToastInfo toastKind = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// toastMsg shows a toast notification.
type toastMsg struct {
	kind  toastKind
	title string
	body  string
}

// toastDismissMsg hides the oldest visible toast.
type toastDismissMsg struct{}

// --- confirmation dialog messages ---

// dialogConfirmMsg confirms the pending dialog action.
type dialogConfirmMsg struct{}

// dialogCancelMsg cancels the pending dialog action.
type dialogCancelMsg struct{}

// openDialogMsg requests the root model to display a confirmation dialog.
type openDialogMsg struct{ req dialogRequest }

// actionResultMsg carries the result of a mutating action.
type actionResultMsg struct {
	success bool
	title   string
	message string
}

// --- status line messages ---

// statusMsg updates the transient status line.
type statusMsg struct{ text string }

// tickRefreshMsg drives periodic auto-refresh.
type tickRefreshMsg struct{ t time.Time }
