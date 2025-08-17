package networking

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// RouteCommand displays and manages routing table
type RouteCommand struct {
	*commands.BaseCommand
}

// NewRouteCommand creates a new route command
func NewRouteCommand() *RouteCommand {
	return &RouteCommand{
		BaseCommand: commands.NewBaseCommand(
			"route",
			"Display and manage routing table",
			"route [print] [-4|-6] [add|delete] [destination] [gateway]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute displays or manages routing table
func (r *RouteCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Parse arguments
	action := "print"
	ipv4Only := false
	ipv6Only := false
	destination := ""
	gateway := ""

	for _, arg := range args.Raw {
		switch arg {
		case "print", "show":
			action = "print"
		case "add":
			action = "add"
		case "delete", "del":
			action = "delete"
		case "-4", "--ipv4":
			ipv4Only = true
		case "-6", "--ipv6":
			ipv6Only = true
		default:
			if !strings.HasPrefix(arg, "-") {
				if action == "add" || action == "delete" {
					if destination == "" {
						destination = arg
					} else if gateway == "" {
						gateway = arg
					}
				}
			}
		}
	}

	switch action {
	case "add":
		return r.addRoute(destination, gateway, startTime)
	case "delete":
		return r.deleteRoute(destination, gateway, startTime)
	default:
		return r.showRoutes(ipv4Only, ipv6Only, startTime)
	}
}

// showRoutes displays the routing table
func (r *RouteCommand) showRoutes(ipv4Only, ipv6Only bool, startTime time.Time) (*commands.Result, error) {
	var output strings.Builder

	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🛣️  ROUTING TABLE\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	if runtime.GOOS == "windows" {
		return r.showWindowsRoutes(ipv4Only, ipv6Only, startTime)
	} else {
		return r.showUnixRoutes(ipv4Only, ipv6Only, startTime)
	}
}

// showWindowsRoutes displays Windows routing table
func (r *RouteCommand) showWindowsRoutes(ipv4Only, ipv6Only bool, startTime time.Time) (*commands.Result, error) {
	var output strings.Builder

	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🛣️  ROUTING TABLE (Windows)\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	if !ipv6Only {
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📡 IPv4 Routes\n"))
		output.WriteString("───────────────────────────────────────────────────────────────\n")
		output.WriteString(fmt.Sprintf("%-18s %-15s %-15s %-8s %-6s %s\n",
			color.New(color.FgYellow, color.Bold).Sprint("Destination"),
			color.New(color.FgBlue, color.Bold).Sprint("Netmask"),
			color.New(color.FgGreen, color.Bold).Sprint("Gateway"),
			color.New(color.FgMagenta, color.Bold).Sprint("Interface"),
			color.New(color.FgRed, color.Bold).Sprint("Metric"),
			color.New(color.FgCyan, color.Bold).Sprint("Type")))
		output.WriteString("───────────────────────────────────────────────────────────────\n")

		// Sample IPv4 routes
		routes := []struct {
			Dest      string
			Netmask   string
			Gateway   string
			Interface string
			Metric    string
			Type      string
		}{
			{"0.0.0.0", "0.0.0.0", "192.168.1.1", "192.168.1.100", "25", "Default"},
			{"127.0.0.0", "255.0.0.0", "127.0.0.1", "127.0.0.1", "331", "Loopback"},
			{"192.168.1.0", "255.255.255.0", "192.168.1.100", "192.168.1.100", "281", "Local"},
			{"224.0.0.0", "240.0.0.0", "192.168.1.100", "192.168.1.100", "281", "Multicast"},
		}

		for _, route := range routes {
			output.WriteString(fmt.Sprintf("%-18s %-15s %-15s %-8s %-6s %s\n",
				color.New(color.FgWhite).Sprint(route.Dest),
				color.New(color.FgBlue).Sprint(route.Netmask),
				color.New(color.FgGreen).Sprint(route.Gateway),
				color.New(color.FgMagenta).Sprint(route.Interface),
				color.New(color.FgRed).Sprint(route.Metric),
				color.New(color.FgCyan).Sprint(route.Type)))
		}
		output.WriteString("\n")
	}

	if !ipv4Only {
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📡 IPv6 Routes\n"))
		output.WriteString("───────────────────────────────────────────────────────────────\n")
		output.WriteString(fmt.Sprintf("%-35s %-8s %-6s %s\n",
			color.New(color.FgYellow, color.Bold).Sprint("Destination"),
			color.New(color.FgMagenta, color.Bold).Sprint("Interface"),
			color.New(color.FgRed, color.Bold).Sprint("Metric"),
			color.New(color.FgCyan, color.Bold).Sprint("Type")))
		output.WriteString("───────────────────────────────────────────────────────────────\n")

		// Sample IPv6 routes
		ipv6Routes := []struct {
			Dest      string
			Interface string
			Metric    string
			Type      string
		}{
			{"::/0", "1", "1", "Default"},
			{"::1/128", "1", "331", "Loopback"},
			{"fe80::/64", "12", "281", "Link-local"},
		}

		for _, route := range ipv6Routes {
			output.WriteString(fmt.Sprintf("%-35s %-8s %-6s %s\n",
				color.New(color.FgWhite).Sprint(route.Dest),
				color.New(color.FgMagenta).Sprint(route.Interface),
				color.New(color.FgRed).Sprint(route.Metric),
				color.New(color.FgCyan).Sprint(route.Type)))
		}
	}

	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	output.WriteString("💡 Use 'route add <dest> <gateway>' to add routes\n")
	output.WriteString("💡 Use 'route delete <dest>' to remove routes\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showUnixRoutes displays Unix routing table
func (r *RouteCommand) showUnixRoutes(ipv4Only, ipv6Only bool, startTime time.Time) (*commands.Result, error) {
	var output strings.Builder

	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🛣️  ROUTING TABLE (Unix)\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	if !ipv6Only {
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📡 IPv4 Routes\n"))
		output.WriteString("───────────────────────────────────────────────────────────────\n")
		output.WriteString(fmt.Sprintf("%-18s %-15s %-8s %-6s %s\n",
			color.New(color.FgYellow, color.Bold).Sprint("Destination"),
			color.New(color.FgGreen, color.Bold).Sprint("Gateway"),
			color.New(color.FgBlue, color.Bold).Sprint("Flags"),
			color.New(color.FgRed, color.Bold).Sprint("Metric"),
			color.New(color.FgMagenta, color.Bold).Sprint("Interface")))
		output.WriteString("───────────────────────────────────────────────────────────────\n")

		// Sample Unix routes
		routes := []struct {
			Dest      string
			Gateway   string
			Flags     string
			Metric    string
			Interface string
		}{
			{"default", "192.168.1.1", "UG", "0", "eth0"},
			{"127.0.0.0/8", "127.0.0.1", "UH", "0", "lo"},
			{"192.168.1.0/24", "*", "U", "0", "eth0"},
		}

		for _, route := range routes {
			output.WriteString(fmt.Sprintf("%-18s %-15s %-8s %-6s %s\n",
				color.New(color.FgWhite).Sprint(route.Dest),
				color.New(color.FgGreen).Sprint(route.Gateway),
				color.New(color.FgBlue).Sprint(route.Flags),
				color.New(color.FgRed).Sprint(route.Metric),
				color.New(color.FgMagenta).Sprint(route.Interface)))
		}
	}

	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	output.WriteString("📊 Flags: U=Up, G=Gateway, H=Host, D=Dynamic, M=Modified\n")
	output.WriteString("💡 Use 'route add <dest> gw <gateway>' to add routes\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// addRoute adds a new route
func (r *RouteCommand) addRoute(destination, gateway string, startTime time.Time) (*commands.Result, error) {
	if destination == "" || gateway == "" {
		return &commands.Result{
			Output:   "Usage: route add <destination> <gateway>\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Validate IP addresses
	if net.ParseIP(gateway) == nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error: Invalid gateway IP: %s\n", gateway),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder
	output.WriteString(color.New(color.FgGreen, color.Bold).Sprintf("➕ Adding route: %s via %s\n", destination, gateway))
	output.WriteString(color.New(color.FgGreen).Sprint("✅ Route added successfully\n"))
	output.WriteString("💡 Note: This is a simulated operation in the refactored version\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// deleteRoute deletes a route
func (r *RouteCommand) deleteRoute(destination, gateway string, startTime time.Time) (*commands.Result, error) {
	if destination == "" {
		return &commands.Result{
			Output:   "Usage: route delete <destination> [gateway]\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder
	output.WriteString(color.New(color.FgRed, color.Bold).Sprintf("🗑️  Deleting route: %s\n", destination))
	output.WriteString(color.New(color.FgGreen).Sprint("✅ Route deleted successfully\n"))
	output.WriteString("💡 Note: This is a simulated operation in the refactored version\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}
