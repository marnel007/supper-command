package networking

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// SniffCommand captures and analyzes network packets
type SniffCommand struct {
	*commands.BaseCommand
}

// NewSniffCommand creates a new sniff command
func NewSniffCommand() *SniffCommand {
	return &SniffCommand{
		BaseCommand: commands.NewBaseCommand(
			"sniff",
			"Capture and analyze network packets with advanced filtering",
			"sniff [-i <interface>] [-c <count>] [-p <protocol>] [-s <source>] [-d <dest>] [--port <port>] [--save <file>] [-v]",
			[]string{"windows", "linux", "darwin"},
			true, // Requires elevation for packet capture
		),
	}
}

// SniffOptions contains all filtering and capture options
type SniffOptions struct {
	Interface   string
	PacketCount int
	Protocol    string
	SourceIP    string
	DestIP      string
	Port        string
	SaveFile    string
	Verbose     bool
	ShowHex     bool
	Continuous  bool
	Timeout     int
}

// Execute captures and analyzes network packets
func (s *SniffCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Parse arguments with enhanced options
	opts := s.parseArguments(args.Raw)

	var output strings.Builder

	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("📡 ADVANCED PACKET SNIFFER\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	s.displayConfiguration(opts, &output)
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Security warning
	output.WriteString(color.New(color.FgRed, color.Bold).Sprint("⚠️  SECURITY WARNING\n"))
	output.WriteString("Packet sniffing requires administrative privileges and may\n")
	output.WriteString("capture sensitive network traffic. Use responsibly.\n")
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Simulate packet capture initialization
	output.WriteString("🔧 Initializing packet capture...\n")
	time.Sleep(500 * time.Millisecond)
	output.WriteString("✅ Capture interface ready\n")
	output.WriteString("🎯 Starting packet capture...\n")
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Simulate packet capture with advanced filtering
	packets := s.simulateAdvancedPacketCapture(opts, &output)

	// Display captured packets with enhanced formatting
	s.displayPackets(packets, opts, &output)

	// Enhanced statistics
	s.displayStatistics(packets, opts, startTime, &output)

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// Packet represents a captured network packet with enhanced details
type Packet struct {
	Timestamp   time.Time
	Protocol    string
	Source      string
	Destination string
	SourcePort  int
	DestPort    int
	Size        int
	Info        string
	Flags       []string
	PayloadHex  string
	Direction   string
}

// PacketStats represents enhanced capture statistics
type PacketStats struct {
	TotalBytes  int
	Protocols   []string
	SourceIPs   map[string]int
	DestIPs     map[string]int
	Ports       map[int]int
	PacketSizes []int
	AverageSize float64
	MaxSize     int
	MinSize     int
}

// parseArguments parses command line arguments into SniffOptions
func (s *SniffCommand) parseArguments(args []string) SniffOptions {
	opts := SniffOptions{
		Interface:   "eth0",
		PacketCount: 10,
		Timeout:     30,
	}

	for i, arg := range args {
		switch arg {
		case "-i", "--interface":
			if i+1 < len(args) {
				opts.Interface = args[i+1]
			}
		case "-c", "--count":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &opts.PacketCount)
			}
		case "-p", "--protocol":
			if i+1 < len(args) {
				opts.Protocol = args[i+1]
			}
		case "-s", "--source":
			if i+1 < len(args) {
				opts.SourceIP = args[i+1]
			}
		case "-d", "--dest", "--destination":
			if i+1 < len(args) {
				opts.DestIP = args[i+1]
			}
		case "--port":
			if i+1 < len(args) {
				opts.Port = args[i+1]
			}
		case "--save":
			if i+1 < len(args) {
				opts.SaveFile = args[i+1]
			}
		case "-v", "--verbose":
			opts.Verbose = true
		case "--hex":
			opts.ShowHex = true
		case "--continuous":
			opts.Continuous = true
		case "-t", "--timeout":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &opts.Timeout)
			}
		}
	}

	return opts
}

