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

spinner() {
    local pid=$!

    local delay=0.07
    local spinstr='⠇⠋⠙⠸⠴⠦'
    local temp_file=$(mktemp)

    local task_name=$1

    shift 1

    echo "$1" > "$temp_file"

    start_time=$(date +%s)
    while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
        local temp=$(cat "$temp_file")
        local elapsed=$(ps -p $pid -o etimes=)
        local clean_line=$(printf "\r\033[K")
        local i=$(($i+1))
        local spin=${spinstr:$i%${#spinstr}:1}
        echo -ne "${clean_line}$spin $task_name ${elapsed}s"
        sleep $delay
    done
    wait $pid
    end_time=$(date +%s)
    duration_seconds=$((end_time - start_time))
    local exit_code=$?

    local elapsed=$(ps -p $pid -o etimes=)

    if [ $exit_code -ne 0 ]; then
        echo -e "${clean_line}$RED_CROSS $task_name ${duration_seconds}s"
        exit $exit_code
    fi

    echo -e "${clean_line}$GREEN_CHECK $task_name ${duration_seconds}s"

    rm "$temp_file"
    return $exit_code
}

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

    # If NON_INTERACTIVE=true, then don't show spinner, instead run waiter function
    if [ "$NON_INTERACTIVE" = "true" ]; then
        "$@" &
        waiter "$task_name" "$@"
        return
    else
        # Run command and redirect stdout to /dev/null
        err_file=$(mktemp)
        "$@" > /dev/null 2> $err_file &
        spinner "$task_name" "$@" 

        # Check if any content in the error file
        if [ -s $err_file ]; then
            local RED_CROSS='\033[31;1m✘\033[0m'
            echo ""
            echo ""
            echo -e "$RED_CROSS Failed to run $task_name"
            echo ""
            echo -e "Error: $(cat $err_file)"
            exit 1
        fi
    fi
}
