#!/bin/bash

# 🚀 SuperShell Setup Script - Agent OS Edition
# Automated installation and configuration for Unix-like systems

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
REPO_URL="https://github.com/your-repo/suppercommand"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.supershell"
VERSION="latest"

# Functions
print_header() {
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}🚀 SuperShell - Agent OS Edition Setup${NC}"
    echo -e "${CYAN}   Next-generation PowerShell/Bash replacement${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_step() {
    echo -e "${PURPLE}🔧 $1${NC}"
}

detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        ARCH="amd64"
        if [[ $(uname -m) == "aarch64" ]]; then
            ARCH="arm64"
        elif [[ $(uname -m) == "armv7l" ]]; then
            ARCH="arm"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
        ARCH="amd64"
        if [[ $(uname -m) == "arm64" ]]; then
            ARCH="arm64"
        fi
    elif [[ "$OSTYPE" == "freebsd"* ]]; then
        OS="freebsd"
        ARCH="amd64"
    else
        print_error "Unsupported operating system: $OSTYPE"
        exit 1
    fi
    
    print_info "Detected OS: $OS ($ARCH)"
}

check_dependencies() {
    print_step "Checking dependencies..."
    
    # Check for required tools
    MISSING_DEPS=()
    
    if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
        MISSING_DEPS+=("curl or wget")
    fi
    
    if ! command -v tar &> /dev/null; then
        MISSING_DEPS+=("tar")
    fi
    
    if [ ${#MISSING_DEPS[@]} -ne 0 ]; then
        print_error "Missing dependencies: ${MISSING_DEPS[*]}"
        print_info "Please install the missing dependencies and try again."
        exit 1
    fi
    
    print_success "All dependencies found"
}

check_permissions() {
    print_step "Checking permissions..."
    
    if [ ! -w "$INSTALL_DIR" ]; then
        print_warning "No write permission to $INSTALL_DIR"
        print_info "Will attempt to install with sudo"
        NEED_SUDO=1
    else
        NEED_SUDO=0
    fi
}

download_binary() {
    print_step "Downloading SuperShell binary..."
    
    BINARY_NAME="supershell-${OS}-${ARCH}"
    if [[ "$OS" == "windows" ]]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
    
    DOWNLOAD_URL="${REPO_URL}/releases/latest/download/${BINARY_NAME}.tar.gz"
    if [[ "$OS" == "windows" ]]; then
        DOWNLOAD_URL="${REPO_URL}/releases/latest/download/${BINARY_NAME%.exe}.zip"
    fi
    
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"
    
    print_info "Downloading from: $DOWNLOAD_URL"
    
    # Download with curl or wget
    if command -v curl &> /dev/null; then
        curl -L -o "supershell.tar.gz" "$DOWNLOAD_URL" || {
            print_error "Failed to download SuperShell"
            exit 1
        }
    else
        wget -O "supershell.tar.gz" "$DOWNLOAD_URL" || {
            print_error "Failed to download SuperShell"
            exit 1
        }
    fi
    
    # Extract
    if [[ "$OS" == "windows" ]]; then
        unzip "supershell.tar.gz"
    else
        tar -xzf "supershell.tar.gz"
    fi
    
    # Make executable
    chmod +x "$BINARY_NAME"
    
    print_success "Binary downloaded and extracted"
}

install_binary() {
    print_step "Installing SuperShell..."
    
    if [ $NEED_SUDO -eq 1 ]; then
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/supershell"
        sudo chmod +x "$INSTALL_DIR/supershell"
    else
        mv "$BINARY_NAME" "$INSTALL_DIR/supershell"
        chmod +x "$INSTALL_DIR/supershell"
    fi
    
    print_success "SuperShell installed to $INSTALL_DIR/supershell"
}

create_config() {
    print_step "Creating configuration directory..."
    
    mkdir -p "$CONFIG_DIR"
    
    # Create default config
    cat > "$CONFIG_DIR/config.yaml" << EOF
# SuperShell Configuration - Agent OS Edition

# General Settings
shell:
  prompt: "SuperShell> "
  history_size: 1000
  auto_complete: true
  color_output: true

# Agent OS Settings
agent:
  enabled: true
  hot_reload: true
  performance_monitoring: true
  plugin_auto_load: true

# Performance Settings
performance:
  command_timeout: 30s
  memory_limit: 100MB
  cache_size: 50MB

# Networking Settings
networking:
  timeout: 10s
  retries: 3
  user_agent: "SuperShell/1.0"

# Development Settings
development:
  debug_mode: false
  log_level: "info"
  profiling: false

# Plugin Settings
plugins:
  directories:
    - "$CONFIG_DIR/plugins"
    - "/usr/local/share/supershell/plugins"
  auto_update: false

# Security Settings
security:
  privilege_escalation: "prompt"
  command_logging: true
  safe_mode: false
EOF
    
    # Create plugins directory
    mkdir -p "$CONFIG_DIR/plugins"
    
    # Create aliases file
    cat > "$CONFIG_DIR/aliases.yaml" << EOF
# SuperShell Aliases
aliases:
  ll: "ls -la"
  la: "ls -la"
  l: "ls -l"
  ...: "cd ../.."
  ....: "cd ../../.."
  grep: "grep --color=auto"
  fgrep: "fgrep --color=auto"
  egrep: "egrep --color=auto"
  h: "history"
  c: "clear"
  q: "exit"
  nmap: "portscan"
  ss: "speedtest"
  sysmon: "perf monitor"
EOF
    
    print_success "Configuration created in $CONFIG_DIR"
}

setup_shell_integration() {
    print_step "Setting up shell integration..."
    
    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        case $SHELL in
            */bash)
                echo 'export PATH="'$INSTALL_DIR':$PATH"' >> ~/.bashrc
                print_info "Added to ~/.bashrc"
                ;;
            */zsh)
                echo 'export PATH="'$INSTALL_DIR':$PATH"' >> ~/.zshrc
                print_info "Added to ~/.zshrc"
                ;;
            */fish)
                echo 'set -gx PATH '$INSTALL_DIR' $PATH' >> ~/.config/fish/config.fish
                print_info "Added to ~/.config/fish/config.fish"
                ;;
            *)
                print_warning "Unknown shell: $SHELL"
                print_info "Please manually add $INSTALL_DIR to your PATH"
                ;;
        esac
    fi
    
    # Create desktop entry for GUI systems
    if [ -d "$HOME/.local/share/applications" ]; then
        cat > "$HOME/.local/share/applications/supershell.desktop" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=SuperShell