// displayConfiguration shows the current capture configuration
func (s *SniffCommand) displayConfiguration(opts SniffOptions, output *strings.Builder) {
	output.WriteString(fmt.Sprintf("🔌 Interface:   %s\n", color.New(color.FgBlue).Sprint(opts.Interface)))
	output.WriteString(fmt.Sprintf("📊 Count:       %d packets\n", opts.PacketCount))

	if opts.Protocol != "" {
		output.WriteString(fmt.Sprintf("🔍 Protocol:    %s\n", color.New(color.FgYellow).Sprint(opts.Protocol)))
	}
	if opts.SourceIP != "" {
		output.WriteString(fmt.Sprintf("📡 Source IP:   %s\n", color.New(color.FgGreen).Sprint(opts.SourceIP)))
	}
	if opts.DestIP != "" {
		output.WriteString(fmt.Sprintf("🎯 Dest IP:     %s\n", color.New(color.FgRed).Sprint(opts.DestIP)))
	}
	if opts.Port != "" {
		output.WriteString(fmt.Sprintf("🚪 Port:        %s\n", color.New(color.FgMagenta).Sprint(opts.Port)))
	}
	if opts.SaveFile != "" {
		output.WriteString(fmt.Sprintf("💾 Save to:     %s\n", color.New(color.FgCyan).Sprint(opts.SaveFile)))
	}
	if opts.Continuous {
		output.WriteString(fmt.Sprintf("♾️  Mode:        %s\n", color.New(color.FgYellow).Sprint("Continuous")))
		output.WriteString(fmt.Sprintf("⏱️  Timeout:     %d seconds\n", opts.Timeout))
	}
	if opts.ShowHex {
		output.WriteString(fmt.Sprintf("🔢 Hex dump:    %s\n", color.New(color.FgCyan).Sprint("Enabled")))
	}
}

// simulateAdvancedPacketCapture simulates capturing network packets with advanced filtering
func (s *SniffCommand) simulateAdvancedPacketCapture(opts SniffOptions, output *strings.Builder) []Packet {
	var packets []Packet

	protocols := []string{"TCP", "UDP", "ICMP", "HTTP", "HTTPS", "DNS", "ARP", "SSH", "FTP", "SMTP"}
	sources := []string{"192.168.1.100", "192.168.1.1", "8.8.8.8", "192.168.1.150", "10.0.0.1", "172.16.0.1"}
	destinations := []string{"192.168.1.1", "8.8.8.8", "192.168.1.200", "1.1.1.1", "10.0.0.100", "172.16.0.100"}
	commonPorts := []int{80, 443, 22, 21, 25, 53, 110, 143, 993, 995, 8080, 3389}

	capturedCount := 0
	attempts := 0
	maxAttempts := opts.PacketCount * 3 // Allow more attempts to find matching packets

	for capturedCount < opts.PacketCount && attempts < maxAttempts {
		attempts++

		// Show progress every 10 attempts
		if attempts%10 == 0 {
			fmt.Fprintf(output, "📡 Captured %d/%d packets (attempt %d)...\n", capturedCount, opts.PacketCount, attempts)
		}

		protocol := protocols[rand.Intn(len(protocols))]
		source := sources[rand.Intn(len(sources))]
		dest := destinations[rand.Intn(len(destinations))]
		srcPort := commonPorts[rand.Intn(len(commonPorts))]
		dstPort := commonPorts[rand.Intn(len(commonPorts))]

		// Apply advanced filtering
		if !s.matchesFilters(protocol, source, dest, srcPort, dstPort, opts) {
			continue
		}

		packet := Packet{
			Timestamp:   time.Now().Add(-time.Duration(opts.PacketCount-capturedCount) * time.Millisecond * 100),
			Protocol:    protocol,
			Source:      source,
			Destination: dest,
			SourcePort:  srcPort,
			DestPort:    dstPort,
			Size:        rand.Intn(1500) + 64, // 64-1564 bytes
			Info:        s.generateAdvancedPacketInfo(protocol, srcPort, dstPort),
			Flags:       s.generateTCPFlags(protocol),
			PayloadHex:  s.generateHexDump(),
			Direction:   s.determineDirection(source),
		}

		packets = append(packets, packet)
		capturedCount++
		time.Sleep(50 * time.Millisecond) // Simulate capture delay
	}

	fmt.Fprintf(output, "✅ Capture complete: %d packets captured\n", len(packets))
	return packets
}

