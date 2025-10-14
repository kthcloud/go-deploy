#!/usr/bin/env bash

RED="\e[31m"
YELLOW="\e[33m"
GREEN="\e[32m"
BLUE="\e[34m"
GRAY="\e[90m"
RESET="\e[0m"

ERROR_LOG="${RED}[ERROR]${RESET}"
WARN_LOG="${YELLOW}[WARN]${RESET}"
INFO_LOG="${GREEN}[INFO]${RESET}"
PRINT_INDENT_LOG="${GRAY}||===>${RESET}"

log_err() {
    echo -e "$ERROR_LOG $1" >&2
}

log_warn() {
    echo -e "$WARN_LOG $1" >&2
}

log_info() {
    echo -e "$INFO_LOG $1" >&2
}

log_print() {
    echo -e "$PRINT_INDENT_LOG $1" >&2
}

get_git_root() {
    git rev-parse --show-toplevel 2>/dev/null || { echo -e "${RED}This is not a git repository. Please run this script from within the Git repository.${RESET}"; exit 1; }
}

check_gofmt() {
    command -v gofmt >/dev/null 2>&1 || { echo -e "${RED}gofmt is not installed. Please install it to continue.${RESET}"; exit 1; }
}

check_go_vet() {
    command -v go >/dev/null 2>&1 || { echo -e "${RED}go is not installed. Please install Go to continue.${RESET}"; exit 1; }
}

check_go_cyclo() {
    command -v gocyclo >/dev/null 2>&1 || { echo -e "${YELLOW}go-cyclo is not installed. Installing...${RESET}"; go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; }
}

check_ineffassign() {
    command -v ineffassign >/dev/null 2>&1 || { echo -e "${YELLOW}ineffassign is not installed. Installing...${RESET}"; go install github.com/gordonklaus/ineffassign@latest; }
}

check_staticcheck() {
    command -v staticcheck >/dev/null 2>&1 || { echo -e "${YELLOW}go-staticcheck is not installed. Installing...${RESET}"; go install honnef.co/go/tools/cmd/staticcheck@latest; }
}

run_gofmt() {
    log_info "${YELLOW}Running gofmt check...${RESET}"
    # Check for formatting issues
    gofmt_output=$(gofmt -l .)
    if [ -n "$gofmt_output" ]; then
        log_err "${RED}gofmt found the following issues:${RESET}"
        echo "$gofmt_output" | while IFS= read -r line; do
            log_print "\t$line"
        done
        gofmt_error=1
    else
        log_print "${GREEN}passed gofmt.${RESET}"
    fi
}

run_go_vet() {
    log_info "${YELLOW}Running go vet check...${RESET}"
    # Run go vet to analyze code for potential issues
    go_vet_output=$(go vet ./... 2>&1)
    if [ -n "$go_vet_output" ]; then
        log_err "${RED}go vet found the following issues:${RESET}"
        echo "$go_vet_output" | while IFS= read -r line; do
            log_print "\t$line"
        done
        go_vet_error=1
    else
        log_print "${GREEN}passed go vet.${RESET}"
    fi
}

run_go_cyclo() {
    log_info "${YELLOW}Running go-cyclo check...${RESET}"
    # Check for cyclomatic complexity
    go_cyclo_output=$(find . -path ./pkg/imp -prune -o -type f -name "*.go" -exec gocyclo -over 15 {} +)
    if [ -n "$go_cyclo_output" ]; then
        log_err "${RED}gocyclo found the following issues:${RESET}"
        echo "$go_cyclo_output" | while IFS= read -r line; do
            log_print "\t$line"
        done
        go_cyclo_error=1
    else
        log_print "${GREEN}passed gocyclo -over 15.${RESET}"
    fi
}

run_ineffassign() {
    log_info "${YELLOW}Running ineffassign check...${RESET}"
    # Check for unused variable assignments
    ineffassign_output=$(ineffassign .)
    if [ -n "$ineffassign_output" ]; then
        log_err "${RED}ineffassign found the following issues:${RESET}"
        echo "$ineffassign_output" | while IFS= read -r line; do
            log_print "\t$line"
        done
        ineffassign_error=1
    else
        log_print "${GREEN}passed ineffassign.${RESET}"
    fi
}

run_staticcheck() {
    log_info "${YELLOW}Running go-staticcheck check...${RESET}"
    # Run staticcheck to analyze code for potential issues
    staticcheck_output=$(staticcheck ./...)
    if [ -n "$staticcheck_output" ]; then
        log_err "${RED}go-staticcheck found the following issues:${RESET}"
        echo "$staticcheck_output" | while IFS= read -r line; do
            log_print "\t$line"
        done
        staticcheck_error=1
    else
        log_print "${GREEN}passed go-staticcheck.${RESET}"
    fi
}

pushd "$(dirname "${BASH_SOURCE[0]}")/.." > /dev/null || { log_err "Failed to change to script directory."; exit 1; }

# Ensure all tools are installed
check_gofmt
check_go_vet
check_go_cyclo
check_ineffassign
check_staticcheck

go mod download

# Initialize error flags
gofmt_error=0
go_vet_error=0
go_cyclo_error=0
ineffassign_error=0
staticcheck_error=0

# Execute checks
run_gofmt
run_go_vet
run_go_cyclo
run_ineffassign
run_staticcheck

# Print a summary and exit with failure if there were any errors
if [ $gofmt_error -eq 1 ] || [ $go_vet_error -eq 1 ] || [ $go_cyclo_error -eq 1 ] || [ $ineffassign_error -eq 1 ] || [ $staticcheck_error -eq 1 ]; then
    echo -e "${RED}SUMMARY: Please resolve the above issues.${RESET}"
    popd > /dev/null
    exit 1
else
    echo -e "${GREEN}All checks passed successfully.${RESET}"
    popd > /dev/null
    exit 0
fi
