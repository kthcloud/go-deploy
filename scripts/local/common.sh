#!/bin/bash

spinner() {
    local pid=$!
    local delay=0.07
    local spinstr='⠇⠋⠙⠸⠴⠦'
    local green_check='\033[32;1m✔\033[0m'
    local red_cross='\033[31;1m✘\033[0m'
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
        echo -e "${clean_line}$red_cross $task_name ${duration_seconds}s"
        exit $exit_code
    fi

    echo -e "${clean_line}$green_check $task_name ${duration_seconds}s"

    rm "$temp_file"
    return $exit_code
}

run_with_spinner() {
    local task_name=$1
    shift  # Remove the first argument which is the task name
    # Run command and redirect stdout to /dev/null
    err_file=$(mktemp)
    "$@" > /dev/null 2> $err_file &
    spinner "$task_name" "$@"

    # Check if any content in the error file
    if [ -s $err_file ]; then
        local red_cross='\033[31;1m✘\033[0m'
        echo ""
        echo ""
        echo -e "$red_cross Failed to run $task_name"
        echo ""
        echo -e "Error: $(cat $err_file)"
        exit 1
    fi
}
