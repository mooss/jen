#!/usr/bin/env bash

{ # Bypass auto reload.

# Strict mode.
set -euo pipefail

####################
# Global variables #
####################

declare -r DEFAULT_MODEL=gemini
# Will hold the full name of the model to execute, whether it is the default or specified one.
ACTUAL_MODEL=

# Find the git repo root if in a git repository, otherwise use current directory.
declare -r GIT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
if [[ -n $GIT_ROOT ]] then
  declare -r SESSION_DIR="${GIT_ROOT}/.jenai/session"
else
  declare -r SESSION_DIR=".jenai/session"
fi

#############
# Arguments #

PROMPT=                # Prompt
POSARGS=()             # Positional arguments
CONTEXT_POS=below      # --context-above sets this to above
CONTEXT_DIRS=()        # --dir
RUNNER=ask-llm         # --dry-run|-n sets this to maybe-tee
CONTEXT_FILES=()       # --file
INTERACTIVE=false      # --interactive|-i
SPECIFIED_MODEL=""     # --model|-m
ONESHOT=false          # --oneshot|-o
EVALUATOR=eval-prompt  # --oneshot|-o and also --paste sets this to cat
PASTE=false            # --paste
SESSION=""             # --session
TEE_FILE=""            # --tee

################
# Associations #

declare -Ar MODELS=(
  [codestral]=openrouter:mistralai/codestral-2501
  [gemini]=openrouter:google/gemini-2.0-flash-001
  [gemini-pro]=openrouter:google/gemini-2.0-pro-exp-02-05:free
  [haiku]=openrouter:anthropic/claude-3-haiku
  [r1]=openrouter:deepseek/deepseek-r1:nitro
  [sonnet]=openrouter:anthropic/claude-3.7-sonnet
  [r1-70]=openrouter:deepseek/deepseek-r1-distill-llama-70b:free
  [qwco]=openrouter:qwen/qwen-2.5-coder-32b-instruct
  [ds-v3]=openrouter:deepseek/deepseek-chat
)

declare -A PROMPTS=(
  [review-diff]='$(perins-jaded-review)

$(git-diff HEAD^ HEAD)'
  [review-staged]='$(perins-jaded-review)

$(git-diff --staged)'
  [commit-message]='$(per-jaded-dev)

$(ins-commit-msg)

$(git-diff --staged)'
  [project-graph]='$(project-graph)

${POSARGS[@]}'
  [test]='Count the files:
$(ls)'
)

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

You are reviewing code.

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

function project-graph() {
  echo "# Instructions

I will present an idea.
Ask me one question at a time about this idea so we can develop a simple, flexible plan than can later be expanded and adapted if needed.

Each question should build on my previous answers, and our end goal is to have:
 1. A set of steps.
 2. A simple identifier for each step (human readable using kebab-case).
 3. And a dependency graph between the identifiers.

Let’s do this iteratively and not go into the details, we want to create a flexible outline than can later be refined in a just-in-time manner.
Remember, only one question at a time.

# Idea"
}

function git-diff() {
  echo '# Diff

## How to read

Lines starting with \`-\` are present in the original file but removed in the new version.
Lines starting with \`+\` are added in the new version.
Lines starting with a space are unchanged context lines present in both versions.

## Diff included below
'
  git diff "$@"
}

#############
# Functions #
#############

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
  --session SESSION   Reuse or create a specific session name (/last for most recent session)
  --tee FILE          Output to both stdout and FILE (overwrites)

Available models:
$(for model in "${!MODELS[@]}"; do echo " - $model"; done)
EOF
}

########################
# Prompt interpolation #

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

# Get the most recent session file
function get-latest-session() {
  find "$SESSION_DIR" -name "*.yaml" -type f -print0 | xargs -0 ls -t 2>/dev/null | head -n1
}

# Evaluate stdin as a prompt.
function eval-prompt() {
  local -r prompt=$(cat)

  # Generate a safe delimiter for the nested cat trick.
  local -r delim=$(openssl rand -hex 16)$(echo "$prompt" | sha512sum - | sed -r 's/^([0-9a-f]+).*/\1/')
  
  # The allexport option allows this bash interpreter to have access to everything we defined and
  # interpolate the prompts.
  source <(cat <<EOF
cat <<$delim
$prompt
$delim
EOF
)
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

function __aichat() {
  local -r session="$1"; model="$2"; shift; shift
  aichat\
    --model "$model"\
    --session "$session" --save-session "$@"
}

###########
# Runners #

function main() {
  # Runner is dry-run or ask-llm.
  generate-prompt | $RUNNER
}

# Sends stdin to the model.
function ask-llm() {
  local session
  if [[ -n $SESSION ]]; then
    session="$SESSION"
  else
    session=$(unique-file-prefix $SESSION_DIR $(date +"%Y-%m-%d_%Hh%Mm") .yaml)
  fi
  local -r session_yaml="$SESSION_DIR/$session.yaml"
  export AICHAT_COMPRESS_THRESHOLD=10000
  export AICHAT_SESSIONS_DIR="$SESSION_DIR"

  mkdir -p "$SESSION_DIR"

  # Only override an existing session model if it was explicitly speficied.
  local model=$ACTUAL_MODEL
  if [[ -z $SPECIFIED_MODEL && -f $session_yaml ]]; then
    model=$(yq -r .model "$session_yaml")
  fi

  # If there's a prompt, execute it non-interactively.
  local -r prompt=$(maybe-cat)
  if [[ -n $prompt ]]; then
    echo "$prompt" | __aichat "$session" "$model"
  fi

  # Cannot tee immediately or it disables syntactic coloration.
  if [[ -n $TEE_FILE ]]; then
    yq -r '.messages[-1].content' "$session_yaml" > "$TEE_FILE"
  fi

  # If asked (or a specific session was asked without a prompt), the session is resumed interactively.
  if [[ $INTERACTIVE == true || ( -n $SESSION && -z $prompt ) ]]; then
    __aichat "$session" "$model"
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
      SPECIFIED_MODEL="$2"
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

    --session)
      [[ $# -gt 1 ]] || die 'Missing value after --session'
      SESSION="$2"
      shift; shift
      ;;

    --tee)
      [[ $# -gt 1 ]] || die 'Missing value after --tee'
      TEE_FILE="$2"
      shift; shift
      ;;

    *)
      [[ "$1" == -* ]] && die "Unrecognized option: $1"
      POSARGS+=("$1")
      shift
      ;;
  esac
done

if [[ $PASTE == true && $ONESHOT == true ]]; then
  die --paste and --oneshot are mutually exclusive
fi

# Positional arguments are required, except when --paste is given.
if [[ $PASTE == false && "${#POSARGS[@]}" -eq 0 ]]; then
  # They are also not required when an existing session has been specified.
  # The problem is that we don't know if it exists, but that will do for now.
  if [[ -n $SESSION ]]; then
    ONESHOT=true # Will feed an empty prompt.
  else
    usage | err ''
    die No positional arguments provided
  fi
fi

# By default, the first argument is the prompt name.
if [[ $ONESHOT == false && $PASTE == false ]]; then
  PROMPT="${POSARGS[0]}"
  POSARGS=("${POSARGS[@]:1}")
fi

# Set defaults and verify values.
ACTUAL_MODEL=$DEFAULT_MODEL
[[ -n $SPECIFIED_MODEL ]] && ACTUAL_MODEL=$SPECIFIED_MODEL
[[ -v "MODELS[$ACTUAL_MODEL]" ]] || die "Unknown model $ACTUAL_MODEL"
ACTUAL_MODEL="${MODELS[$ACTUAL_MODEL]}"
[[ -z $PROMPT ]] || [[ -v "PROMPTS[$PROMPT]" ]] || die "Unknown prompt $PROMPT"

# Handle /last special session value.
if [[ $SESSION == "/last" ]]; then
  [[ -d $SESSION_DIR ]] || die "Cannot find latest session, there is no session dir"
  SESSION=$(basename "$(get-latest-session)" .yaml)
  [[ -n $SESSION ]] || die "No previous session found"
  err "Last session is $SESSION."
fi

main

exit 0 # Needed to prevent bash to keep reading when the file changed.
}