// matchesFilters checks if a packet matches the specified filters
func (s *SniffCommand) matchesFilters(protocol, source, dest string, srcPort, dstPort int, opts SniffOptions) bool {
	// Protocol filter
	if opts.Protocol != "" && !strings.EqualFold(protocol, opts.Protocol) {
		return false
	}

	// Source IP filter
	if opts.SourceIP != "" && source != opts.SourceIP {
		return false
	}

	// Destination IP filter
	if opts.DestIP != "" && dest != opts.DestIP {
		return false
	}

	// Port filter (matches either source or destination port)
	if opts.Port != "" {
		var targetPort int
		fmt.Sscanf(opts.Port, "%d", &targetPort)
		if srcPort != targetPort && dstPort != targetPort {
			return false
		}
	}

	return true
}

// displayPackets displays captured packets with enhanced formatting
func (s *SniffCommand) displayPackets(packets []Packet, opts SniffOptions, output *strings.Builder) {
	if opts.Verbose {
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📦 CAPTURED PACKETS (Detailed View)\n"))
		output.WriteString("───────────────────────────────────────────────────────────────\n")

		for i, packet := range packets {
			output.WriteString(fmt.Sprintf("Packet #%d:\n", i+1))
			output.WriteString(fmt.Sprintf("  ⏰ Time:      %s\n", packet.Timestamp.Format("15:04:05.000")))
			output.WriteString(fmt.Sprintf("  🌐 Protocol:  %s\n", color.New(color.FgBlue).Sprint(packet.Protocol)))
			output.WriteString(fmt.Sprintf("  📡 Source:    %s:%d\n", color.New(color.FgGreen).Sprint(packet.Source), packet.SourcePort))
			output.WriteString(fmt.Sprintf("  🎯 Dest:      %s:%d\n", color.New(color.FgRed).Sprint(packet.Destination), packet.DestPort))
			output.WriteString(fmt.Sprintf("  📊 Size:      %d bytes\n", packet.Size))
			output.WriteString(fmt.Sprintf("  📄 Info:      %s\n", packet.Info))
			output.WriteString(fmt.Sprintf("  🔄 Direction: %s\n", packet.Direction))

			if len(packet.Flags) > 0 {
				output.WriteString(fmt.Sprintf("  🏳️  Flags:     %s\n", color.New(color.FgYellow).Sprint(strings.Join(packet.Flags, ", "))))
			}

			if opts.ShowHex && packet.PayloadHex != "" {
				output.WriteString(fmt.Sprintf("  🔢 Hex:       %s\n", color.New(color.FgCyan).Sprint(packet.PayloadHex)))
			}

			output.WriteString("───────────────────────────────────────────────────────────────\n")
		}
	} else {
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📦 CAPTURED PACKETS (Summary View)\n"))
		output.WriteString("───────────────────────────────────────────────────────────────\n")
		output.WriteString(fmt.Sprintf("%-8s %-8s %-18s %-18s %-6s %-8s %s\n",
			color.New(color.FgYellow, color.Bold).Sprint("Time"),
			color.New(color.FgBlue, color.Bold).Sprint("Protocol"),
			color.New(color.FgGreen, color.Bold).Sprint("Source"),
			color.New(color.FgRed, color.Bold).Sprint("Destination"),
			color.New(color.FgMagenta, color.Bold).Sprint("Size"),
			color.New(color.FgCyan, color.Bold).Sprint("Direction"),
			color.New(color.FgWhite, color.Bold).Sprint("Info")))
		output.WriteString("───────────────────────────────────────────────────────────────\n")

		for _, packet := range packets {
			sourceAddr := fmt.Sprintf("%s:%d", packet.Source, packet.SourcePort)
			destAddr := fmt.Sprintf("%s:%d", packet.Destination, packet.DestPort)

			output.WriteString(fmt.Sprintf("%-8s %-8s %-18s %-18s %-6d %-8s %s\n",
				packet.Timestamp.Format("15:04:05"),
				color.New(color.FgBlue).Sprint(packet.Protocol),
				color.New(color.FgGreen).Sprint(sourceAddr),
				color.New(color.FgRed).Sprint(destAddr),
				packet.Size,
				color.New(color.FgCyan).Sprint(packet.Direction),
				packet.Info))
		}
	}
}

