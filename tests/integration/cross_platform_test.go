package integration

import (
	"context"
	"runtime"
	"testing"

	"suppercommand/internal/managers/firewall"
	"suppercommand/internal/managers/performance"
	"suppercommand/internal/managers/server"
	"suppercommand/internal/types"
)

func TestCrossPlatformFirewall(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create firewall manager: %v", err)
	}

	ctx := context.Background()

	// Test status retrieval across platforms
	status, err := manager.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get firewall status: %v", err)
	}

	// Validate platform-specific behavior
	switch runtime.GOOS {
	case "windows":
		if status.Platform != types.PlatformWindows {
			t.Errorf("Expected Windows platform, got %v", status.Platform)
		}
	case "linux":
		if status.Platform != types.PlatformLinux {
			t.Errorf("Expected Linux platform, got %v", status.Platform)
		}
	case "darwin":
		if status.Platform != types.PlatformDarwin {
			t.Errorf("Expected Darwin platform, got %v", status.Platform)
		}
	}

	// Test rule listing
	rules, err := manager.ListRules(ctx)
	if err != nil {
		t.Fatalf("Failed to list firewall rules: %v", err)
	}

	if rules == nil {
		t.Error("Rules should not be nil")
	}

	t.Logf("Platform: %s, Enabled: %t, Rules: %d",
		status.Platform, status.Enabled, len(rules))
}

func TestCrossPlatformPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	factory := performance.NewFactory()
	analyzer, err := factory.CreateAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create performance analyzer: %v", err)
	}

	ctx := context.Background()

	// Test metrics collection across platforms
	metrics, err := analyzer.CollectMetrics(ctx)
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Validate metrics structure
	if metrics.CPU == nil {
		t.Error("CPU metrics should not be nil")
	}

	if metrics.Memory == nil {
		t.Error("Memory metrics should not be nil")
	}

	if metrics.Disk == nil {
		t.Error("Disk metrics should not be nil")
	}

	// Platform-specific validations
	switch runtime.GOOS {
	case "windows":
		// Windows-specific checks
		if metrics.CPU.Cores <= 0 {
			t.Error("CPU cores should be positive on Windows")
		}
	case "linux":
		// Linux-specific checks
		if len(metrics.CPU.LoadAverage) != 3 {
			t.Error("Linux should provide 3 load average values")
		}
	case "darwin":
		// macOS-specific checks
		if metrics.Memory.Total <= 0 {
			t.Error("Memory total should be positive on macOS")
		}
	}

	// Test analysis
	analysis, err := analyzer.AnalyzeMetrics(ctx, metrics)
	if err != nil {
		t.Fatalf("Failed to analyze metrics: %v", err)
	}

	if analysis.OverallScore < 0 || analysis.OverallScore > 100 {
		t.Errorf("Overall score should be between 0-100, got %f", analysis.OverallScore)
	}

	t.Logf("Platform: %s, CPU: %.1f%%, Memory: %.1f%%, Score: %.1f",
		runtime.GOOS, metrics.CPU.Usage, metrics.Memory.Usage, analysis.OverallScore)
}

func TestCrossPlatformServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create server manager: %v", err)
	}

	ctx := context.Background()

	// Test health status across platforms
	health, err := manager.GetHealthStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get health status: %v", err)
	}

	// Validate health components
	expectedComponents := []string{"CPU", "Memory", "Disk", "Network"}
	for _, component := range expectedComponents {
		if _, exists := health.Components[component]; !exists {
			t.Errorf("Component %s should exist", component)
		}
	}

	// Test service listing
	services, err := manager.ListServices(ctx)
	if err != nil {
		t.Fatalf("Failed to list services: %v", err)
	}

	if services == nil {
		t.Error("Services should not be nil")
	}

	// Platform-specific service checks
	switch runtime.GOOS {
	case "windows":
		// Windows should have some system services
		if len(services) == 0 {
			t.Log("Warning: No services found on Windows (might be expected in test environment)")
		}
	case "linux":
		// Linux should have systemd or init services
		if len(services) == 0 {
			t.Log("Warning: No services found on Linux (might be expected in test environment)")
		}
	}

	// Test user sessions
	users, err := manager.GetActiveUsers(ctx)
	if err != nil {
		t.Fatalf("Failed to get active users: %v", err)
	}

	if users == nil {
		t.Error("Users should not be nil")
	}

	t.Logf("Platform: %s, Health: %s, Services: %d, Users: %d",
		runtime.GOOS, health.Overall, len(services), len(users))
}

func TestPlatformSpecificFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	switch runtime.GOOS {
	case "windows":
		t.Run("Windows-specific", testWindowsSpecific)
	case "linux":
		t.Run("Linux-specific", testLinuxSpecific)
	case "darwin":
		t.Run("macOS-specific", testMacOSSpecific)
	default:
		t.Skipf("Platform %s not supported for specific tests", runtime.GOOS)
	}
}

