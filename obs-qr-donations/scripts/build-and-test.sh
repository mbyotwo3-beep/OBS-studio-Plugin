#!/usr/bin/env bash
# 
# Build and Test Script for OBS QR Donations Plugin (Linux/macOS)
#
# Usage:
#   ./scripts/build-and-test.sh [--full-sdk] [--skip-tests]
#

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Parse arguments
FULL_SDK=false
SKIP_TESTS=false

for arg in "$@"; do
    case $arg in
        --full-sdk)
            FULL_SDK=true
            shift
            ;;
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --help)
            echo "Usage: $0 [--full-sdk] [--skip-tests]"
            echo "  --full-sdk    Build with full Breez SDK instead of stub"
            echo "  --skip-tests  Skip running tests after build"
            exit 0
            ;;
    esac
done

echo -e "${CYAN}🚀 OBS QR Donations - Build & Test Script${NC}"
echo "========================================"
echo ""

# Step 1: Check prerequisites
echo -e "${YELLOW}📋 Step 1: Checking prerequisites...${NC}"

check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "  ${GREEN}✅ $1 found${NC}"
        return 0
    else
        echo -e "  ${RED}❌ $1 not found${NC}"
        return 1
    fi
}

MISSING_PREREQS=false
check_command cmake || MISSING_PREREQS=true
check_command git || MISSING_PREREQS=true
check_command qmake || MISSING_PREREQS=true

if [ "$MISSING_PREREQS" = true ]; then
    echo -e "\n${RED}❌ Missing prerequisites. Please install required tools.${NC}"
    echo "See BUILD_GUIDE.md for installation instructions."
    exit 1
fi

# Step 2: Configure CMake
echo -e "\n${YELLOW}📦 Step 2: Configuring CMake build...${NC}"

CMAKE_ARGS="-B build -S . -DBREEZ_USE_STUB=ON -DBREEZ_STUB_SIMULATE=ON"

if [ "$FULL_SDK" = true ]; then
    echo -e "  ${CYAN}Building with full Breez SDK...${NC}"
    CMAKE_ARGS="-B build -S ."
fi

if cmake $CMAKE_ARGS; then
    echo -e "  ${GREEN}✅ CMake configuration successful${NC}"
else
    echo -e "  ${RED}❌ CMake configuration failed${NC}"
    echo -e "\n${YELLOW}💡 Troubleshooting tips:${NC}"
    echo "  - Make sure OBS Studio development files are installed"
    echo "  - Check that Qt6 is properly installed"
    echo "  - Try: sudo apt install obs-studio-dev qt6-base-dev"
    exit 1
fi

# Step 3: Build the plugin
echo -e "\n${YELLOW}🔨 Step 3: Building plugin...${NC}"

if cmake --build build --config Release; then
    echo -e "  ${GREEN}✅ Build successful${NC}"
else
    echo -e "  ${RED}❌ Build failed${NC}"
    exit 1
fi

# Step 4: Verify build artifacts
echo -e "\n${YELLOW}🔍 Step 4: Verifying build artifacts...${NC}"

PLUGIN_FILE="build/libobs-qr-donations.so"
if [ -f "$PLUGIN_FILE" ]; then
    SIZE=$(du -h "$PLUGIN_FILE" | cut -f1)
    echo -e "  ${GREEN}✅ Plugin binary found: $PLUGIN_FILE ($SIZE)${NC}"
else
    echo -e "  ${RED}❌ Plugin binary not found${NC}"
    exit 1
fi

# Step 5: Run tests
if [ "$SKIP_TESTS" = false ]; then
    echo -e "\n${YELLOW}🧪 Step 5: Running integration tests...${NC}"
    
    if command -v python3 &> /dev/null; then
        python3 scripts/run_integration_test.py || echo -e "  ${YELLOW}⚠️  Some tests failed${NC}"
    else
        echo -e "  ${YELLOW}⚠️  Python3 not found, skipping tests${NC}"
    fi
else
    echo -e "\n${YELLOW}⏭️  Step 5: Skipping tests (--skip-tests specified)${NC}"
fi

# Step 6: Installation instructions
echo -e "\n${YELLOW}📦 Step 6: Installation${NC}"
echo -e "  ${CYAN}Plugin built successfully!${NC}"
echo ""
echo "  To install, run one of the following:"
echo "    • ./install.sh"
echo "    • sudo cmake --install build"
echo ""

echo -e "${GREEN}🎉 Build process complete!${NC}"
echo "========================================"