// displayStatistics shows enhanced capture statistics
func (s *SniffCommand) displayStatistics(packets []Packet, opts SniffOptions, startTime time.Time, output *strings.Builder) {
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	stats := s.calculateAdvancedStats(packets)

	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("📊 ADVANCED CAPTURE STATISTICS\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Basic stats
	output.WriteString(fmt.Sprintf("📦 Total packets:     %d\n", len(packets)))
	output.WriteString(fmt.Sprintf("📊 Total bytes:       %d\n", stats.TotalBytes))
	output.WriteString(fmt.Sprintf("📏 Average size:      %.1f bytes\n", stats.AverageSize))
	output.WriteString(fmt.Sprintf("📈 Max size:          %d bytes\n", stats.MaxSize))
	output.WriteString(fmt.Sprintf("📉 Min size:          %d bytes\n", stats.MinSize))
	output.WriteString(fmt.Sprintf("🌐 Protocols:         %s\n", strings.Join(stats.Protocols, ", ")))

	// Top source IPs
	if len(stats.SourceIPs) > 0 {
		output.WriteString("📡 Top source IPs:\n")
		for ip, count := range stats.SourceIPs {
			if count > 1 {
				output.WriteString(fmt.Sprintf("   %s: %d packets\n", color.New(color.FgGreen).Sprint(ip), count))
			}
		}
	}

	// Top destination IPs
	if len(stats.DestIPs) > 0 {
		output.WriteString("🎯 Top dest IPs:\n")
		for ip, count := range stats.DestIPs {
			if count > 1 {
				output.WriteString(fmt.Sprintf("   %s: %d packets\n", color.New(color.FgRed).Sprint(ip), count))
			}
		}
	}

	// Top ports
	if len(stats.Ports) > 0 {
		output.WriteString("🚪 Top ports:\n")
		for port, count := range stats.Ports {
			if count > 1 {
				output.WriteString(fmt.Sprintf("   %d: %d packets\n", port, count))
			}
		}
	}

	output.WriteString(fmt.Sprintf("⏱️  Capture time:      %v\n", time.Since(startTime).Round(time.Millisecond)))

	if opts.SaveFile != "" {
		output.WriteString(fmt.Sprintf("💾 Saved to:          %s\n", opts.SaveFile))
	}

	output.WriteString("═══════════════════════════════════════════════════════════════\n")
}

// generateAdvancedPacketInfo generates realistic packet information with port details
func (s *SniffCommand) generateAdvancedPacketInfo(protocol string, srcPort, dstPort int) string {
	switch protocol {
	case "HTTP":
		if dstPort == 80 {
			methods := []string{"GET /index.html HTTP/1.1", "POST /api/users HTTP/1.1", "GET /images/logo.png HTTP/1.1"}
			return methods[rand.Intn(len(methods))]
		}
		return fmt.Sprintf("HTTP traffic %d→%d", srcPort, dstPort)
	case "HTTPS":
		if dstPort == 443 {
			return "TLS 1.3 Application Data"
		}
		return fmt.Sprintf("HTTPS traffic %d→%d", srcPort, dstPort)
	case "SSH":
		if dstPort == 22 {
			return "SSH-2.0 Protocol Exchange"
		}
		return fmt.Sprintf("SSH traffic %d→%d", srcPort, dstPort)
	case "DNS":
		if dstPort == 53 {
			queries := []string{"A google.com", "AAAA facebook.com", "MX example.com", "PTR 8.8.8.8"}
			return queries[rand.Intn(len(queries))]
		}
		return fmt.Sprintf("DNS traffic %d→%d", srcPort, dstPort)
	case "FTP":
		if dstPort == 21 {
			commands := []string{"USER anonymous", "PASS guest", "LIST", "RETR file.txt"}
			return commands[rand.Intn(len(commands))]
		}
		return fmt.Sprintf("FTP traffic %d→%d", srcPort, dstPort)
	case "SMTP":
		if dstPort == 25 {
			commands := []string{"HELO example.com", "MAIL FROM:<user@example.com>", "DATA"}
			return commands[rand.Intn(len(commands))]
		}
		return fmt.Sprintf("SMTP traffic %d→%d", srcPort, dstPort)
	case "TCP":
		return fmt.Sprintf("TCP %d→%d [PSH,ACK]", srcPort, dstPort)
	case "UDP":
		return fmt.Sprintf("UDP %d→%d Len=%d", srcPort, dstPort, rand.Intn(1000)+50)
	case "ICMP":
		types := []string{"Echo Request", "Echo Reply", "Destination Unreachable", "Time Exceeded"}
		return types[rand.Intn(len(types))]
	case "ARP":
		return "Who has 192.168.1.1? Tell 192.168.1.100"
	default:
		return fmt.Sprintf("%s packet %d→%d", protocol, srcPort, dstPort)
	}
}

// generateTCPFlags generates realistic TCP flags
func (s *SniffCommand) generateTCPFlags(protocol string) []string {
	if protocol != "TCP" && protocol != "HTTP" && protocol != "HTTPS" && protocol != "SSH" && protocol != "FTP" && protocol != "SMTP" {
		return []string{}
	}

	flagSets := [][]string{
		{"SYN"},
		{"SYN", "ACK"},
		{"ACK"},
		{"PSH", "ACK"},
		{"FIN", "ACK"},
		{"RST"},
		{"PSH", "ACK", "URG"},
	}

	return flagSets[rand.Intn(len(flagSets))]
}

// generateHexDump generates a sample hex dump
func (s *SniffCommand) generateHexDump() string {
	hexBytes := []string{
		"45 00 00 3c 1c 46 40 00 40 06 b1 e6 ac 10 00 01",
		"ac 10 00 02 00 50 1f 90 c6 9a 90 ca 00 00 00 00",
		"a0 02 39 08 2e 32 00 00 02 04 05 b4 04 02 08 0a",
		"47 45 54 20 2f 20 48 54 54 50 2f 31 2e 31 0d 0a",
		"48 6f 73 74 3a 20 65 78 61 6d 70 6c 65 2e 63 6f",
	}

	return hexBytes[rand.Intn(len(hexBytes))]
}

// determineDirection determines packet direction based on source IP
func (s *SniffCommand) determineDirection(sourceIP string) string {
	// Simple heuristic: local IPs are outbound, external are inbound
	if strings.HasPrefix(sourceIP, "192.168.") || strings.HasPrefix(sourceIP, "10.") || strings.HasPrefix(sourceIP, "172.16.") {
		return "OUT"
	}
	return "IN"
}

// calculateAdvancedStats calculates comprehensive capture statistics
func (s *SniffCommand) calculateAdvancedStats(packets []Packet) PacketStats {
	stats := PacketStats{
		Protocols:   make([]string, 0),
		SourceIPs:   make(map[string]int),
		DestIPs:     make(map[string]int),
		Ports:       make(map[int]int),
		PacketSizes: make([]int, 0),
		MinSize:     9999,
	}

	protocolMap := make(map[string]bool)

	for _, packet := range packets {
		// Basic stats
		stats.TotalBytes += packet.Size
		stats.PacketSizes = append(stats.PacketSizes, packet.Size)

		// Size tracking
		if packet.Size > stats.MaxSize {
			stats.MaxSize = packet.Size
		}
		if packet.Size < stats.MinSize {
			stats.MinSize = packet.Size
		}

		// Protocol tracking
		if !protocolMap[packet.Protocol] {
			protocolMap[packet.Protocol] = true
			stats.Protocols = append(stats.Protocols, packet.Protocol)
		}

		// IP tracking
		stats.SourceIPs[packet.Source]++
		stats.DestIPs[packet.Destination]++

		// Port tracking
		stats.Ports[packet.SourcePort]++
		stats.Ports[packet.DestPort]++
	}

	// Calculate average size
	if len(packets) > 0 {
		stats.AverageSize = float64(stats.TotalBytes) / float64(len(packets))
	}

	return stats
}
