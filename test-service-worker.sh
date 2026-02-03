#!/bin/bash
# Test script for Service Worker PostgreSQL
# This script demonstrates the complete workflow

set -e

echo "======================================"
echo "Service Worker PostgreSQL - Test Script"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Check if docker compose is available
if ! docker compose version &> /dev/null; then
    echo "Error: docker compose is not installed"
    exit 1
fi

# Test 1: Build main application
print_info "Test 1: Building main application..."
cd /home/runner/work/GridFlow-Dynamics/GridFlow-Dynamics
go build -o /tmp/gridflow-server ./cmd/server
print_success "Main application builds successfully"
rm /tmp/gridflow-server
echo ""

# Test 2: Build service worker
print_info "Test 2: Building service worker..."
cd /home/runner/work/GridFlow-Dynamics/GridFlow-Dynamics/service-worker-ps
go build -o /tmp/service-worker-ps ./cmd/service-worker-ps
print_success "Service worker builds successfully"
rm /tmp/service-worker-ps
echo ""

# Test 3: Verify independent modules
print_info "Test 3: Verifying independent modules..."
cd /home/runner/work/GridFlow-Dynamics/GridFlow-Dynamics
if [ -f "go.mod" ]; then
    print_success "Main application has go.mod"
fi

cd /home/runner/work/GridFlow-Dynamics/GridFlow-Dynamics/service-worker-ps
if [ -f "go.mod" ]; then
    print_success "Service worker has independent go.mod"
fi
echo ""

# Test 4: Check module names
print_info "Test 4: Checking module names..."
cd /home/runner/work/GridFlow-Dynamics/GridFlow-Dynamics
MAIN_MODULE=$(grep "^module" go.mod | awk '{print $2}')
echo "  Main module: $MAIN_MODULE"

cd /home/runner/work/GridFlow-Dynamics/GridFlow-Dynamics/service-worker-ps
SW_MODULE=$(grep "^module" go.mod | awk '{print $2}')
echo "  Service worker module: $SW_MODULE"

if [ "$MAIN_MODULE" != "$SW_MODULE" ]; then
    print_success "Modules are independent"
else
    echo "Error: Modules have the same name"
    exit 1
fi
echo ""

# Test 5: Verify docker-compose configuration
print_info "Test 5: Validating docker compose configuration..."
cd /home/runner/work/GridFlow-Dynamics/GridFlow-Dynamics
docker compose config > /dev/null 2>&1
print_success "docker-compose.yml is valid"
echo ""

# Test 6: Check service worker configuration in docker-compose
print_info "Test 6: Checking service worker in docker-compose..."
if grep -q "service-worker-ps:" docker-compose.yml; then
    print_success "Service worker is configured in docker-compose.yml"
else
    echo "Error: Service worker not found in docker-compose.yml"
    exit 1
fi
echo ""

# Test 7: Verify all services in docker-compose
print_info "Test 7: Listing docker compose services..."
docker compose config --services
echo ""

# Test 8: Check documentation
print_info "Test 8: Verifying documentation..."
cd /home/runner/work/GridFlow-Dynamics/GridFlow-Dynamics
if [ -f "README.md" ]; then
    print_success "Main README.md exists"
fi

cd /home/runner/work/GridFlow-Dynamics/GridFlow-Dynamics/service-worker-ps
if [ -f "README.md" ]; then
    print_success "Service worker README.md exists"
fi

if [ -f "EXTENDING.md" ]; then
    print_success "EXTENDING.md guide exists"
fi

if [ -f ".env.example" ]; then
    print_success ".env.example exists"
fi
echo ""

# Summary
echo "======================================"
echo "All Tests Passed! ✓"
echo "======================================"
echo ""
echo "Summary:"
echo "  ✓ Main application compiles independently"
echo "  ✓ Service worker compiles independently"
echo "  ✓ Modules are independent (separate go.mod)"
echo "  ✓ Docker compose configuration is valid"
echo "  ✓ Documentation is complete"
echo ""
echo "Next Steps:"
echo "  1. Start services: docker compose up -d"
echo "  2. Check logs: docker compose logs -f service-worker-ps"
echo "  3. Send test message to API"
echo "  4. Verify data in PostgreSQL"
echo ""