Comment=Next-generation PowerShell/Bash replacement
Exec=supershell
Icon=terminal
Terminal=true
Categories=System;TerminalEmulator;
EOF
        print_info "Desktop entry created"
    fi
}

verify_installation() {
    print_step "Verifying installation..."
    
    # Test if supershell is accessible
    if command -v supershell &> /dev/null; then
        VERSION_OUTPUT=$(supershell --version 2>/dev/null || echo "SuperShell installed")
        print_success "SuperShell is accessible: $VERSION_OUTPUT"
    else
        print_error "SuperShell not found in PATH"
        print_info "You may need to restart your shell or source your profile"
        return 1
    fi
    
    # Test basic functionality
    if supershell -c "help" &> /dev/null; then
        print_success "Basic functionality test passed"
    else
        print_warning "Basic functionality test failed"
    fi
    
    # Test Agent OS
    if supershell -c "dev profile" &> /dev/null; then
        print_success "Agent OS features are working"
    else
        print_warning "Agent OS features may not be working correctly"
    fi
}

show_completion_message() {
    echo ""
    echo -e "${GREEN}🎉 SuperShell installation completed successfully!${NC}"
    echo ""
    echo -e "${CYAN}📚 Quick Start:${NC}"
    echo -e "  ${YELLOW}supershell${NC}                 # Start SuperShell"
    echo -e "  ${YELLOW}supershell -c help${NC}         # Show help"
    echo -e "  ${YELLOW}supershell -c 'dev profile'${NC} # View performance stats"
    echo ""
    echo -e "${CYAN}🔥 Agent OS Features:${NC}"
    echo -e "  ${YELLOW}dev reload${NC}                 # Hot reload commands"
    echo -e "  ${YELLOW}perf stats${NC}                 # Performance monitoring"
    echo -e "  ${YELLOW}dev test <command>${NC}         # Interactive testing"
    echo -e "  ${YELLOW}dev docs${NC}                   # Generate documentation"
    echo ""
    echo -e "${CYAN}📖 Documentation:${NC}"
    echo -e "  Config: ${BLUE}$CONFIG_DIR/config.yaml${NC}"
    echo -e "  Aliases: ${BLUE}$CONFIG_DIR/aliases.yaml${NC}"
    echo -e "  Online: ${BLUE}$REPO_URL${NC}"
    echo ""
    echo -e "${CYAN}💡 Next Steps:${NC}"
    echo -e "  1. Restart your terminal or run: ${YELLOW}source ~/.bashrc${NC}"
    echo -e "  2. Run: ${YELLOW}supershell${NC}"
    echo -e "  3. Try: ${YELLOW}dev profile${NC} and ${YELLOW}perf stats${NC}"
    echo ""
    echo -e "${PURPLE}🚀 Happy networking and automation!${NC}"
}

cleanup() {
    if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
}

# Main installation process
main() {
    # Set trap for cleanup
    trap cleanup EXIT
    
    print_header
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version=*)
                VERSION="${1#*=}"
                shift
                ;;
            --install-dir=*)
                INSTALL_DIR="${1#*=}"
                shift
                ;;
            --config-dir=*)
                CONFIG_DIR="${1#*=}"
                shift
                ;;
            -h|--help)
                echo "SuperShell Setup Script"
                echo ""
                echo "Usage: $0 [options]"
                echo ""
                echo "Options:"
                echo "  --version=VERSION     Install specific version (default: latest)"
                echo "  --install-dir=DIR     Installation directory (default: /usr/local/bin)"
                echo "  --config-dir=DIR      Configuration directory (default: ~/.supershell)"
                echo "  -h, --help           Show this help message"
                echo ""
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    print_info "Installing SuperShell to: $INSTALL_DIR"
    print_info "Configuration directory: $CONFIG_DIR"
    echo ""
    
    # Installation steps
    detect_os
    check_dependencies
    check_permissions
    download_binary
    install_binary
    create_config
    setup_shell_integration
    verify_installation
    show_completion_message
    
    print_success "Installation completed!"
}

# Run main function
main "$@" 