func testWindowsSpecific(t *testing.T) {
	// Test Windows-specific firewall features
	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create firewall manager: %v", err)
	}

	ctx := context.Background()
	status, err := manager.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get firewall status: %v", err)
	}

	if status.Platform != types.PlatformWindows {
		t.Errorf("Expected Windows platform, got %v", status.Platform)
	}

	// Test Windows firewall profiles
	if status.Profile == "" {
		t.Error("Windows firewall should have a profile")
	}

	t.Logf("Windows firewall profile: %s", status.Profile)
}

func testLinuxSpecific(t *testing.T) {
	// Test Linux-specific features
	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create firewall manager: %v", err)
	}

	ctx := context.Background()
	status, err := manager.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get firewall status: %v", err)
	}

	if status.Platform != types.PlatformLinux {
		t.Errorf("Expected Linux platform, got %v", status.Platform)
	}

	// Test Linux firewall type (ufw or iptables)
	if status.Profile == "" {
		t.Error("Linux firewall should indicate type (UFW or iptables)")
	}

	t.Logf("Linux firewall type: %s", status.Profile)
}

func testMacOSSpecific(t *testing.T) {
	// Test macOS-specific features
	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create firewall manager: %v", err)
	}

	ctx := context.Background()
	status, err := manager.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get firewall status: %v", err)
	}

	if status.Platform != types.PlatformDarwin {
		t.Errorf("Expected Darwin platform, got %v", status.Platform)
	}

	// Test macOS firewall (pfctl)
	if status.Profile == "" {
		t.Error("macOS firewall should have profile information")
	}

	t.Logf("macOS firewall profile: %s", status.Profile)
}

func TestEndToEndWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Test complete workflow: Performance -> Server -> Firewall
	t.Run("Performance Analysis", func(t *testing.T) {
		factory := performance.NewFactory()
		analyzer, err := factory.CreateAnalyzer()
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}

		// Collect and analyze metrics
		metrics, err := analyzer.CollectMetrics(ctx)
		if err != nil {
			t.Fatalf("Failed to collect metrics: %v", err)
		}

		analysis, err := analyzer.AnalyzeMetrics(ctx, metrics)
		if err != nil {
			t.Fatalf("Failed to analyze metrics: %v", err)
		}

		t.Logf("Performance score: %.1f", analysis.OverallScore)
	})

	t.Run("Server Health Check", func(t *testing.T) {
		factory := server.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		health, err := manager.GetHealthStatus(ctx)
		if err != nil {
			t.Fatalf("Failed to get health: %v", err)
		}

		t.Logf("Server health: %s", health.Overall)
	})

	t.Run("Firewall Status", func(t *testing.T) {
		factory := firewall.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		status, err := manager.GetStatus(ctx)
		if err != nil {
			t.Fatalf("Failed to get status: %v", err)
		}

		t.Logf("Firewall enabled: %t, Rules: %d", status.Enabled, status.RuleCount)
	})
}

func TestResourceUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test that our components don't use excessive resources
	ctx := context.Background()

	// Measure memory usage during operations
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Perform operations
	factories := []interface{}{
		firewall.NewFactory(),
		performance.NewFactory(),
		server.NewFactory(),
	}

	for _, factory := range factories {
		switch f := factory.(type) {
		case *firewall.Factory:
			if manager, err := f.CreateManager(); err == nil {
				manager.GetStatus(ctx)
				manager.ListRules(ctx)
			}
		case *performance.Factory:
			if analyzer, err := f.CreateAnalyzer(); err == nil {
				if metrics, err := analyzer.CollectMetrics(ctx); err == nil {
					analyzer.AnalyzeMetrics(ctx, metrics)
				}
			}
		case *server.Factory:
			if manager, err := f.CreateManager(); err == nil {
				manager.GetHealthStatus(ctx)
				manager.ListServices(ctx)
			}
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	memUsed := m2.Alloc - m1.Alloc
	t.Logf("Memory used during operations: %d bytes", memUsed)

	// Reasonable memory usage threshold (adjust as needed)
	if memUsed > 50*1024*1024 { // 50MB
		t.Errorf("Excessive memory usage: %d bytes", memUsed)
	}
}

func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	numGoroutines := 10

	// Test concurrent firewall operations
	t.Run("Concurrent Firewall", func(t *testing.T) {
		factory := firewall.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		done := make(chan bool, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- true }()
				_, err := manager.GetStatus(ctx)
				if err != nil {
					t.Errorf("Concurrent firewall operation failed: %v", err)
				}
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	// Test concurrent performance operations
	t.Run("Concurrent Performance", func(t *testing.T) {
		factory := performance.NewFactory()
		analyzer, err := factory.CreateAnalyzer()
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}

		done := make(chan bool, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- true }()
				_, err := analyzer.CollectMetrics(ctx)
				if err != nil {
					t.Errorf("Concurrent performance operation failed: %v", err)
				}
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}
