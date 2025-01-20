#!/bin/sh

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

function get_git_root() {
    git rev-parse --show-toplevel 2>/dev/null || { echo -e "${RED}This is not a git repository. Please run this script from within the Git repository.${NC}"; exit 1; }
}

function check_gofmt() {
    command -v gofmt >/dev/null 2>&1 || { echo -e "${RED}gofmt is not installed. Please install it to continue.${NC}"; exit 1; }
}

function check_go_vet() {
    command -v go >/dev/null 2>&1 || { echo -e "${RED}go is not installed. Please install Go to continue.${NC}"; exit 1; }
}

function check_go_cyclo() {
    command -v gocyclo >/dev/null 2>&1 || { echo -e "${YELLOW}go-cyclo is not installed. Installing...${NC}"; go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; }
}

function check_ineffassign() {
    command -v ineffassign >/dev/null 2>&1 || { echo -e "${YELLOW}ineffassign is not installed. Installing...${NC}"; go install github.com/gordonklaus/ineffassign@latest; }
}

function run_gofmt() {
    echo -e "${YELLOW}Running gofmt check...${NC}"
    # Check for formatting issues
    gofmt_output=$(gofmt -l .)
    if [ -n "$gofmt_output" ]; then
        echo -e "${RED}gofmt found the following issues:${NC}"
        echo "$gofmt_output"
        gofmt_error=1
    else
        echo -e "${GREEN}passed gofmt.${NC}"
    fi
}

function run_go_vet() {
    echo -e "${YELLOW}Running go vet check...${NC}"
    # Run go vet to analyze code for potential issues
    go_vet_output=$(go vet ./... 2>&1)
    if [ -n "$go_vet_output" ]; then
        echo -e "${RED}go vet found the following issues:${NC}"
        echo "$go_vet_output"
        go_vet_error=1
    else
        echo -e "${GREEN}passed go vet.${NC}"
    fi
}

function run_go_cyclo() {
    echo -e "${YELLOW}Running go-cyclo check...${NC}"
    # Check for cyclomatic complexity
    go_cyclo_output=$(gocyclo -over 15 .)
    if [ -n "$go_cyclo_output" ]; then
        echo -e "${RED}gocyclo found the following issues:${NC}"
        echo "$go_cyclo_output"
        go_cyclo_error=1
    else
        echo -e "${GREEN}passed gocyclo -over 15.${NC}"
    fi
}

function run_ineffassign() {
    echo -e "${YELLOW}Running ineffassign check...${NC}"
    # Check for unused variable assignments
    ineffassign_output=$(ineffassign .)
    if [ -n "$ineffassign_output" ]; then
        echo -e "${RED}ineffassign found the following issues:${NC}"
        echo "$ineffassign_output"
        ineffassign_error=1
    else
        echo -e "${GREEN}passed ineffassign.${NC}"
    fi
}

# Get the git root, to make sure the scripts are run in the correct location
git_root=$(get_git_root)

# Change to the Git root directory
cd "$git_root" || { echo -e "${RED}Failed to change to the Git root directory.${NC}"; exit 1; }

# Ensure all tools are installed
check_gofmt
check_go_vet
check_go_cyclo
check_ineffassign

# Initialize error flags
gofmt_error=0
go_vet_error=0
go_cyclo_error=0
ineffassign_error=0

# Execute checks
run_gofmt
run_go_vet
run_go_cyclo
run_ineffassign

# Print a summary and exit with failure if there were any errors
if [ $gofmt_error -eq 1 ] || [ $go_vet_error -eq 1 ] || [ $go_cyclo_error -eq 1 ] || [ $ineffassign_error -eq 1 ]; then
    echo -e "${RED}SUMMARY: Please resolve the above issues.${NC}"
    exit 1
else
    echo -e "${GREEN}All checks passed successfully.${NC}"
    exit 0
fi
