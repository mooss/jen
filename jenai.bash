#!/usr/bin/env bash

{ # Bypass auto reload.

# Strict mode.
set -euo pipefail

# Automatically export all functions and variables.
# This is what allows prompt interpretation to work.
set -o allexport

####################
# Global variables #
####################

DEFAULT_MODEL=gemini

#############
# Arguments #

PROMPT=                # Prompt
POSARGS=()             # Positional arguments
CONTEXT_POS=below      # --context-above sets this to above
CONTEXT_DIRS=()        # --dir
RUNNER=ask-llm         # --dry-run|-n sets this to maybe-tee
CONTEXT_FILES=()       # --file
INTERACTIVE=false      # --interactive|-i
MODEL=""               # --model|-m
ONESHOT=false          # --oneshot|-o
EVALUATOR=eval-prompt  # --oneshot|-o and also --paste sets this to cat
PASTE=false            # --paste
TEE_FILE=""            # --tee
VARIABLES=()           # --var|-v

################
# Associations #

declare -Ar MODELS=(
  [codestral]=openrouter:mistralai/codestral-2501
  [gemini]=openrouter:google/gemini-2.0-flash-001
  [gemini-pro]=openrouter:google/gemini-2.0-pro-exp-02-05:free
  [haiku]=openrouter:anthropic/claude-3-haiku
  [r1]=openrouter:deepseek/deepseek-r1
  [sonnet]=openrouter:anthropic/claude-3.7-sonnet
  [r1-70]=openrouter:deepseek/deepseek-r1-distill-llama-70b:free
  [qwco]=openrouter:qwen/qwen-2.5-coder-32b-instruct
  [ds-v3]=openrouter:deepseek/deepseek-chat
)

declare -A PROMPTS=(
  [review-diff]='$(perins-jaded-review)

# Diffs

$(git diff HEAD^ HEAD)'
  [review-staged]='$(perins-jaded-review)

# Diffs

$(git diff --staged)'
  [commit-message]='$(per-jaded-dev)

$(ins-commit-msg)

# Diffs

$(git diff --staged)'
  [test]='Count the files:
$(ls)'
)

#############
# Functions #
#############

###########
# Prompts #
###########

function per-jaded-dev() {
  echo "# Persona

You are a senior developer.
You have other things to do, your time is precious, so is mine and you'd rather work on something interesting right now.
You are extremely factual and to the point."
}

function ins-code-review() {
  echo "# Instructions

You are reviewing code, be on the lookout for:
 - Logic errors and potential bugs.
 - Performance bottlenecks and inefficient algorithms.
 - Security vulnerabilities (e.g., injection flaws, insecure data handling).
 - Code clarity and maintainability (adherence to style guide, complex logic).
 - Missing error handling or edge case coverage.
 - Adherence to best practices (for programming language or framework).
 - Out-of-place print statement.
 - Any indication that something is not finished (like \`TODO:\` or \`WIP\` comments).

Do not explain the changes that have been made.
I know what they are, I made them!

Don't waste time on pointless details, be direct, we'll discuss later if we need it.
Do not mention good things about the code, I don't care, the only good thing is silence."
}

function ins-commit-msg() {
  echo '# Instructions

Analyse a diff and write a concise, informative and well-formatted commit message.
The commit message should clearly and accurately summarize the changes in the diff.

## Format

```
Short summary of what changed and why it changed

(optional) Longer description of the changes, not everything should be
explained, especially not the obvious things.
```

## Guidelines

For the short summary:
- Use the imperative, present tense: "change", not "changed" nor "changes".
- Capitalize the first letter.
- No dot (.) at the end.
- Keep it short (ideally 50 characters or less, definitely under 72).
- Focus on *what* was changed and *why*.

For the optional long description:
- Use the imperative, present tense.
- Wrap lines at 72 characters.
- Explain the *what* and *why* of the change, *not the how*. The code itself explains the *how*.
- Include motivation for the change and how it addresses the issue.
- Can be ignored if the change is very simple or already well-explained in the short summary'
}

function perins-jaded-review() {
  per-jaded-dev; echo; echo; ins-code-review
}

##############
# Primitives #

# Combines arguments and stdin.
function ecat() {
  maybe-cat
  [[ $# -gt 0 ]] && echo "$@"
  true
}

function err() {
  ecat "$@" >&2
}

function die() {
  err "$@"
  exit 23
}

function fileinfo() {
  du -sh "$1"
  wc -l "$1"
}

function maybe-tee() {
  if [[ -n $TEE_FILE ]]; then
    tee "$TEE_FILE"
  else
    cat
  fi
}

function maybe-cat() {
  if ! [[ -t 0 ]]; then
    cat
  fi
}

#################
# Git utilities #

##################
# Help functions #

function usage() {
  cat << EOF
Usage: $0 PROMPT [ARGS...] [OPTIONS]

Dispatch prompts to an LLM.

Options:
  --context-above     Put the context files and dir above the instructions
  --dir DIR           Include all files in directory as context
  --dry-run|-n        Print interpolated prompt without sending to LLM
  --file FILE         Include specific file(s) as context
  --help|-h           Display this help message
  --interactive|-i    Start an interactive aichat session
  --list|-l           List all available prompts
  --model|-m MODEL    Model name (default is $DEFAULT_MODEL, prompts may have their own default)
  --oneshot|-o        Use positional arguments as the prompt (which is not evaluated)
  --paste             Use the content of the clipboard as the prompt (which is not evaluated)
  --tee FILE          Output to both stdout and FILE (overwrites)
  --var|-v KEY=VALUE  Set variables for prompt interpolation

Available models:
$(for model in "${!MODELS[@]}"; do echo " - $model"; done)
EOF
}

######################
# Prompt evapolation #

# Concatenate a file, prefixed with its filename.
function context-cat(){
  local -r file="$1"
  if [[ -r "$file" ]]; then
    echo -e "\n==> $file <=="
    cat "$file"
    echo
  else
    die "Cannot read file: $file"
  fi
}

# Print context files and directories.
function context() {
  local -r out=$(__context)
  [[ -n $out ]] && echo -e "# Additional context (files)\n$out"
}

function __context() {
  # Files.
  for file in "${CONTEXT_FILES[@]}"; do
    context-cat "$file"
  done

  # Directories.
  for dir in "${CONTEXT_DIRS[@]}"; do
    if [[ -d "$dir" ]]; then
      find "$dir" -type f -print0 | while IFS= read -r -d $'\0' file; do
        context-cat "$file"
      done
    else
      die "Directory not found: $dir"
    fi
  done
}

# Resolve prompt source.
function prompt() {
  if [[ $PASTE == true ]]; then
    xclip -selection clipboard -o
    echo
  elif [[ $ONESHOT == true ]]; then
    echo "${POSARGS[@]}"
  else
    echo "${PROMPTS[$PROMPT]}"
  fi
}

# Resolve full model name.
function model() {
  echo "${MODELS[$MODEL]}"
}

function unique-file-prefix() {
  local -r directory="$1"
  local -r prefix="$2"
  local -r suffix="$3"

  local unique="$prefix"
  local counter=1
  while [[ -e $directory/$unique$suffix ]]; do
    unique="$prefix.$counter"
    ((counter++))
  done

  echo "$unique"
}

# Evaluate stdin as a prompt.
function eval-prompt() {
  local -r prompt=$(cat)

  # Generate a safe delimiter for the nested cat trick.
  local -r delim=$(openssl rand -hex 16)$(echo "$prompt" | sha512sum - | sed -r 's/^([0-9a-f]+).*/\1/')
  
  # The allexport option allows this bash interpreter to have access to everything we defined and
  # evapolate the prompts.
  cat <<EOF | env bash
cat <<$delim
$prompt
$delim
EOF
}

function generate-prompt() {
  if [[ $CONTEXT_POS == above ]]; then
    context; echo; echo
  fi

  prompt | $EVALUATOR

  if [[ $CONTEXT_POS == below ]]; then
    echo; echo; context
  fi
}

###########
# Runners #

function main() {
  # Runner is dry-run or ask-llm.
  generate-prompt | $RUNNER
}

# Sends stdin to the model.
function ask-llm() {
  local -r session_dir=".jenai/session"
  local -r session=$(unique-file-prefix $session_dir $(date +"%Y-%m-%d_%Hh%Mm") .yaml)
  local -r session_yaml="$session_dir/$session.yaml"

  mkdir -p "$session_dir"

  # Non-interactive session for the first question.
  AICHAT_SESSIONS_DIR="$session_dir" aichat --model "$(model)" --save-session --session $session

  # Cannot tee immediately or it disabled syntactic coloration.
  if [[ -n $TEE_FILE ]]; then
    yq -r '.messages[-1].content' "$session_yaml" > "$TEE_FILE"
  fi

  # If asked, the session is resumed interactively.
  if [[ $INTERACTIVE == true ]]; then
    # stdin is "spent", so this call should be interactive.
    AICHAT_SESSIONS_DIR="$session_dir" aichat --model "$(model)" --save-session --session $session
  fi
  
  err ''
  err 'End of session.'
  fileinfo "$session_yaml" | err
}

#################
# Script proper #
#################

####################
# Argument parsing #

# Process arguments.
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --context-above)
      shift
      CONTEXT_POS=above
      ;;

    --dir)
      shift
      while [[ $# -gt 0 ]] && [[ ! "$1" =~ ^-.* ]]; do
        CONTEXT_DIRS+=("$1")
        shift
      done
      ;;

    --dry-run|-n)
      RUNNER=maybe-tee
      shift
      ;;

    --file)
      shift
      while [[ $# -gt 0 ]] && [[ ! "$1" =~ ^-.* ]]; do
        CONTEXT_FILES+=("$1")
        shift
      done
      ;;

    --help|-h)
      usage; exit 0
      ;;

    --interactive|-i)
      INTERACTIVE=true
      shift
      ;;

    --list|-l)
      echo "${!PROMPTS[@]}"; exit 0
      ;;

    --model|-m)
      [[ $# -gt 1 ]] || die 'Missing value after -m|--model'
      MODEL="$2"
      shift; shift
      ;;

    --oneshot|-o)
      EVALUATOR=cat
      ONESHOT=true
      shift
      ;;

    --paste)
      EVALUATOR=cat
      PASTE=true
      shift
      ;;

    --tee)
      [[ $# -gt 1 ]] || die 'Missing value after --tee'
      TEE_FILE="$2"
      shift; shift
      ;;

    --var|-v) # TODO: Export the variable to the evapolater to make them available in the prompts.
      shift
      while [[ $# -gt 0 ]] && [[ ! "$1" =~ ^-.* ]]; do
        VARIABLES+=("$1")
        shift
      done
      ;;

    *) # TODO: provide mechanism to make the posargs convenient to process inside prompts.
      [[ "$1" == -* ]] && die "Unrecognized option: $1"
      POSARGS+=("$1")
      shift
      ;;
  esac
done

if [[ $PASTE == true && $ONESHOT == true ]]; then
  die --paste and --oneshot are mutually exclusive
fi

# Positional arguments are required, except When paste is given.
if [[ $PASTE == false && "${#POSARGS[@]}" -eq 0 ]]; then
  usage | err ''
  die No positional arguments provided.
fi

# By default, the first argument is the prompt name.
if [[ $ONESHOT == false && $PASTE == false ]]; then
  PROMPT="${POSARGS[0]}"
  POSARGS=("${POSARGS[@]:1}")
fi

# Set defaults and verify values.
[[ -z $MODEL ]] && MODEL=$DEFAULT_MODEL
[[ -v "MODELS[$MODEL]" ]] || die "Unknown model $MODEL"
[[ -z $PROMPT ]] || [[ -v "PROMPTS[$PROMPT]" ]] || die "Unknown prompt $PROMPT"

main

exit 0 # Needed to prevent bash to keep reading when the file changed.
}
