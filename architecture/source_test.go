package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestFeatureHasNoPrivateProductOrExecutionOwner(t *testing.T) {
	root := ".."
	const feature = "github.com/MalenkiySolovey/solovey-component-server-protection"
	const host = "github.com/MalenkiySolovey/solovey-host"
	if err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == "node_modules" || strings.HasPrefix(entry.Name(), ".") && name != root) {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			return err
		}
		aliases := map[string]string{}
		for _, item := range file.Imports {
			imported, err := strconv.Unquote(item.Path.Value)
			if err != nil {
				return err
			}
			if strings.Contains(imported, ".") && imported != feature && !strings.HasPrefix(imported, feature+"/") && imported != host && !strings.HasPrefix(imported, host+"/") && imported != "golang.org/x/sys/windows" {
				t.Errorf("%s imports unapproved module %s", relative, imported)
			}
			if imported == "os/exec" || imported == "unsafe" || strings.HasPrefix(imported, "database/") || strings.HasPrefix(imported, "net/http") {
				t.Errorf("%s imports execution/storage owner %s", relative, imported)
			}
			if imported == "io/fs" && relative != "service/fronting/reconcile_v2.go" {
				t.Errorf("%s imports filesystem authority", relative)
			}
			if imported == "os" && relative != "service/operations/manager.go" && relative != "service/operations/pid_unix.go" {
				t.Errorf("%s imports filesystem authority", relative)
			}
			if imported == "syscall" && relative != "service/operations/pid_unix.go" {
				t.Errorf("%s imports syscall authority", relative)
			}
			if imported == "golang.org/x/sys/windows" && relative != "service/operations/pid_windows.go" {
				t.Errorf("%s imports platform authority", relative)
			}
			alias := filepath.Base(imported)
			if item.Name != nil {
				alias = item.Name.Name
			}
			aliases[alias] = imported
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Signal" {
					identifier, imported := selector.X.(*ast.Ident)
					if imported && aliases[identifier.Name] == "syscall" {
						zero := len(call.Args) == 1
						if zero {
							literal, ok := call.Args[0].(*ast.BasicLit)
							zero = ok && literal.Kind == token.INT && literal.Value == "0"
						}
						if relative != "service/operations/pid_unix.go" || !zero {
							t.Errorf("%s attempts a mutating process signal", relative)
						}
					} else if relative != "service/operations/pid_unix.go" {
						t.Errorf("%s attempts process signaling", relative)
					}
				}
			}
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			identifier, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			imported := aliases[identifier.Name]
			allowed := true
			switch imported {
			case "io/fs":
				allowed = selector.Sel.Name == "ErrNotExist"
			case "os":
				allowed = selector.Sel.Name == "Getpid" && relative == "service/operations/manager.go" || (selector.Sel.Name == "FindProcess" || selector.Sel.Name == "ErrProcessDone") && relative == "service/operations/pid_unix.go"
			case "syscall":
				allowed = selector.Sel.Name == "Signal" || selector.Sel.Name == "EPERM" || selector.Sel.Name == "ESRCH"
			case "golang.org/x/sys/windows":
				allowed = strings.Contains("|OpenProcess|PROCESS_QUERY_LIMITED_INFORMATION|ERROR_INVALID_PARAMETER|CloseHandle|GetExitCodeProcess|", "|"+selector.Sel.Name+"|")
			case "net":
				allowed = !strings.HasPrefix(selector.Sel.Name, "Listen") && !strings.HasPrefix(selector.Sel.Name, "Dial") || selector.Sel.Name == "Dialer" && relative == "service/udpguard/plain_udp_probe.go"
			}
			if !allowed {
				t.Errorf("%s uses unowned operation %s.%s", relative, imported, selector.Sel.Name)
			}
			return true
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
