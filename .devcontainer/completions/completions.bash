# Shell completion for kubectl, kind and docker (bash).
# Sourced from ~/.bashrc by .devcontainer/post-create.sh.

# kubectl's completion needs bash-completion loaded first.
if ! type _init_completion &>/dev/null; then
  [ -r /usr/share/bash-completion/bash_completion ] && . /usr/share/bash-completion/bash_completion
fi
command -v kubectl &>/dev/null && source <(kubectl completion bash)
command -v kind    &>/dev/null && source <(kind completion bash)
command -v docker  &>/dev/null && source <(docker completion bash 2>/dev/null) 2>/dev/null || true
alias k=kubectl
command -v kubectl &>/dev/null && complete -o default -F __start_kubectl k
