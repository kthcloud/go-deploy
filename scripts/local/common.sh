#!/bin/bash

RESET="\033[0m"

GREEN_BOLD="\033[32;1m"
RED_BOLD="\033[31;1m"
TEAL_BOLD="\033[36;1m"
BLUE_BOLD="\033[34;1m"
ORANGE_BOLD="\033[33;1m"
WHITE_BOLD="\033[37;1m"

PINK_BOLD="\033[95;1m"

RED_CROSS="${RED_BOLD}✘${RESET}"
GREEN_CHECK="${GREEN_BOLD}✔${RESET}"
BLUE_RIGHT_ARROW="${BLUE_BOLD}➡${RESET}"

waiter() {
    local pid=$!

    local task_name=$1
    echo -e "${BLUE_RIGHT_ARROW}  ${task_name}"

    wait $pid
    local exit_code=$?

    if [ $exit_code -ne 0 ]; then
        echo -e "${RED_CROSS}  ${task_name}"
        exit $exit_code
    fi

    echo -e "${GREEN_CHECK}  ${task_name}"
    return $exit_code
}

run_task() {
    local task_name=$1
    shift  # Remove the first argument which is the task name

    "$@" &
    waiter "$task_name" "$@"
}
